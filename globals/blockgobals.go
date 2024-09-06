package globals

import (
	structs "ArchivosP1/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ReadFileBlocks lee todos los bloques asignados a un archivo (como users.txt) y devuelve su contenido completo
func ReadFileBlocks(file *os.File, sb *structs.Superblock, inode *structs.Inode) (string, error) {
	var contenido string

	// Recorre todos los bloques asignados al archivo
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		blockOffset := int64(sb.S_block_start + blockIndex*int32(structs.BlockSize))
		var fileBlock structs.FileBlock

		err := fileBlock.Decode(file, blockOffset)
		if err != nil {
			return "", fmt.Errorf("error leyendo bloque %d: %w", blockIndex, err)
		}

		contenido += string(fileBlock.B_content[:])
	}

	return strings.TrimRight(contenido, "\x00"), nil // Remover caracteres nulos al final
}

// WriteFileBlocks escribe contenido en los bloques asignados a un archivo (como users.txt)
func WriteFileBlocks(file *os.File, sb *structs.Superblock, inode *structs.Inode, contenido string) error {
	bloques := dividirEnBloques([]byte(contenido), structs.BlockSize)

	for i, bloque := range bloques {
		// Si el bloque no está asignado, obtener uno nuevo
		if inode.I_block[i] == -1 {
			blockIndex, err := sb.GetNextFreeBlock(file)
			if err != nil {
				return fmt.Errorf("error obteniendo siguiente bloque libre: %w", err)
			}
			inode.I_block[i] = blockIndex
		}

		blockOffset := int64(sb.S_block_start + inode.I_block[i]*int32(binary.Size(structs.FileBlock{})))
		var fileBlock structs.FileBlock

		// Leer el bloque existente para evitar reescribir si no es necesario
		err := fileBlock.Decode(file, blockOffset)
		if err != nil {
			return fmt.Errorf("error leyendo bloque existente: %w", err)
		}

		if string(fileBlock.B_content[:]) != string(bloque) {
			// Solo reescribir el bloque si ha cambiado
			copy(fileBlock.B_content[:], bloque)
			err = fileBlock.Encode(file, blockOffset)
			if err != nil {
				return fmt.Errorf("error escribiendo bloque: %w", err)
			}
		}
	}

	// Actualizar el tamaño del inodo con el nuevo contenido
	inode.I_size = int32(len(contenido))

	return nil
}

// Función auxiliar para dividir contenido en bloques
func dividirEnBloques(data []byte, blockSize int) [][]byte {
	var bloques [][]byte
	for i := 0; i < len(data); i += blockSize {
		end := i + blockSize
		if end > len(data) {
			end = len(data)
		}
		bloques = append(bloques, data[i:end])
	}
	return bloques
}

// AssignNewBlock asigna un nuevo bloque a un inodo dado si no hay espacio disponible en los bloques actuales.
func AssignNewBlock(file *os.File, sb *structs.Superblock, inode *structs.Inode) error {
	// Iterar sobre los bloques del inodo para encontrar uno vacío (-1)
	for i := 0; i < len(inode.I_block); i++ {
		if inode.I_block[i] == -1 {
			// Obtener un nuevo bloque del superbloque
			newBlock, err := sb.GetNextFreeBlock(file)
			if err != nil {
				return fmt.Errorf("error obteniendo un nuevo bloque: %w", err)
			}
			inode.I_block[i] = newBlock

			// Actualizar el bitmap de bloques para reflejar que se ha usado un nuevo bloque
			err = structs.UpdateBlockBitmap(file, sb, newBlock)
			if err != nil {
				return fmt.Errorf("error actualizando el bitmap de bloques: %w", err)
			}

			// Devolver el éxito de la asignación
			return nil
		}
	}
	return fmt.Errorf("no hay bloques disponibles en el inodo para asignar un nuevo bloque")
}

// FindInUsersFile busca un grupo o usuario en el archivo users.txt
func FindInUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, name, entityType string) (string, error) {
	contenido, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return "", err
	}

	lineas := strings.Split(contenido, "\n")
	for _, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) < 3 {
			continue // Ignorar líneas mal formadas
		}

		// Buscar coincidencias con el tipo y nombre
		tipo, nombre := campos[1], campos[2]
		if tipo == entityType && nombre == name {
			return linea, nil
		}
	}

	return "", fmt.Errorf("%s '%s' no encontrado en users.txt", entityType, name)
}

func AddToUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, entry string) error {
	contenido, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return err
	}

	// Agregar la nueva entrada (grupo o usuario)
	contenido += entry + "\n"

	bloques := dividirEnBloques([]byte(contenido), structs.BlockSize)

	for i := 0; i < len(bloques); i++ {
		if inode.I_block[i] == -1 {
			// Asignar un nuevo bloque solo si no está asignado
			newBlock, err := sb.GetNextFreeBlock(file)
			if err != nil {
				return fmt.Errorf("error obteniendo un nuevo bloque: %w", err)
			}
			inode.I_block[i] = newBlock
		}

		blockOffset := int64(sb.S_block_start + inode.I_block[i]*int32(binary.Size(structs.FileBlock{})))
		var fileBlock structs.FileBlock

		// Leer el bloque existente
		err := fileBlock.Decode(file, blockOffset)
		if err != nil {
			return fmt.Errorf("error leyendo bloque existente: %w", err)
		}

		// Evitar reescribir bloques si no es necesario
		if string(fileBlock.B_content[:]) != string(bloques[i]) {
			copy(fileBlock.B_content[:], bloques[i])
			err := fileBlock.Encode(file, blockOffset)
			if err != nil {
				return fmt.Errorf("error escribiendo bloque: %w", err)
			}
		}
	}

	// Actualizar el tamaño del inodo
	inode.I_size = int32(len(contenido))

	// Actualizar el inodo en el archivo
	err = inode.Encode(file, int64(sb.S_inode_start+int32(binary.Size(*inode))))
	if err != nil {
		return fmt.Errorf("error actualizando inodo: %w", err)
	}

	// Actualizar el Superblock
	err = sb.Encode(file, int64(sb.S_bm_block_start))
	if err != nil {
		return fmt.Errorf("error actualizando Superblock: %w", err)
	}

	return nil
}

// GetNextID obtiene el siguiente ID disponible para un grupo o usuario en users.txt
func GetNextID(file *os.File, sb *structs.Superblock, inode *structs.Inode) (int, error) {
	contenido, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return 0, err
	}

	// Encontrar el ID más alto en el archivo
	lineas := strings.Split(contenido, "\n")
	maxID := 0
	for _, linea := range lineas {
		if linea == "" {
			continue
		}
		datos := strings.Split(linea, ",")
		if len(datos) == 0 {
			continue
		}
		id, err := strconv.Atoi(datos[0])
		if err == nil && id > maxID {
			maxID = id
		}
	}

	return maxID + 1, nil
}

// RemoveFromUsersFile elimina un grupo o usuario del archivo users.txt (cambio lógico, marcando como eliminado)
func RemoveFromUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, name, entityType string) error {
	// Leer el contenido del archivo users.txt
	contenido, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return err
	}

	// Buscar la línea correspondiente al grupo o usuario y marcarla como eliminada (ID 0)
	lineas := strings.Split(contenido, "\n")
	for i, linea := range lineas {
		if strings.Contains(linea, fmt.Sprintf(",%s,%s", entityType, name)) {
			datos := strings.Split(linea, ",")
			if len(datos) > 0 && datos[0] != "0" {
				// Marcar como eliminado cambiando el ID a 0
				datos[0] = "0"
				lineas[i] = strings.Join(datos, ",")
				break
			}
		}
	}

	// Actualizar el contenido de users.txt
	nuevoContenido := strings.Join(lineas, "\n")
	err = WriteFileBlocks(file, sb, inode, nuevoContenido)
	if err != nil {
		return fmt.Errorf("error actualizando users.txt para eliminar %s '%s': %v", entityType, name, err)
	}

	return nil
}

func assignNewBlock(file *os.File, sb *structs.Superblock, inode *structs.Inode) (int32, error) {
	// Recorrer los bloques directos
	for i := 0; i < 12; i++ {
		if inode.I_block[i] == -1 {
			// Asignar un nuevo bloque directo
			blockIndex, err := sb.GetNextFreeBlock(file)
			if err != nil {
				return -1, fmt.Errorf("error asignando bloque directo: %w", err)
			}
			inode.I_block[i] = blockIndex
			return blockIndex, nil
		}
	}

	// Manejar bloques indirectos cuando los bloques directos están llenos
	return assignIndirectBlock(file, sb, inode)
}

// Manejo de bloques indirectos
func assignIndirectBlock(file *os.File, sb *structs.Superblock, inode *structs.Inode) (int32, error) {
	// Asignar un bloque indirecto simple
	if inode.I_block[12] == -1 {
		blockIndex, err := sb.GetNextFreeBlock(file)
		if err != nil {
			return -1, fmt.Errorf("error asignando bloque indirecto: %w", err)
		}
		inode.I_block[12] = blockIndex
	}

	// Leer y actualizar el bloque indirecto simple
	var indirectBlock structs.PointerBlock
	blockOffset := int64(sb.S_block_start + inode.I_block[12]*int32(structs.BlockSize))
	err := indirectBlock.Decode(file, blockOffset)
	if err != nil {
		return -1, fmt.Errorf("error leyendo bloque indirecto: %w", err)
	}

	for i := 0; i < len(indirectBlock.B_pointers); i++ {
		if indirectBlock.B_pointers[i] == -1 {
			blockIndex, err := sb.GetNextFreeBlock(file)
			if err != nil {
				return -1, fmt.Errorf("error asignando nuevo bloque: %w", err)
			}
			indirectBlock.B_pointers[i] = int64(blockIndex)
			err = indirectBlock.Encode(file, blockOffset)
			if err != nil {
				return -1, fmt.Errorf("error escribiendo bloque indirecto: %w", err)
			}
			return blockIndex, nil
		}
	}

	return -1, fmt.Errorf("no hay bloques disponibles en el bloque indirecto simple")
}

// UpdateLineInUsersFile actualiza una línea en el archivo users.txt, buscando por nombre de usuario o grupo
func UpdateLineInUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, newLine, name, entityType string) error {
	contenido, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return err
	}

	// Buscar la línea correspondiente al usuario/grupo y reemplazarla con la nueva línea
	lineas := strings.Split(contenido, "\n")
	for i, linea := range lineas {
		if strings.Contains(linea, fmt.Sprintf(",%s,%s", entityType, name)) {
			lineas[i] = newLine
			break
		}
	}

	// Volver a escribir el archivo users.txt
	nuevoContenido := strings.Join(lineas, "\n")
	err = WriteFileBlocks(file, sb, inode, nuevoContenido)
	if err != nil {
		return fmt.Errorf("error actualizando %s '%s' en users.txt: %w", entityType, name, err)
	}

	return nil
}
