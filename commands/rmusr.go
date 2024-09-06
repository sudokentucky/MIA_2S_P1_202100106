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

// RMUSR : Estructura para el comando RMUSR
type RMUSR struct {
	User string
}

// ParserRmusr : Parseo de argumentos para el comando rmusr
func ParserRmusr(tokens []string) (*RMUSR, error) {
	// Inicializar el comando RMUSR
	cmd := &RMUSR{}

	// Expresión regular para encontrar el parámetro -user
	re := regexp.MustCompile(`-user=[^\s]+`)
	matches := re.FindString(strings.Join(tokens, " "))

	if matches == "" {
		return nil, fmt.Errorf("falta el parámetro -user")
	}

	// Extraer el valor del parámetro -user
	param := strings.SplitN(matches, "=", 2)
	if len(param) != 2 {
		return nil, fmt.Errorf("formato incorrecto para -user")
	}
	cmd.User = param[1]

	// Ejecutar la lógica del comando rmusr
	err := commandRmusr(cmd)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

// commandRmusr : Ejecuta el comando RMUSR
func commandRmusr(rmusr *RMUSR) error {
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

	// Verificar si el usuario existe en el archivo users.txt
	_, err = globals.FindInUsersFile(file, sb, &usersInode, rmusr.User, "U")
	if err != nil {
		return fmt.Errorf("el usuario '%s' no existe", rmusr.User)
	}

	// Marcar el usuario como eliminado (ID a "0")
	err = globals.RemoveFromUsersFile(file, sb, &usersInode, rmusr.User, "U")
	if err != nil {
		return fmt.Errorf("error eliminando el usuario '%s': %v", rmusr.User, err)
	}

	fmt.Printf("Usuario '%s' eliminado exitosamente\n", rmusr.User)
	return nil
}
