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

// MKUSR : Estructura para el comando MKUSR
type MKUSR struct {
	User string
	Pass string
	Grp  string
}

func validateParamLength(param string, maxLength int, paramName string) error {
	if len(param) > maxLength {
		return fmt.Errorf("%s debe tener un máximo de %d caracteres", paramName, maxLength)
	}
	return nil
}

// ParserMkusr : Parseo de argumentos para el comando mkusr
func ParserMkusr(tokens []string) (*MKUSR, error) {
	// Inicializar el comando MKUSR
	cmd := &MKUSR{}

	// Expresión regular para encontrar los parámetros -user, -pass, -grp
	reUser := regexp.MustCompile(`-user=[^\s]+`)
	rePass := regexp.MustCompile(`-pass=[^\s]+`)
	reGrp := regexp.MustCompile(`-grp=[^\s]+`)

	// Buscar los parámetros
	matchesUser := reUser.FindString(strings.Join(tokens, " "))
	matchesPass := rePass.FindString(strings.Join(tokens, " "))
	matchesGrp := reGrp.FindString(strings.Join(tokens, " "))

	// Verificar que se proporcionen todos los parámetros
	if matchesUser == "" {
		return nil, fmt.Errorf("falta el parámetro -user")
	}
	if matchesPass == "" {
		return nil, fmt.Errorf("falta el parámetro -pass")
	}
	if matchesGrp == "" {
		return nil, fmt.Errorf("falta el parámetro -grp")
	}

	// Extraer los valores de los parámetros
	cmd.User = strings.SplitN(matchesUser, "=", 2)[1]
	cmd.Pass = strings.SplitN(matchesPass, "=", 2)[1]
	cmd.Grp = strings.SplitN(matchesGrp, "=", 2)[1]

	// Validar longitudes de los parámetros
	if err := validateParamLength(cmd.User, 10, "Usuario"); err != nil {
		return nil, err
	}
	if err := validateParamLength(cmd.Pass, 10, "Contraseña"); err != nil {
		return nil, err
	}
	if err := validateParamLength(cmd.Grp, 10, "Grupo"); err != nil {
		return nil, err
	}

	// Ejecutar la lógica del comando mkusr
	err := commandMkusr(cmd)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando mkusr: %v", err)
	}

	return cmd, nil
}

// commandMkusr : Ejecuta el comando MKUSR con depuración
func commandMkusr(mkusr *MKUSR) error {
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

	// Mostrar el Superblock antes de agregar el usuario
	fmt.Println("\nSuperblock antes de agregar el usuario:")
	sb.Print()

	// Leer el inodo de users.txt
	var usersInode structs.Inode
	inodeOffset := int64(sb.S_inode_start + int32(binary.Size(usersInode)))
	err = usersInode.Decode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

	// Mostrar el inodo de users.txt antes de agregar el usuario
	fmt.Println("\nInodo de users.txt antes de agregar usuario:")
	usersInode.Print()

	// Verificar si el grupo existe
	_, err = globals.FindInUsersFile(file, sb, &usersInode, mkusr.Grp, "G")
	if err != nil {
		return fmt.Errorf("el grupo '%s' no existe", mkusr.Grp)
	}

	// Verificar si el usuario ya existe
	_, err = globals.FindInUsersFile(file, sb, &usersInode, mkusr.User, "U")
	if err == nil {
		return fmt.Errorf("el usuario '%s' ya existe", mkusr.User)
	}

	// Obtener el siguiente ID disponible
	nextUserID, err := globals.GetNextID(file, sb, &usersInode)
	if err != nil {
		return fmt.Errorf("error obteniendo el siguiente ID para el usuario: %v", err)
	}

	// Crear una nueva entrada de usuario
	newUserEntry := fmt.Sprintf("%d,U,%s,%s,%s", nextUserID, mkusr.Grp, mkusr.User, mkusr.Pass)
	err = globals.AddToUsersFile(file, sb, &usersInode, newUserEntry)
	if err != nil {
		return fmt.Errorf("error agregando el usuario '%s' a users.txt: %v", mkusr.User, err)
	}

	// Actualizar el inodo de users.txt en el archivo
	err = usersInode.Encode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	// Mostrar el inodo de users.txt después de agregar el usuario
	fmt.Println("\nInodo de users.txt después de agregar usuario:")
	usersInode.Print()

	// Mostrar el contenido de users.txt después de agregar el usuario
	contenido, err := globals.ReadFileBlocks(file, sb, &usersInode)
	if err != nil {
		return fmt.Errorf("error leyendo el contenido de users.txt: %v", err)
	}
	fmt.Println("\nContenido de users.txt después de agregar usuario:")
	fmt.Println(contenido)

	// Mostrar el Superblock después de agregar el usuario
	fmt.Println("\nSuperblock después de agregar el usuario:")
	sb.Print()

	// Mostrar bloques del sistema de archivos
	err = sb.PrintBlocks(path)
	if err != nil {
		return fmt.Errorf("error imprimiendo los bloques del sistema: %v", err)
	}

	// Mostrar inodos del sistema de archivos
	err = sb.PrintInodes(path)
	if err != nil {
		return fmt.Errorf("error imprimiendo los inodos del sistema: %v", err)
	}

	fmt.Printf("\nUsuario creado exitosamente: %s\n", mkusr.User)
	return nil
}
