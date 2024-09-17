package commands

import (
	structs "backend/Structs"
	globals "backend/globals"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// MKGRP : Estructura para el comando MKGRP
type MKGRP struct {
	Name string
}

// ParserMkgrp : Parseo de argumentos para el comando mkgrp y captura de los mensajes importantes
func ParserMkgrp(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer // Buffer para capturar los mensajes importantes para el usuario

	// Inicializar el comando MKGRP
	cmd := &MKGRP{}

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

	// Ejecutar la lógica del comando mkgrp
	err := commandMkgrp(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	// Retornar los mensajes importantes capturados en el buffer
	return outputBuffer.String(), nil
}

func commandMkgrp(mkgrp *MKGRP, outputBuffer *bytes.Buffer) error {
	fmt.Fprintln(outputBuffer, "======================= MKGRP =======================")
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

	// Cargar el Superblock utilizando el descriptor del archivo
	_, sb, _, err := globals.GetMountedPartitionRep(globals.UsuarioActual.Id)
	if err != nil {
		return fmt.Errorf("no se pudo cargar el Superblock: %v", err)
	}

	// Leer el inodo de users.txt
	var usersInode structs.Inode
	// Calcular el offset del inodo de users.txt, esta en el inodo 1
	inodeOffset := int64(sb.S_inode_start + int32(binary.Size(usersInode))) //ubuacion de los bloques de users.txt
	// Decodificar el inodo de users.txt
	err = usersInode.Decode(file, inodeOffset)
	usersInode.UpdateAtime() // Actualizar la última fecha de acceso
	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

	// Verificar si el grupo ya existe en users.txt
	_, err = globals.FindInUsersFile(file, sb, &usersInode, mkgrp.Name, "G")
	if err == nil {
		return fmt.Errorf("el grupo '%s' ya existe", mkgrp.Name)
	}

	// Obtener el siguiente ID disponible para el nuevo grupo
	nextGroupID, err := calculateNextID(file, sb, &usersInode)
	if err != nil {
		return fmt.Errorf("error calculando el siguiente ID: %v", err)
	}

	// Crear la nueva entrada de grupo con el siguiente ID
	newGroupEntry := fmt.Sprintf("%d,G,%s", nextGroupID, mkgrp.Name)

	// Usar la función modular para crear el grupo en users.txt
	err = globals.AddEntryToUsersFile(file, sb, &usersInode, newGroupEntry, mkgrp.Name, "G")
	if err != nil {
		return fmt.Errorf("error creando el grupo '%s': %v", mkgrp.Name, err)
	}

	// Actualizar el inodo de users.txt
	err = usersInode.Encode(file, inodeOffset)
	usersInode.UpdateAtime()
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	fmt.Fprintf(outputBuffer, "Grupo creado exitosamente: %s\n", mkgrp.Name)

	fmt.Fprintf(outputBuffer, "===========================================================")
	return nil
}

// calculateNextID : Calcula el siguiente ID disponible para un grupo o usuario en users.txt
func calculateNextID(file *os.File, sb *structs.Superblock, inode *structs.Inode) (int, error) {
	// Leer el contenido de users.txt
	contenido, err := globals.ReadFileBlocks(file, sb, inode)
	if err != nil {
		return -1, fmt.Errorf("error leyendo el contenido de users.txt: %v", err)
	}

	// Buscar el mayor ID en el archivo
	lineas := strings.Split(contenido, "\n")
	maxID := 0
	for _, linea := range lineas {
		if linea == "" {
			continue
		}

		campos := strings.Split(linea, ",")
		if len(campos) < 3 {
			continue // Ignorar líneas mal formadas
		}

		// Convertir el primer campo (ID) a entero
		id, err := strconv.Atoi(campos[0])
		if err != nil {
			continue // Ignorar IDs mal formados
		}

		// Actualizar el maxID si encontramos uno mayor
		if id > maxID {
			maxID = id
		}
	}

	// Devolver el siguiente ID disponible
	return maxID + 1, nil
}
