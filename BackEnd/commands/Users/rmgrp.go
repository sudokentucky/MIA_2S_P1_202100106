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

	// Marcar el grupo como eliminado (cambiar el ID a "0")
	err = globals.RemoveFromUsersFile(file, sb, &usersInode, rmgrp.Name, "G")
	if err != nil {
		return fmt.Errorf("error eliminando el grupo '%s': %v", rmgrp.Name, err)
	}

	// Eliminar todos los usuarios asociados al grupo
	err = RemoveUsersFromGroup(file, sb, &usersInode, rmgrp.Name)
	if err != nil {
		return fmt.Errorf("error eliminando los usuarios asociados al grupo '%s': %v", rmgrp.Name, err)
	}

	// Actualizar el inodo de users.txt en el archivo
	err = usersInode.Encode(file, inodeOffset) // Usar el descriptor de archivo
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	// Mostrar el contenido de users.txt después de eliminar el grupo y sus usuarios
	contenido, err := globals.ReadFileBlocks(file, sb, &usersInode)
	if err != nil {
		return fmt.Errorf("error leyendo el contenido de users.txt: %v", err)
	}
	fmt.Fprintln(outputBuffer, "\nContenido de users.txt después de eliminar el grupo y usuarios:")
	fmt.Fprintln(outputBuffer, contenido)

	// Mostrar mensaje de éxito
	fmt.Fprintf(outputBuffer, "Grupo '%s' eliminado exitosamente, junto con sus usuarios.\n", rmgrp.Name)

	return nil
}

// RemoveUsersFromGroup : Elimina los usuarios asociados a un grupo
func RemoveUsersFromGroup(file *os.File, sb *structs.Superblock, usersInode *structs.Inode, groupName string) error {
	// Leer el contenido de users.txt
	contenido, err := globals.ReadFileBlocks(file, sb, usersInode)
	if err != nil {
		return fmt.Errorf("error leyendo el contenido de users.txt: %v", err)
	}

	// Separar las líneas del archivo
	lineas := strings.Split(contenido, "\n")
	modificado := false

	// Recorrer las líneas y eliminar usuarios asociados al grupo
	for i, linea := range lineas {
		if linea == "" {
			continue
		}
		partes := strings.Split(linea, ",")
		if len(partes) == 5 && partes[2] == groupName {
			// Marcar el usuario como eliminado
			partes[0] = "0"
			lineas[i] = strings.Join(partes, ",")
			modificado = true
		}
	}

	// Si se ha modificado alguna línea, guardar los cambios
	if modificado {
		contenidoActualizado := strings.Join(lineas, "\n")
		err = globals.WriteUsersBlocks(file, sb, usersInode, contenidoActualizado)
		if err != nil {
			return fmt.Errorf("error guardando los cambios en users.txt: %v", err)
		}
	}

	return nil
}
