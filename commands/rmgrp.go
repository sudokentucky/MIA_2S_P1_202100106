package commands

import (
	structs "ArchivosP1/Structs"
	globals "ArchivosP1/globals"
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

// ParserRmgrp : Parseo de argumentos para el comando rmgrp
func ParserRmgrp(tokens []string) (*RMGRP, error) {
	// Inicializar el comando RMGRP
	cmd := &RMGRP{}

	// Expresión regular para encontrar el parámetro -name
	re := regexp.MustCompile(`-name=[^\s]+`)
	matches := re.FindString(strings.Join(tokens, " "))

	if matches == "" {
		return nil, fmt.Errorf("falta el parámetro -name")
	}

	// Extraer el valor del parámetro -name
	param := strings.SplitN(matches, "=", 2)
	if len(param) != 2 {
		return nil, fmt.Errorf("formato incorrecto para -name")
	}
	cmd.Name = param[1]

	// Ejecutar la lógica del comando rmgrp
	err := commandRmgrp(cmd)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

// commandRmgrp : Ejecuta el comando RMGRP
func commandRmgrp(rmgrp *RMGRP) error {
	// Verificar si hay una sesión activa y si el usuario es root
	if !globals.IsLoggedIn() {
		return fmt.Errorf("no hay ninguna sesión activa")
	}
	if globals.UsuarioActual.Name != "root" {
		return fmt.Errorf("solo el usuario root puede ejecutar este comando")
	}

	// Verificar que la partición está montada
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
	inodeOffset := int64(sb.S_inode_start + int32(binary.Size(usersInode)))
	err = usersInode.Decode(file, inodeOffset)
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

	fmt.Printf("Grupo '%s' eliminado exitosamente\n", rmgrp.Name)
	return nil
}
