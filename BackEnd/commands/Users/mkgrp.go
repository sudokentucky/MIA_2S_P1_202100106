package commands

import (
	structs "backend/Structs"
	globals "backend/globals"
	utils "backend/utils"
	"bytes"
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
		fmt.Println("Error:", err) // Mensaje de depuración
		return "", err
	}

	// Retornar los mensajes importantes capturados en el buffer
	return outputBuffer.String(), nil
}

// commandMkgrp : Ejecuta el comando MKGRP con mensajes capturados en el buffer
func commandMkgrp(mkgrp *MKGRP, outputBuffer *bytes.Buffer) error {
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
	fmt.Fprintln(outputBuffer, "Superblock antes de agregar el grupo:")
	sb.Print() // Mensaje de depuración en consola

	// Leer el inodo de users.txt
	var usersInode structs.Inode
	inodeOffset := int64(sb.S_inode_start + int32(binary.Size(usersInode)))
	err = usersInode.Decode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

	// Mostrar el estado del inodo de users.txt antes de la modificación
	fmt.Fprintln(outputBuffer, "Inodo de users.txt antes de agregar grupo:")
	usersInode.Print() // Mensaje de depuración en consola

	// Mostrar el contenido de users.txt antes de agregar el grupo
	contenidoAntes, err := globals.ReadFileBlocks(file, sb, &usersInode)
	if err != nil {
		return fmt.Errorf("error leyendo contenido de users.txt antes de agregar grupo: %v", err)
	}
	fmt.Fprintln(outputBuffer, "Contenido de users.txt antes de agregar grupo:")
	fmt.Fprintln(outputBuffer, contenidoAntes)

	// Verificar si el grupo ya existe
	_, err = globals.FindInUsersFile(file, sb, &usersInode, mkgrp.Name, "G")
	if err == nil {
		return fmt.Errorf("el grupo '%s' ya existe", mkgrp.Name)
	}

	// Obtener el siguiente ID disponible
	nextGroupID, err := globals.GetNextID(file, sb, &usersInode)
	if err != nil {
		return fmt.Errorf("error obteniendo el siguiente ID para el grupo: %v", err)
	}

	// Crear la nueva entrada del grupo
	newGroupEntry := fmt.Sprintf("%d,G,%s\n", nextGroupID, mkgrp.Name)

	// Verificar si hay espacio disponible en el bloque actual
	var currentBlock structs.FileBlock
	offset := int64(sb.S_block_start) + int64(usersInode.I_block[0])*int64(binary.Size(currentBlock))

	err = utils.ReadFromFile(file, offset, &currentBlock)
	if err != nil {
		return fmt.Errorf("error leyendo el bloque actual de users.txt: %v", err)
	}

	// Buscar el primer byte nulo en el bloque actual
	posicionNulo := strings.IndexByte(string(currentBlock.B_content[:]), 0)

	if posicionNulo != -1 {
		// Hay espacio en el bloque actual
		libre := 64 - (posicionNulo + len(newGroupEntry)) // 64 es el tamaño de FileBlock.B_content
		if libre >= 0 {
			copy(currentBlock.B_content[posicionNulo:], []byte(newGroupEntry))
			// Escribir el bloque actualizado
			err = utils.WriteToFile(file, offset, &currentBlock)
			if err != nil {
				return fmt.Errorf("error escribiendo en users.txt: %v", err)
			}
		} else {
			// Si no hay suficiente espacio, crear un nuevo bloque
			err = asignarNuevoBloqueYEscribir(file, sb, &usersInode, newGroupEntry)
			if err != nil {
				return fmt.Errorf("error asignando nuevo bloque para users.txt: %v", err)
			}
		}
	} else {
		// Bloque lleno, asignar un nuevo bloque
		err = asignarNuevoBloqueYEscribir(file, sb, &usersInode, newGroupEntry)
		if err != nil {
			return fmt.Errorf("error asignando nuevo bloque para users.txt: %v", err)
		}
	}

	// Actualizar el inodo de users.txt
	err = usersInode.Encode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	// Mostrar el contenido de users.txt después de agregar el grupo
	contenidoDespues, err := globals.ReadFileBlocks(file, sb, &usersInode)
	if err != nil {
		return fmt.Errorf("error leyendo contenido de users.txt después de agregar grupo: %v", err)
	}
	fmt.Fprintln(outputBuffer, "Contenido de users.txt después de agregar grupo:")
	fmt.Fprintln(outputBuffer, contenidoDespues)

	// Mostrar el estado del inodo de users.txt después de la modificación
	fmt.Fprintln(outputBuffer, "Inodo de users.txt después de agregar grupo:")
	usersInode.Print() // Mensaje de depuración en consola

	// Mostrar el Superblock después de agregar el grupo
	fmt.Fprintln(outputBuffer, "Superblock después de agregar el grupo:")
	sb.Print() // Mensaje de depuración en consola

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

	fmt.Fprintln(outputBuffer, "Grupo creado exitosamente:", mkgrp.Name)
	return nil
}

// asignarNuevoBloqueYEscribir : Asigna un nuevo bloque y escribe el contenido en él
func asignarNuevoBloqueYEscribir(file *os.File, sb *structs.Superblock, usersInode *structs.Inode, newGroupEntry string) error {
	for i, block := range usersInode.I_block {
		if block == -1 {
			// Asignar un nuevo bloque
			usersInode.I_block[i] = sb.S_first_blo
			sb.S_free_blocks_count--
			sb.S_first_blo++

			// Escribir la nueva entrada en el nuevo bloque
			var newFileBlock structs.FileBlock
			copy(newFileBlock.B_content[:], []byte(newGroupEntry))

			offset := int64(sb.S_block_start) + int64(usersInode.I_block[i])*int64(binary.Size(newFileBlock))

			// Escribir el nuevo bloque
			err := utils.WriteToFile(file, offset, &newFileBlock)
			if err != nil {
				return fmt.Errorf("error escribiendo nuevo bloque en users.txt: %v", err)
			}

			// Actualizar el bitmap de bloques
			err = structs.UpdateBlockBitmap(file, sb, usersInode.I_block[i])
			if err != nil {
				return fmt.Errorf("error actualizando bitmap de bloques: %v", err)
			}

			// Escribir el superbloque actualizado
			err = sb.Encode(file, int64(sb.S_inode_start))
			if err != nil {
				return fmt.Errorf("error actualizando el superbloque: %v", err)
			}

			return nil
		}
	}
	return fmt.Errorf("no hay bloques disponibles")
}
