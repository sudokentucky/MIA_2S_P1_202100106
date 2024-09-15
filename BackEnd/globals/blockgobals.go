package globals

import (
	structs "backend/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// ReadFileBlocks lee todos los bloques asignados a un archivo (como users.txt) y devuelve su contenido completo
func ReadFileBlocks(file *os.File, sb *structs.Superblock, inode *structs.Inode) (string, error) {
	var contenido string

	// Iterar sobre los bloques asignados al inodo
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break // Detenemos cuando encontramos un bloque no asignado
		}

		// Calculamos el desplazamiento del bloque en el archivo usando el inicio del bloque
		blockOffset := int64(sb.S_block_start + blockIndex*64) // 64 es el tamaño fijo de FileBlock
		var fileBlock structs.FileBlock

		// Usamos la función Decode para deserializar el bloque desde el archivo
		err := fileBlock.Decode(file, blockOffset)
		if err != nil {
			return "", fmt.Errorf("error leyendo bloque %d: %w", blockIndex, err)
		}

		// Concatenamos el contenido del bloque al resultado total
		contenido += string(fileBlock.B_content[:]) // Convertimos el array a string
	}

	// Eliminamos caracteres nulos (\x00) que puedan aparecer al final del contenido
	return strings.TrimRight(contenido, "\x00"), nil
}

func WriteFileBlocks(file *os.File, sb *structs.Superblock, inode *structs.Inode, contenido string) error {
	bloques := dividirEnBloques([]byte(contenido), 64) // Dividimos el contenido en bloques de 64 bytes

	for _, bloque := range bloques {
		// Usar la función unificada para asignar o encontrar un bloque directo
		blockIndex, err := assignNewBlock(file, sb, inode)
		if err != nil {
			return err // Devuelve el error si no se pueden asignar más bloques
		}

		blockOffset := int64(sb.S_block_start + blockIndex*64)

		// Leer el bloque existente para evitar reescribir si no es necesario
		var fileBlock structs.FileBlock
		err = fileBlock.Decode(file, blockOffset)
		if err != nil {
			return fmt.Errorf("error leyendo bloque existente: %w", err)
		}

		// Si hay espacio en el bloque actual, escribimos el contenido
		usedSpace := len(strings.TrimRight(string(fileBlock.B_content[:]), "\x00"))
		if usedSpace < 64 {
			copy(fileBlock.B_content[usedSpace:], bloque)
			err = fileBlock.Encode(file, blockOffset)
			if err != nil {
				return fmt.Errorf("error escribiendo bloque: %w", err)
			}
		}
	}

	// Actualizar el tamaño del inodo con el nuevo contenido
	inode.I_size = int32(len(contenido))

	// Guardar el inodo actualizado en el archivo
	err := inode.Encode(file, int64(sb.S_inode_start+int32(binary.Size(*inode))))
	if err != nil {
		return fmt.Errorf("error actualizando inodo: %w", err)
	}

	// Actualizar el bitmap de inodos después de escribir el archivo
	err = sb.UpdateBitmapInode(file)
	if err != nil {
		return fmt.Errorf("error actualizando bitmap de inodos: %w", err)
	}

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AssignNewBlock asigna un nuevo bloque a un inodo dado si no hay espacio disponible en los bloques actuales.
func AssignNewBlock(file *os.File, sb *structs.Superblock, inode *structs.Inode) (int32, error) {
	// Iterar sobre los bloques directos del inodo
	for i := 0; i < len(inode.I_block); i++ {
		if inode.I_block[i] == -1 {
			// No hay bloque asignado, asignar uno nuevo
			newBlock, err := sb.FindNextFreeBlock(file)
			if err != nil {
				return -1, fmt.Errorf("error obteniendo un nuevo bloque: %w", err)
			}
			inode.I_block[i] = newBlock

			// Actualizar el bitmap de bloques sin pasar el índice del bloque
			err = sb.UpdateBitmapBlock(file)
			if err != nil {
				return -1, fmt.Errorf("error actualizando el bitmap de bloques: %w", err)
			}

			// Devolver el índice del nuevo bloque asignado
			return newBlock, nil
		}

		// Si el bloque ya está asignado, verificar si hay espacio disponible
		blockOffset := int64(sb.S_block_start + inode.I_block[i]*64)
		var fileBlock structs.FileBlock

		// Leer el bloque existente
		err := fileBlock.Decode(file, blockOffset)
		if err != nil {
			return -1, fmt.Errorf("error leyendo bloque existente: %w", err)
		}

		// Verificar si el bloque tiene espacio disponible
		usedSpace := len(strings.TrimRight(string(fileBlock.B_content[:]), "\x00"))
		if usedSpace < 64 {
			// Bloque tiene espacio, no es necesario asignar un nuevo bloque
			return inode.I_block[i], nil
		}
	}

	// Si todos los bloques están asignados y llenos, devolver error
	return -1, fmt.Errorf("no hay bloques disponibles en el inodo para asignar un nuevo bloque")
}

func AddOrUpdateInUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, entry, name, entityType string) error {
	contenido, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return fmt.Errorf("error leyendo blocks de users.txt: %w", err)
	}

	// Buscar si el usuario/grupo ya existe usando la nueva función auxiliar
	_, index, err := findLineInUsersFile(contenido, name, entityType)
	if err == nil {
		// Si lo encontramos, actualizamos la línea
		lineas := strings.Split(contenido, "\n")
		lineas[index] = entry // Actualizamos la línea
		contenido = strings.Join(lineas, "\n")
	} else {
		// Si no se encuentra, agregamos la nueva entrada
		contenido += entry + "\n"
	}

	// Escribir el contenido actualizado en los bloques del archivo
	err = WriteFileBlocks(file, sb, inode, contenido)
	if err != nil {
		return fmt.Errorf("error escribiendo en users.txt: %w", err)
	}

	return nil
}

func FindInUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, name, entityType string) (string, error) {
	contenido, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return "", err
	}

	// Usamos la nueva función auxiliar para buscar la línea
	linea, _, err := findLineInUsersFile(contenido, name, entityType)
	if err != nil {
		return "", err
	}

	return linea, nil
}

func RemoveFromUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, name, entityType string) error {
	contenido, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return err
	}

	// Buscar la línea que queremos eliminar usando la nueva función auxiliar
	linea, index, err := findLineInUsersFile(contenido, name, entityType)
	if err != nil {
		return err
	}

	// Marcar la entrada como eliminada (ID = 0)
	lineas := strings.Split(contenido, "\n")
	campos := strings.Split(linea, ",")
	if len(campos) > 0 && campos[0] != "0" {
		campos[0] = "0" // Marcar como eliminada
		lineas[index] = strings.Join(campos, ",")
	}

	// Actualizar el contenido del archivo users.txt
	nuevoContenido := strings.Join(lineas, "\n")
	err = WriteFileBlocks(file, sb, inode, nuevoContenido)
	if err != nil {
		return fmt.Errorf("error escribiendo en users.txt para eliminar %s '%s': %v", entityType, name, err)
	}

	return nil
}

// assignNewBlock asigna un nuevo bloque a un inodo dado si no hay espacio disponible en los bloques directos.
func assignNewBlock(file *os.File, sb *structs.Superblock, inode *structs.Inode) (int32, error) {
	// Recorrer los bloques directos (máximo 12 bloques)
	for i := 0; i < 12; i++ {
		if inode.I_block[i] == -1 {
			// Asignar un nuevo bloque directo
			blockIndex, err := sb.FindNextFreeBlock(file)
			if err != nil {
				return -1, fmt.Errorf("error asignando bloque directo: %w", err)
			}
			inode.I_block[i] = blockIndex
			return blockIndex, nil
		}
	}

	return -1, fmt.Errorf("todos los bloques directos están llenos")
}

// findLineInUsersFile busca una línea en el archivo users.txt según nombre y tipo
func findLineInUsersFile(contenido string, name, entityType string) (string, int, error) {
	// Dividir el contenido en líneas
	lineas := strings.Split(contenido, "\n")
	for i, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) < 3 {
			continue // Ignorar líneas mal formadas
		}

		tipo, nombre := campos[1], campos[2]
		// Buscar coincidencias exactas con el tipo y nombre
		if tipo == entityType && nombre == name {
			return linea, i, nil // Devolver la línea y el índice
		}
	}

	return "", -1, fmt.Errorf("%s '%s' no encontrado en users.txt", entityType, name)
}
