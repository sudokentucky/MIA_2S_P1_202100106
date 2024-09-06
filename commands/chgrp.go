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

// CHGRP : Estructura para el comando CHGRP
type CHGRP struct {
	User string
	Grp  string
}

// ParserChgrp : Parseo de argumentos para el comando chgrp
func ParserChgrp(tokens []string) (*CHGRP, error) {
	// Inicializar el comando CHGRP
	cmd := &CHGRP{}

	// Expresión regular para encontrar los parámetros -user y -grp
	reUser := regexp.MustCompile(`-user=[^\s]+`)
	reGrp := regexp.MustCompile(`-grp=[^\s]+`)

	// Buscar los parámetros
	matchesUser := reUser.FindString(strings.Join(tokens, " "))
	matchesGrp := reGrp.FindString(strings.Join(tokens, " "))

	if matchesUser == "" {
		return nil, fmt.Errorf("falta el parámetro -user")
	}
	if matchesGrp == "" {
		return nil, fmt.Errorf("falta el parámetro -grp")
	}

	// Extraer los valores de los parámetros
	cmd.User = strings.SplitN(matchesUser, "=", 2)[1]
	cmd.Grp = strings.SplitN(matchesGrp, "=", 2)[1]

	// Ejecutar la lógica del comando chgrp
	err := commandChgrp(cmd)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

// commandChgrp : Ejecuta el comando CHGRP
func commandChgrp(chgrp *CHGRP) error {
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
	userLine, err := globals.FindInUsersFile(file, sb, &usersInode, chgrp.User, "U")
	if err != nil {
		return fmt.Errorf("el usuario '%s' no existe", chgrp.User)
	}

	// Verificar si el grupo existe en el archivo users.txt
	_, err = globals.FindInUsersFile(file, sb, &usersInode, chgrp.Grp, "G")
	if err != nil {
		return fmt.Errorf("el grupo '%s' no existe o está eliminado", chgrp.Grp)
	}

	// Actualizar el grupo del usuario en users.txt
	updatedUserLine := updateGroupInLine(userLine, chgrp.Grp)
	err = globals.UpdateLineInUsersFile(file, sb, &usersInode, updatedUserLine, chgrp.User, "U")
	if err != nil {
		return fmt.Errorf("error actualizando el grupo del usuario '%s': %v", chgrp.User, err)
	}

	fmt.Printf("El grupo del usuario '%s' ha sido cambiado exitosamente a '%s'\n", chgrp.User, chgrp.Grp)
	return nil
}

// updateGroupInLine : Actualiza el grupo en la línea del usuario en el archivo users.txt
func updateGroupInLine(userLine, newGroup string) string {
	fields := strings.Split(userLine, ",")
	if len(fields) > 3 {
		fields[2] = newGroup // Cambia el grupo en el campo correspondiente
	}
	return strings.Join(fields, ",")
}
