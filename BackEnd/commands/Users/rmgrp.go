package commands

import (
	structs "backend/Structs"
	globals "backend/globals"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// RMGRP : Estructura para el comando RMGRP
type RMGRP struct {
	Name string
}

// ParserRmgrp : Parseo de argumentos para el comando rmgrp y captura de mensajes importantes
func ParserRmgrp(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer // Buffer para capturar los mensajes importantes para el usuario

	// Inicializar el comando RMGRP
	cmd := &RMGRP{}

	// Expresión regular para encontrar el parámetro -name
	re := regexp.MustCompile(`-name=[^\s]+`)
	matches := re.FindString(strings.Join(tokens, " "))

	if matches == "" {
		return "", fmt.Errorf("falta el parámetro -name")
	}

	// Extraer el valor del parámetro -name
	param := strings.SplitN(matches, "=", 2)
	if len(param) != 2 {
		return "", fmt.Errorf("formato incorrecto para -name")
	}
	cmd.Name = param[1]

	// Ejecutar la lógica del comando rmgrp
	err := commandRmgrp(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	// Retornar los mensajes importantes capturados en el buffer
	return outputBuffer.String(), nil
}

// commandRmgrp : Ejecuta el comando RMGRP con captura de mensajes importantes en el buffer
func commandRmgrp(rmgrp *RMGRP, outputBuffer *bytes.Buffer) error {
	fmt.Fprintln(outputBuffer, "======================= RMGRP =======================")
	// Verificar si hay una sesión activa y si el usuario es root
	if !globals.IsLoggedIn() {
		return fmt.Errorf("no hay ninguna sesión activa")
	}
	if globals.UsuarioActual.Name != "root" {
		return fmt.Errorf("solo el usuario root puede ejecutar este comando")
	}

	// Verificar que la partición esté montada
	_, path, err := globals.GetMountedPartition(globals.UsuarioActual.Id)
	if err != nil {
		return fmt.Errorf("no se puede encontrar la partición montada: %v", err)
	}

	// Abrir el archivo de la partición
	file, err := os.OpenFile(path, os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("no se puede abrir el archivo de la partición: %v", err)
	}
	defer file.Close()

	// Cargar el Superblock
	_, sb, _, err := globals.GetMountedPartitionRep(globals.UsuarioActual.Id)
	if err != nil {
		return fmt.Errorf("no se pudo cargar el Superblock: %v", err)
	}

	// Leer el inodo de users.txt
	var usersInode structs.Inode
	inodeOffset := int64(sb.S_inode_start + int32(binary.Size(usersInode))) //posicion del contenido de users.txt
	err = usersInode.Decode(file, inodeOffset)                              // Usar el descriptor de archivo
	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

	// Verificar si el grupo existe
	_, err = globals.FindInUsersFile(file, sb, &usersInode, rmgrp.Name, "G")
	if err != nil {
		return fmt.Errorf("el grupo '%s' no existe", rmgrp.Name)
	}

	// Cambiar el estado del grupo y de los usuarios asociados en una sola llamada
	err = UpdateEntityStateOrRemoveUsers(file, sb, &usersInode, rmgrp.Name, "G", "0")
	if err != nil {
		return fmt.Errorf("error eliminando el grupo y usuarios asociados: %v", err)
	}

	// Actualizar el inodo de users.txt en el archivo
	err = usersInode.Encode(file, inodeOffset) // Usar el descriptor de archivo
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	// Mostrar mensaje de éxito
	fmt.Fprintf(outputBuffer, "Grupo '%s' eliminado exitosamente, junto con sus usuarios.\n", rmgrp.Name)
	fmt.Println("\nInodos actualizados:")
	sb.PrintInodes(file.Name())
	fmt.Println("\nBloques de datos actualizados:")
	sb.PrintBlocks(file.Name())

	fmt.Fprintln(outputBuffer, "=====================================================")

	return nil
}

// UpdateEntityStateOrRemoveUsers : Cambia el estado de un grupo/usuario y elimina usuarios asociados a un grupo
func UpdateEntityStateOrRemoveUsers(file *os.File, sb *structs.Superblock, usersInode *structs.Inode, name string, entityType string, newState string) error {
	// Leer el contenido actual de users.txt
	contenido, err := globals.ReadFileBlocks(file, sb, usersInode)
	if err != nil {
		return fmt.Errorf("error leyendo el contenido de users.txt: %v", err)
	}

	// Separar las líneas del archivo
	lineas := strings.Split(contenido, "\n")
	modificado := false

	// Variable para detectar si es un grupo
	var groupName string
	if entityType == "G" {
		groupName = name
	}

	// Recorrer las líneas para actualizar el estado de un grupo/usuario y eliminar usuarios asociados si es un grupo
	for i, linea := range lineas {
		linea = strings.TrimSpace(linea) // Eliminar espacios en blanco adicionales
		if linea == "" {
			continue
		}

		partes := strings.Split(linea, ",")
		if len(partes) < 3 {
			continue // Ignorar líneas mal formadas
		}

		tipo := partes[1]
		nombre := partes[2]

		// Verificar si coincide el tipo de entidad (usuario o grupo) y el nombre
		if tipo == entityType && nombre == name {
			// Cambiar el estado del grupo o usuario
			partes[0] = newState
			lineas[i] = strings.Join(partes, ",")
			modificado = true

			// Si es un grupo, busca y elimina a los usuarios asociados
			if entityType == "G" {
				// Recorrer de nuevo todas las líneas para eliminar usuarios de ese grupo
				for j, lineaUsuario := range lineas {
					lineaUsuario = strings.TrimSpace(lineaUsuario)
					if lineaUsuario == "" {
						continue
					}
					partesUsuario := strings.Split(lineaUsuario, ",")
					if len(partesUsuario) == 5 && partesUsuario[2] == groupName {
						// Marcar el usuario como eliminado
						partesUsuario[0] = "0"
						lineas[j] = strings.Join(partesUsuario, ",")
					}
				}
			}
			break // Solo necesitamos modificar una entrada del grupo/usuario
		}
	}

	// Si se modificó alguna línea, guardar los cambios en el archivo
	if modificado {
		contenidoActualizado := strings.Join(lineas, "\n")

		// Limpiar los bloques asignados al archivo antes de escribir
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

		// Reescribir todo el contenido en los bloques después de limpiar
		err = globals.WriteUsersBlocks(file, sb, usersInode, contenidoActualizado)
		if err != nil {
			return fmt.Errorf("error guardando los cambios en users.txt: %v", err)
		}
	} else {
		return fmt.Errorf("%s '%s' no encontrado en users.txt", entityType, name)
	}

	return nil
}
