package commands

import (
	structs "backend/Structs"
	globals "backend/globals"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// CHGRP : Estructura para el comando CHGRP
type CHGRP struct {
	User string
	Grp  string
}

// ParserChgrp : Parseo de argumentos para el comando chgrp
func ParserChgrp(tokens []string) (string, error) {
	// Inicializar el comando CHGRP
	var outputBuffer strings.Builder
	cmd := &CHGRP{}

	// Expresión regular para encontrar los parámetros -user y -grp
	reUser := regexp.MustCompile(`-usr=[^\s]+`)
	reGrp := regexp.MustCompile(`-grp=[^\s]+`)

	// Buscar los parámetros
	matchesUser := reUser.FindString(strings.Join(tokens, " "))
	matchesGrp := reGrp.FindString(strings.Join(tokens, " "))

	if matchesUser == "" {
		return "", fmt.Errorf("falta el parámetro -usr")
	}
	if matchesGrp == "" {
		return "", fmt.Errorf("falta el parámetro -grp")
	}

	// Extraer los valores de los parámetros
	cmd.User = strings.SplitN(matchesUser, "=", 2)[1]
	cmd.Grp = strings.SplitN(matchesGrp, "=", 2)[1]

	// Ejecutar la lógica del comando chgrp
	err := commandChgrp(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	return outputBuffer.String(), nil
}

// commandChgrp : Ejecuta el comando CHGRP
func commandChgrp(chgrp *CHGRP, outputBuffer *strings.Builder) error {
	fmt.Fprintln(outputBuffer, "======================= CHGRP =======================")
	// Verificar si hay una sesión activa y si el usuario es root
	if !globals.IsLoggedIn() {
		return fmt.Errorf("no hay ninguna sesión activa")
	}
	if globals.UsuarioActual.Name != "root" {
		return fmt.Errorf("solo el usuario root puede ejecutar este comando")
	}

	// Verificar que la partición esté montada
	partition, path, err := globals.GetMountedPartition(globals.UsuarioActual.Id)
	if err != nil {
		return fmt.Errorf("no se puede encontrar la partición montada: %v", err)
	}

	// Abrir el archivo de la partición
	file, err := os.OpenFile(path, os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("no se puede abrir el archivo de la partición: %v", err)
	}
	defer file.Close()

	// Cargar el Superblock usando el descriptor de archivo
	_, sb, _, err := globals.GetMountedPartitionRep(globals.UsuarioActual.Id)
	if err != nil {
		return fmt.Errorf("no se pudo cargar el Superblock: %v", err)
	}

	// Leer el inodo de users.txt
	var usersInode structs.Inode
	inodeOffset := int64(sb.S_inode_start + int32(binary.Size(usersInode))) //ubuacion de los bloques de users.txt
	err = usersInode.Decode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

	// Cambiar el grupo del usuario
	err = ChangeUserGroup(file, sb, &usersInode, chgrp.User, chgrp.Grp)
	if err != nil {
		return fmt.Errorf("error cambiando el grupo del usuario '%s': %v", chgrp.User, err)
	}

	//Guardar el superbloque
	err = sb.Encode(file, int64(partition.Part_start))
	if err != nil {
		return fmt.Errorf("error guardando el superbloque: %v", err)
	}

	// Mensaje de confirmación
	fmt.Fprintf(outputBuffer, "El grupo del usuario '%s' ha sido cambiado exitosamente a '%s'\n", chgrp.User, chgrp.Grp)
	fmt.Println("\nInodos")
	sb.PrintInodes(file.Name())
	fmt.Println("\nBloques")
	sb.PrintBlocks(file.Name())
	fmt.Fprintln(outputBuffer, "=====================================================")
	return nil
}

// ChangeUserGroup : Cambia el grupo de un usuario en el archivo users.txt y reorganiza el contenido
func ChangeUserGroup(file *os.File, sb *structs.Superblock, usersInode *structs.Inode, userName, newGroup string) error {
	// Leer el contenido actual de users.txt
	contenidoActual, err := globals.ReadFileBlocks(file, sb, usersInode)
	if err != nil {
		return fmt.Errorf("error leyendo el contenido de users.txt: %w", err)
	}

	// Eliminar líneas vacías o con espacios innecesarios del contenido actual
	lineas := strings.Split(strings.TrimSpace(contenidoActual), "\n")
	var nuevoContenido []string
	var usuarioModificado bool
	var grupoEncontrado bool

	// Procesar el contenido del archivo y convertir a objetos de User y Group
	var usuarios []structs.User
	var grupos []structs.Group

	// Separar usuarios y grupos
	for _, linea := range lineas {
		partes := strings.Split(linea, ",")
		if len(partes) < 3 {
			continue // Saltar líneas mal formadas
		}

		// Identificar si es un grupo o un usuario
		tipo := strings.TrimSpace(partes[1])
		if tipo == "G" {
			// Crear un objeto de tipo Group
			group := structs.NewGroup(partes[0], partes[2])
			grupos = append(grupos, *group)
		} else if tipo == "U" && len(partes) >= 5 {
			// Crear un objeto de tipo User
			user := structs.NewUser(partes[0], partes[2], partes[3], partes[4])
			usuarios = append(usuarios, *user)
		}
	}

	// Verificar si el nuevo grupo existe y no está eliminado
	var nuevoIDGrupo string
	for _, group := range grupos {
		if group.Group == newGroup && group.GID != "0" {
			nuevoIDGrupo = group.GID
			grupoEncontrado = true
			break
		}
	}

	if !grupoEncontrado {
		return fmt.Errorf("el grupo '%s' no existe o está eliminado", newGroup)
	}

	// Modificar el grupo del usuario si existe
	for i, usuario := range usuarios {
		if usuario.Name == userName && usuario.Id != "0" { // Verificar que el usuario no esté eliminado
			// Cambiar el grupo del usuario y actualizar su ID al ID del nuevo grupo
			fmt.Printf("Cambiando el grupo del usuario '%s' al grupo '%s' (ID grupo: %s)\n", usuario.Name, newGroup, nuevoIDGrupo)
			usuarios[i].Group = newGroup
			usuarios[i].Id = nuevoIDGrupo // Cambiar el ID del usuario al ID del grupo destino
			fmt.Printf("Nuevo estado del usuario: %s\n", usuarios[i].ToString())
			usuarioModificado = true
		}
	}

	if !usuarioModificado {
		return fmt.Errorf("el usuario '%s' no existe o está eliminado", userName)
	}

	// Reorganizar el contenido para agrupar los usuarios bajo sus grupos correspondientes
	for _, group := range grupos {
		nuevoContenido = append(nuevoContenido, group.ToString()) // Agregar grupo al contenido

		// Agregar usuarios asociados al grupo
		for _, usuario := range usuarios {
			if usuario.Group == group.Group {
				nuevoContenido = append(nuevoContenido, usuario.ToString())
			}
		}
	}

	// Limpiar los bloques asignados antes de escribir el nuevo contenido
	for _, blockIndex := range usersInode.I_block {
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

	// Reescribir el contenido agrupado en los bloques de `users.txt`
	// Aquí aseguramos que el contenido se divida correctamente entre los bloques
	err = WriteContentToBlocks(file, sb, usersInode, nuevoContenido)
	if err != nil {
		return fmt.Errorf("error guardando los cambios en users.txt: %v", err)
	}

	// Actualizar el tamaño del archivo (i_size)
	usersInode.I_size = int32(len(strings.Join(nuevoContenido, "\n")))

	// Actualizar tiempos de modificación y cambio
	usersInode.UpdateMtime()
	usersInode.UpdateCtime()

	// Guardar el inodo actualizado en el archivo
	inodeOffset := int64(sb.S_inode_start + int32(binary.Size(*usersInode)))
	err = usersInode.Encode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %w", err)
	}

	return nil
}

// WriteContentToBlocks escribe el contenido de users.txt dividiéndolo en bloques correctamente
func WriteContentToBlocks(file *os.File, sb *structs.Superblock, usersInode *structs.Inode, contenido []string) error {
	// Convertir el contenido en una cadena
	contenidoFinal := strings.Join(contenido, "\n") + "\n"
	data := []byte(contenidoFinal)

	// Tamaño máximo del bloque
	blockSize := int(sb.S_block_size)

	// Escribir el contenido por bloques
	for i, blockIndex := range usersInode.I_block {
		if blockIndex == -1 {
			break // No hay más bloques asignados
		}

		// Dividir los datos en bloques de tamaño máximo
		start := i * blockSize
		end := start + blockSize
		if end > len(data) {
			end = len(data)
		}

		// Crear un bloque con el contenido correspondiente
		var fileBlock structs.FileBlock
		copy(fileBlock.B_content[:], data[start:end])

		// Mostrar el bloque escrito para depuración
		fmt.Printf("Escribiendo bloque %d: %s\n", blockIndex, string(fileBlock.B_content[:]))
	}

	return nil
}
