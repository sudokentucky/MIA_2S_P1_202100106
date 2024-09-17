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

// RMUSR : Estructura para el comando RMUSR
type RMUSR struct {
	User string
}

// ParserRmusr : Parseo de argumentos para el comando rmusr y captura de mensajes importantes
func ParserRmusr(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer // Buffer para capturar los mensajes importantes para el usuario

	// Inicializar el comando RMUSR
	cmd := &RMUSR{}

	// Expresión regular para encontrar el parámetro -user
	re := regexp.MustCompile(`-usr=[^\s]+`)
	matches := re.FindString(strings.Join(tokens, " "))

	if matches == "" {
		return "", fmt.Errorf("falta el parámetro -usr")
	}

	// Extraer el valor del parámetro -user
	param := strings.SplitN(matches, "=", 2)
	if len(param) != 2 {
		return "", fmt.Errorf("formato incorrecto para -user")
	}
	cmd.User = param[1]

	// Ejecutar la lógica del comando rmusr
	err := commandRmusr(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	// Retornar los mensajes importantes capturados en el buffer
	return outputBuffer.String(), nil
}

// commandRmusr : Ejecuta el comando RMUSR y captura los mensajes importantes en un buffer
func commandRmusr(rmusr *RMUSR, outputBuffer *bytes.Buffer) error {
	fmt.Fprintln(outputBuffer, "======================= RMUSR =======================")
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

	// Verificar si el usuario existe en el archivo users.txt
	_, err = globals.FindInUsersFile(file, sb, &usersInode, rmusr.User, "U")
	if err != nil {
		return fmt.Errorf("el usuario '%s' no existe", rmusr.User)
	}

	// Marcar el usuario como eliminado
	err = UpdateUserState(file, sb, &usersInode, rmusr.User)
	if err != nil {
		return fmt.Errorf("error eliminando el usuario '%s': %v", rmusr.User, err)
	}

	// Actualizar el inodo de users.txt
	err = usersInode.Encode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	// Mensaje de éxito
	fmt.Fprintf(outputBuffer, "Usuario '%s' eliminado exitosamente.\n", rmusr.User)
	fmt.Println("\nBloques:")
	sb.PrintBlocks(file.Name())
	fmt.Println("\nInodos:")
	sb.PrintInodes(file.Name())
	fmt.Fprintf(outputBuffer, "===========================================================")
	return nil
}

// UpdateUserState : Cambia el estado de un usuario a eliminado (ID=0) y actualiza el archivo
func UpdateUserState(file *os.File, sb *structs.Superblock, usersInode *structs.Inode, userName string) error {
	// Leer el contenido actual de users.txt
	contenido, err := globals.ReadFileBlocks(file, sb, usersInode)
	if err != nil {
		return fmt.Errorf("error leyendo el contenido de users.txt: %v", err)
	}

	// Separar las líneas del archivo
	lineas := strings.Split(contenido, "\n")
	modificado := false

	// Recorrer las líneas para buscar y modificar el estado del usuario
	for i, linea := range lineas {
		linea = strings.TrimSpace(linea) // Eliminar espacios en blanco adicionales
		if linea == "" {
			continue
		}

		// Crear un objeto User a partir de la línea
		usuario := crearUsuarioDesdeLinea(linea)

		// Verificar si es el usuario que queremos eliminar
		if usuario != nil && usuario.Name == userName {
			// Eliminar el usuario (cambiar ID a "0")
			usuario.Eliminar()

			// Actualizar la línea en el archivo
			lineas[i] = usuario.ToString()
			modificado = true
			break // Una vez que encontramos y modificamos el usuario, podemos salir
		}
	}

	// Si no se encontró el usuario, retornar error
	if !modificado {
		return fmt.Errorf("usuario '%s' no encontrado en users.txt", userName)
	}

	// Limpiar y actualizar las líneas antes de escribir
	contenidoActualizado := limpiarYActualizarContenido(lineas)

	// Escribir los cambios al archivo
	return escribirCambiosEnArchivo(file, sb, usersInode, contenidoActualizado)
}

// crearUsuarioDesdeLinea : Crea un objeto User a partir de una línea del archivo
func crearUsuarioDesdeLinea(linea string) *structs.User {
	partes := strings.Split(linea, ",")
	if len(partes) >= 5 && partes[1] == "U" {
		return structs.NewUser(partes[0], partes[2], partes[3], partes[4])
	}
	return nil
}

// limpiarYActualizarContenido : Elimina líneas vacías y devuelve el contenido actualizado como string
func limpiarYActualizarContenido(lineas []string) string {
	var contenidoActualizado []string
	for _, linea := range lineas {
		if strings.TrimSpace(linea) != "" {
			contenidoActualizado = append(contenidoActualizado, linea)
		}
	}
	return strings.Join(contenidoActualizado, "\n") + "\n"
}

// escribirCambiosEnArchivo : Limpia los bloques y escribe el contenido actualizado en el archivo
func escribirCambiosEnArchivo(file *os.File, sb *structs.Superblock, usersInode *structs.Inode, contenido string) error {
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
		err := fileBlock.Encode(file, blockOffset)
		if err != nil {
			return fmt.Errorf("error escribiendo bloque limpio %d: %w", blockIndex, err)
		}
	}

	// Reescribir todo el contenido en los bloques después de limpiar
	err := globals.WriteUsersBlocks(file, sb, usersInode, contenido)
	if err != nil {
		return fmt.Errorf("error guardando los cambios en users.txt: %v", err)
	}

	return nil
}
