package globals

import (
	structs "backend/Structs"
	"encoding/binary"
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

	// Eliminar caracteres nulos (\x00) del final del contenido
	return strings.TrimRight(contenido, "\x00"), nil
}

// blockHasSpace verifica si un bloque tiene espacio disponible
func blockHasSpace(file *os.File, sb *structs.Superblock, blockIndex int32) (bool, int, error) {
	blockOffset := int64(sb.S_block_start + blockIndex*64)
	var fileBlock structs.FileBlock

	// Leer el bloque actual
	err := fileBlock.Decode(file, blockOffset)
	if err != nil {
		return false, 0, fmt.Errorf("error leyendo bloque: %w", err)
	}

	// Calcular el espacio usado en el bloque
	usedSpace := 0
	for _, b := range fileBlock.B_content {
		if b != 0 {
			usedSpace++
		}
	}

	// Verificar si el bloque tiene espacio disponible
	return usedSpace < 64, usedSpace, nil
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
			return newBlock, nil
		}

		// Verificar si el bloque actual tiene espacio disponible
		hasSpace, usedSpace, err := blockHasSpace(file, sb, inode.I_block[i])
		fmt.Println("Bloque actual:", inode.I_block[i]) // Depuración
		if err != nil {
			return -1, err
		}
		fmt.Printf("Bloque %d verificado, espacio usado: %d/64\n", inode.I_block[i], usedSpace) // Depuración
		if hasSpace {
			fmt.Println("El bloque aun tiene espacio, se usara el mismo") // Si hay espacio disponible, devolver el bloque actual
			return inode.I_block[i], nil                                  // Devolver el bloque actual si tiene espacio disponible

		}
	}

	// Error si todos los bloques están llenos
	fmt.Println("Error: Todos los bloques asignados están llenos, no hay más espacio en el inodo.")
	return -1, fmt.Errorf("todos los bloques asignados están llenos")
}

// WriteFileBlocks escribe el contenido completo de un archivo en bloques asignados
func WriteFileBlocks(file *os.File, sb *structs.Superblock, inode *structs.Inode, nuevoContenido string) error {
	// Leer el contenido actual existente en los bloques
	contenidoExistente, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return fmt.Errorf("error leyendo contenido existente de users.txt: %w", err)
	}

	// Concatenar el contenido nuevo al contenido existente
	contenido := contenidoExistente + nuevoContenido
	data := []byte(contenido) // Convertir todo el contenido concatenado a bytes
	bytesEscritos := 0        // Contador de bytes escritos

	// Recorrer los bloques actuales y agregar el contenido nuevo
	for bytesEscritos < len(data) {
		// Buscar el bloque con espacio disponible o asignar un nuevo bloque si es necesario
		blockIndex, err := assignNewBlock(file, sb, inode) // Asignar un nuevo bloque si es necesario
		if err != nil {
			return err
		}

		blockOffset := int64(sb.S_block_start + blockIndex*64) // Calcular la posición del bloque
		var fileBlock structs.FileBlock                        // Crear un bloque temporal para leer y escribir

		// Leer el bloque actual solo si ya está en uso y no está lleno
		if inode.I_block[blockIndex] != -1 {
			if err := fileBlock.Decode(file, blockOffset); err != nil {
				return fmt.Errorf("error leyendo bloque: %w", err)
			}
		}

		// Calcular el espacio usado y el espacio disponible en el bloque
		usedSpace := len(strings.TrimRight(string(fileBlock.B_content[:]), "\x00"))
		spaceLeft := 64 - usedSpace

		// Si no queda espacio en el bloque, saltar al siguiente bloque
		if spaceLeft == 0 {
			continue
		}

		// Escribir solo la parte que cabe en el bloque actual
		bytesToWrite := min(spaceLeft, len(data[bytesEscritos:]))
		copy(fileBlock.B_content[usedSpace:], data[bytesEscritos:bytesEscritos+bytesToWrite])

		// Guardar el bloque actualizado en el archivo
		if err := fileBlock.Encode(file, blockOffset); err != nil {
			return fmt.Errorf("error escribiendo bloque: %w", err)
		}

		// Avanzar el puntero de escritura
		bytesEscritos += bytesToWrite

		// Si el bloque se llenó y queda contenido por escribir, asignar un nuevo bloque
		if bytesEscritos < len(data) && spaceLeft == 0 {
			blockAssigned := false
			for i := range inode.I_block {
				if inode.I_block[i] == -1 {
					newBlockIndex, err := assignNewBlock(file, sb, inode)
					if err != nil {
						return fmt.Errorf("error asignando nuevo bloque: %w", err)
					}
					inode.I_block[i] = newBlockIndex
					blockAssigned = true
					break
				}
			}
			if !blockAssigned {
				return fmt.Errorf("error: no se pudo asignar un nuevo bloque")
			}
		}
	}

	// Actualizar el tamaño del inodo con el nuevo contenido (contenido total)
	inode.I_size = int32(len(contenido))

	// Guardar el inodo actualizado
	err = inode.Encode(file, int64(sb.S_inode_start+int32(binary.Size(*inode))))
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

	// Aquí evitamos volver a leer o concatenar el contenido anterior.
	// Solo vamos a agregar el nuevo grupo al final de los bloques.

	fmt.Println("=== Escribiendo nuevo contenido en users.txt ===")
	fmt.Println(entry) // Solo imprimimos la nueva entrada

	// Escribir solo la nueva entrada al final de los bloques
	err = WriteFileBlocks(file, sb, inode, entry+"\n") // Solo el nuevo grupo
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
	return WriteFileBlocks(file, sb, inode, strings.Join(lineas, "\n"))
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
