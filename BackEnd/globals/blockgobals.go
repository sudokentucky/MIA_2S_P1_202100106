package globals

import (
	structs "backend/Structs"
	"fmt"
	"os"
	"strings"
)

// ReadFileBlocks lee todos los bloques asignados a un archivo y devuelve su contenido completo
func ReadFileBlocks(file *os.File, sb *structs.Superblock, inode *structs.Inode) (string, error) {
	var contenido string

	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break // No hay más bloques asignados
		}

		blockOffset := int64(sb.S_block_start + blockIndex*64) // Calcular la posición del bloque
		var fileBlock structs.FileBlock

		// Leer el bloque desde el archivo
		err := fileBlock.Decode(file, blockOffset)
		if err != nil {
			return "", fmt.Errorf("error leyendo bloque %d: %w", blockIndex, err)
		}

		// Concatenar el contenido del bloque al resultado total
		contenido += string(fileBlock.B_content[:])
	}

	// Actualizar el tiempo de último acceso
	inode.UpdateAtime()

	// Guardar el inodo actualizado en el archivo
	inodeOffset := int64(sb.S_inode_start) + int64(inode.I_block[0])*int64(sb.S_inode_size)
	err := inode.Encode(file, inodeOffset)
	if err != nil {
		return "", fmt.Errorf("error actualizando el inodo: %w", err)
	}

	return strings.TrimRight(contenido, "\x00"), nil
}

// assignNewBlock asigna un nuevo bloque al inodo si es necesario
func assignNewBlock(file *os.File, sb *structs.Superblock, inode *structs.Inode) (int32, error) {
	fmt.Println("=== Iniciando la asignación de un nuevo bloque ===")
	for i := 0; i < len(inode.I_block); i++ {
		if inode.I_block[i] == -1 {
			// Asignar un nuevo bloque si no hay bloque actual
			newBlock, err := sb.FindNextFreeBlock(file)
			if err != nil {
				return -1, fmt.Errorf("error asignando nuevo bloque: %w", err)
			}

			// Verificar si el bloque recién asignado es válido
			if newBlock == -1 {
				return -1, fmt.Errorf("bloque asignado inválido")
			}

			inode.I_block[i] = newBlock

			// Actualizar el bitmap de bloques
			if err := sb.UpdateBitmapBlock(file); err != nil {
				return -1, fmt.Errorf("error actualizando bitmap de bloques: %w", err)
			}

			// Confirmar asignación del nuevo bloque
			fmt.Printf("Nuevo bloque asignado: %d (índice I_block[%d])\n", newBlock, i)
			return newBlock, nil //retornar el índice del bloque asignado
		}

	}
	// Error si todos los bloques están llenos
	fmt.Println("Error: Todos los bloques asignados están llenos, no hay más espacio en el inodo.")
	return -1, fmt.Errorf("todos los bloques asignados están llenos")
}

// Funcion que escribe el contenido en los bloques de users.txt, recibe el archivo, el superbloque, el inodo y el nuevo contenido
func WriteUsersBlocks(file *os.File, sb *structs.Superblock, inode *structs.Inode, nuevoContenido string) error {
	// Leer el contenido actual de los bloques asignados al inodo
	contenidoExistente, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return fmt.Errorf("error leyendo contenido existente de users.txt: %w", err)
	}

	// Combinar el contenido existente con el nuevo contenido
	contenidoTotal := contenidoExistente + nuevoContenido

	// Obtener el tamaño actual del contenido existente
	sizeContenidoExistente := len(contenidoExistente)

	// Dividir el contenido total en bloques de 64 bytes
	blocks, err := structs.SplitContent(contenidoTotal)
	if err != nil {
		return fmt.Errorf("error al dividir el contenido en bloques: %w", err)
	}

	// Iterar sobre los bloques generados y escribirlos en los bloques del inodo
	for i, block := range blocks {
		// Si hemos alcanzado el límite de bloques asignados al inodo, necesitamos asignar un nuevo bloque
		if i >= len(inode.I_block) || inode.I_block[i] == -1 {
			// Intentar asignar un nuevo bloque al inodo
			newBlockIndex, err := assignNewBlock(file, sb, inode)
			if err != nil {
				return fmt.Errorf("error asignando nuevo bloque: %w", err)
			}
			inode.I_block[i] = newBlockIndex
		}

		// Calcular el offset del bloque en el archivo
		blockOffset := int64(sb.S_block_start + inode.I_block[i]*64)

		// Escribir el contenido del bloque en la partición
		err = block.Encode(file, blockOffset)
		if err != nil {
			return fmt.Errorf("error escribiendo el bloque %d: %w", inode.I_block[i], err)
		}
	}

	// Incrementar el tamaño del archivo en el inodo (i_size)
	nuevoTamano := sizeContenidoExistente + len(nuevoContenido)
	inode.I_size = int32(nuevoTamano)

	// Actualizar los tiempos de modificación y cambio
	inode.UpdateMtime()
	inode.UpdateCtime()

	// Guardar los cambios del inodo en el archivo
	inodeOffset := int64(sb.S_inode_start) + int64(inode.I_block[0])*int64(sb.S_inode_size)
	err = inode.Encode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error actualizando el inodo: %w", err)
	}

	return nil
}
func InsertIntoUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, entry string) error {
	// Leer todo el contenido de los bloques actuales
	contenidoActual, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return fmt.Errorf("error leyendo contenido existente de users.txt: %w", err)
	}

	// Agregar la nueva entrada al contenido en memoria
	contenidoNuevo := contenidoActual + entry + "\n"

	// Dividir el contenido en bloques de 64 bytes
	blocks, err := structs.SplitContent(contenidoNuevo)
	if err != nil {
		return fmt.Errorf("error dividiendo el contenido en bloques: %w", err)
	}

	// Escribir los bloques nuevamente en el sistema de archivos
	for i, block := range blocks {
		// Si necesitamos un nuevo bloque, asignarlo
		if i >= len(inode.I_block) || inode.I_block[i] == -1 {
			newBlockIndex, err := assignNewBlock(file, sb, inode)
			if err != nil {
				return fmt.Errorf("error asignando nuevo bloque: %w", err)
			}
			inode.I_block[i] = newBlockIndex
		}

		// Calcular el offset y escribir el contenido del bloque
		blockOffset := int64(sb.S_block_start + inode.I_block[i]*64)
		err = block.Encode(file, blockOffset)
		if err != nil {
			return fmt.Errorf("error escribiendo bloque %d: %w", inode.I_block[i], err)
		}
	}

	// Actualizar el tamaño del archivo
	inode.I_size = int32(len(contenidoNuevo))

	// Actualizar tiempos de modificación
	inode.UpdateMtime()
	inode.UpdateCtime()

	// Guardar el inodo actualizado en el archivo
	inodeOffset := int64(sb.S_inode_start) + int64(inode.I_block[0])*int64(sb.S_inode_size)
	err = inode.Encode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error actualizando inodo: %w", err)
	}

	return nil
}

// AddEntryToUsersFile añade una entrada al archivo users.txt (ya sea grupo o usuario) con depuración
func AddEntryToUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, entry, name, entityType string) error {
	// Leer el contenido actual de users.txt
	contenidoActual, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return fmt.Errorf("error leyendo blocks de users.txt: %w", err)
	}

	// Verificar si el grupo/usuario ya existe
	_, _, err = findLineInUsersFile(contenidoActual, name, entityType)
	if err == nil {
		// Si ya existe, no se crea el grupo/usuario, se retorna sin hacer nada
		fmt.Printf("El %s '%s' ya existe en users.txt\n", entityType, name)
		return nil
	}
	fmt.Println("=== Escribiendo nuevo contenido en users.txt ===")
	fmt.Println(entry) // Solo imprimimos la nueva entrada

	// Escribir solo la nueva entrada al final de los bloques
	err = WriteUsersBlocks(file, sb, inode, entry+"\n") // Solo el nuevo grupo
	if err != nil {
		return fmt.Errorf("error agregando entrada a users.txt: %w", err)
	}

	// Depuración: Mostrar el estado del inodo después de la modificación
	fmt.Println("\n=== Estado del inodo después de la modificación ===")
	sb.PrintInodes(file.Name())

	// Depuración: Mostrar el estado de los bloques después de la modificación
	fmt.Println("\n=== Estado de los bloques después de la modificación ===")
	sb.PrintBlocks(file.Name())

	return nil
}

// CreateGroup añade un nuevo grupo en el archivo users.txt
func CreateGroup(file *os.File, sb *structs.Superblock, inode *structs.Inode, groupName string) error {
	groupEntry := fmt.Sprintf("%d,G,%s", sb.S_inodes_count+1, groupName)
	return AddEntryToUsersFile(file, sb, inode, groupEntry, groupName, "G")
}

// CreateUser añade un nuevo usuario en el archivo users.txt
func CreateUser(file *os.File, sb *structs.Superblock, inode *structs.Inode, userName, userPassword, groupName string) error {
	userEntry := fmt.Sprintf("%d,U,%s,%s,%s", sb.S_inodes_count+1, userName, groupName, userPassword)
	return AddEntryToUsersFile(file, sb, inode, userEntry, userName, "U")
}

func RemoveFromUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, name, entityType string) error {
	contenido, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return err
	}

	// Buscar la línea que queremos eliminar
	linea, index, err := findLineInUsersFile(contenido, name, entityType)
	if err != nil {
		return err
	}

	// Marcar la entrada como eliminada
	lineas := strings.Split(contenido, "\n")
	campos := strings.Split(linea, ",")
	if len(campos) > 0 && campos[0] != "0" {
		campos[0] = "0"
		lineas[index] = strings.Join(campos, ",")
	}

	// Escribir el nuevo contenido
	return WriteUsersBlocks(file, sb, inode, strings.Join(lineas, "\n"))
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

// findLineInUsersFile busca una línea en el archivo users.txt según nombre y tipo
func findLineInUsersFile(contenido string, name, entityType string) (string, int, error) {
	// Dividir el contenido en líneas
	lineas := strings.Split(contenido, "\n")

	// Depuración: Imprimir el contenido que estamos procesando
	fmt.Println("Contenido de users.txt:")
	fmt.Println(contenido)

	for i, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) < 3 {
			continue // Ignorar líneas mal formadas
		}

		tipo, nombre := strings.TrimSpace(campos[1]), strings.TrimSpace(campos[2])

		// Depuración: Mostrar qué tipo y nombre estamos comparando
		fmt.Printf("Comparando: tipo='%s', nombre='%s'\n", tipo, nombre)

		// Buscar coincidencias exactas con el tipo y nombre
		if tipo == entityType && nombre == name {
			return linea, i, nil // Devolver la línea y el índice
		}
	}

	return "", -1, fmt.Errorf("%s '%s' no encontrado en users.txt", entityType, name)
}
