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

// MKGRP : Estructura para el comando MKGRP
type MKGRP struct {
	Name string
}

// ParserMkgrp : Parseo de argumentos para el comando mkgrp
func ParserMkgrp(tokens []string) (*MKGRP, error) {
	// Inicializar el comando MKGRP
	cmd := &MKGRP{}

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

	// Ejecutar la lógica del comando mkgrp
	err := commandMkgrp(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}
	return cmd, nil
}

// commandMkgrp : Ejecuta el comando MKGRP con depuración
func commandMkgrp(mkgrp *MKGRP) error {
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

	// Mostrar el estado actual del Superblock antes de agregar el grupo
	fmt.Println("Superblock antes de agregar el grupo:")
	sb.Print()

	// Leer el inodo de users.txt
	var usersInode structs.Inode
	inodeOffset := int64(sb.S_inode_start + int32(binary.Size(usersInode)))
	err = usersInode.Decode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

	// Mostrar el estado del inodo de users.txt antes de la modificación
	fmt.Println("Inodo de users.txt antes de agregar grupo:")
	usersInode.Print()

	// Mostrar el contenido de users.txt antes de agregar el grupo
	contenidoAntes, err := globals.ReadFileBlocks(file, sb, &usersInode)
	if err != nil {
		return fmt.Errorf("error leyendo contenido de users.txt antes de agregar grupo: %v", err)
	}
	fmt.Println("Contenido de users.txt antes de agregar grupo:")
	fmt.Println(contenidoAntes)

	// Usar FindInUsersFile para verificar si el grupo ya existe
	_, err = globals.FindInUsersFile(file, sb, &usersInode, mkgrp.Name, "G")
	if err == nil {
		return fmt.Errorf("el grupo '%s' ya existe", mkgrp.Name)
	}

	// Obtener el siguiente ID disponible
	nextGroupID, err := globals.GetNextID(file, sb, &usersInode)
	if err != nil {
		return fmt.Errorf("error obteniendo el siguiente ID para el grupo: %v", err)
	}

	// Agregar el nuevo grupo al archivo users.txt
	newGroupEntry := fmt.Sprintf("%d,G,%s", nextGroupID, mkgrp.Name)
	err = globals.AddToUsersFile(file, sb, &usersInode, newGroupEntry)
	if err != nil {
		return fmt.Errorf("error agregando el grupo '%s' a users.txt: %v", mkgrp.Name, err)
	}

	// Mostrar el contenido de users.txt después de agregar el grupo
	contenidoDespues, err := globals.ReadFileBlocks(file, sb, &usersInode)
	if err != nil {
		return fmt.Errorf("error leyendo contenido de users.txt después de agregar grupo: %v", err)
	}
	fmt.Println("Contenido de users.txt después de agregar grupo:")
	fmt.Println(contenidoDespues)

	// Actualizar el inodo de users.txt
	err = usersInode.Encode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	// Mostrar el estado del inodo de users.txt después de la modificación
	fmt.Println("Inodo de users.txt después de agregar grupo:")
	usersInode.Print()

	// Mostrar el Superblock después de agregar el grupo
	fmt.Println("Superblock después de agregar el grupo:")
	sb.Print()

	// Mostrar los bloques que contiene el archivo
	err = sb.PrintBlocks(path)
	if err != nil {
		return fmt.Errorf("error imprimiendo los bloques: %v", err)
	}

	// Mostrar los inodos después de la modificación
	err = sb.PrintInodes(path)
	if err != nil {
		return fmt.Errorf("error imprimiendo los inodos: %v", err)
	}

	fmt.Println("Grupo creado exitosamente:", mkgrp.Name)
	return nil
}
