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

// InsertIntoUsersFile inserta una nueva entrada en el archivo users.txt
func InsertIntoUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, entry string) error {
	// Leer el contenido actual de los bloques asignados al inodo
	contenidoActual, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return fmt.Errorf("error leyendo el contenido de users.txt: %w", err)
	}

	// Eliminar líneas vacías o con espacios innecesarios del contenido actual
	lineas := strings.Split(strings.TrimSpace(contenidoActual), "\n")

	// Obtener el grupo desde la nueva entrada
	partesEntry := strings.Split(entry, ",")
	if len(partesEntry) < 4 { // Se espera al menos UID, U, Grupo, Usuario, Contraseña
		return fmt.Errorf("entrada de usuario inválida: %s", entry)
	}
	userGrupo := partesEntry[2] // El grupo del usuario se encuentra en la tercera posición

	// Buscar el ID del grupo correspondiente en el contenido actual
	var groupID string
	var nuevoContenido []string
	usuarioInsertado := false

	// Recorrer las líneas de `users.txt` para encontrar el grupo correspondiente
	for _, linea := range lineas {
		partes := strings.Split(linea, ",")
		// Agregar la línea actual al nuevo contenido
		nuevoContenido = append(nuevoContenido, strings.TrimSpace(linea))

		// Si encontramos el grupo correcto
		if len(partes) > 2 && partes[1] == "G" && partes[2] == userGrupo {
			groupID = partes[0] // Obtener el ID del grupo

			// Insertar el usuario justo después del grupo si no se ha insertado ya
			if groupID != "" && !usuarioInsertado {
				usuarioConGrupo := fmt.Sprintf("%s,U,%s,%s,%s", groupID, partesEntry[2], partesEntry[3], partesEntry[4])
				nuevoContenido = append(nuevoContenido, usuarioConGrupo)
				usuarioInsertado = true
			}
		}
	}

	// Verificar si el grupo fue encontrado
	if groupID == "" {
		return fmt.Errorf("el grupo '%s' no existe", userGrupo)
	}

	// Combinar todas las líneas en un solo contenido para escribir en el archivo, eliminando posibles líneas en blanco
	contenidoNuevo := strings.Join(nuevoContenido, "\n") + "\n"
	fmt.Println("=== Escribiendo nuevo contenido en users.txt ===")
	fmt.Println(contenidoNuevo)

	// Limpiar los bloques asignados al archivo
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break // No hay más bloques asignados
		}

		blockOffset := int64(sb.S_block_start + blockIndex*sb.S_block_size)
		var fileBlock structs.FileBlock

		// Limpiar el contenido del bloque
		fileBlock.ClearContent()

		// Escribir el bloque vacío de nuevo
		err = fileBlock.Encode(file, blockOffset)
		if err != nil {
			return fmt.Errorf("error escribiendo bloque limpio %d: %w", blockIndex, err)
		}
	}

	// Reescribir todo el contenido línea por línea
	err = WriteUsersBlocks(file, sb, inode, contenidoNuevo)
	if err != nil {
		return fmt.Errorf("error escribiendo el nuevo contenido en users.txt: %w", err)
	}

	// Actualizar el tamaño del archivo (i_size)
	inode.I_size = int32(len(contenidoNuevo))

	// Actualizar tiempos de modificación y cambio
	inode.UpdateMtime()
	inode.UpdateCtime()

	// Guardar el inodo actualizado en el archivo
	inodeOffset := int64(sb.S_inode_start + int32(binary.Size(*inode)))
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

// FindInUsersFile busca una entrada en el archivo users.txt según nombre y tipo
func FindInUsersFile(file *os.File, sb *structs.Superblock, inode *structs.Inode, name, entityType string) (string, error) {
	contenido, err := ReadFileBlocks(file, sb, inode)
	if err != nil {
		return "", err
	}

	// Usamos la función auxiliar para buscar la línea
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

		// Determinar si es un grupo o un usuario según el entityType
		if entityType == "G" && len(campos) == 3 {
			// Es un grupo
			grupo := structs.NewGroup(campos[0], campos[2]) // Crear instancia de Group
			if grupo.Tipo == entityType && grupo.Group == name {
				return grupo.ToString(), i, nil // Devolver la línea y el índice
			}
		} else if entityType == "U" && len(campos) == 5 {
			// Es un usuario
			usuario := structs.NewUser(campos[0], campos[2], campos[3], campos[4]) // Crear instancia de User
			if usuario.Tipo == entityType && usuario.Name == name {
				return usuario.ToString(), i, nil // Devolver la línea y el índice
			}
		}
	}

	return "", -1, fmt.Errorf("%s '%s' no encontrado en users.txt", entityType, name)
}
