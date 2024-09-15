package commands

import (
	structs "backend/Structs"
	globals "backend/globals"
	"bytes"
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

// validateParamLength : Valida que los parámetros no excedan una longitud máxima
func validateParamLength(param string, maxLength int, paramName string) error {
	if len(param) > maxLength {
		return fmt.Errorf("%s debe tener un máximo de %d caracteres", paramName, maxLength)
	}
	return nil
}

// ParserMkusr : Parseo de argumentos para el comando mkusr y captura de los mensajes importantes
func ParserMkusr(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer // Buffer para capturar los mensajes importantes para el usuario

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
		return "", fmt.Errorf("falta el parámetro -user")
	}
	if matchesPass == "" {
		return "", fmt.Errorf("falta el parámetro -pass")
	}
	if matchesGrp == "" {
		return "", fmt.Errorf("falta el parámetro -grp")
	}

	// Extraer los valores de los parámetros
	cmd.User = strings.SplitN(matchesUser, "=", 2)[1]
	cmd.Pass = strings.SplitN(matchesPass, "=", 2)[1]
	cmd.Grp = strings.SplitN(matchesGrp, "=", 2)[1]

	// Validar longitudes de los parámetros
	if err := validateParamLength(cmd.User, 10, "Usuario"); err != nil {
		return "", err
	}
	if err := validateParamLength(cmd.Pass, 10, "Contraseña"); err != nil {
		return "", err
	}
	if err := validateParamLength(cmd.Grp, 10, "Grupo"); err != nil {
		return "", err
	}

	// Ejecutar la lógica del comando mkusr
	err := commandMkusr(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	// Retornar los mensajes importantes capturados en el buffer
	return outputBuffer.String(), nil
}

// commandMkusr : Ejecuta el comando MKUSR con captura de mensajes importantes en el buffer
func commandMkusr(mkusr *MKUSR, outputBuffer *bytes.Buffer) error {
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
	inodeOffset := int64(sb.S_inode_start)
	err = usersInode.Decode(file, inodeOffset) // Usar el descriptor de archivo
	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

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
	nextUserID, err := calculateNextID(file, sb, &usersInode)
	if err != nil {
		return fmt.Errorf("error calculando el siguiente ID para el usuario: %v", err)
	}

	// Crear una nueva entrada de usuario
	newUserEntry := fmt.Sprintf("%d,U,%s,%s,%s", nextUserID, mkusr.Grp, mkusr.User, mkusr.Pass)

	// Agregar la entrada al archivo users.txt
	err = globals.AddOrUpdateInUsersFile(file, sb, &usersInode, newUserEntry, mkusr.User, "U")
	if err != nil {
		return fmt.Errorf("error agregando el usuario '%s' a users.txt: %v", mkusr.User, err)
	}

	// Actualizar el inodo de users.txt
	err = usersInode.Encode(file, inodeOffset) // Usar el descriptor de archivo
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	// Mostrar mensaje de éxito importante
	fmt.Fprintf(outputBuffer, "Usuario '%s' agregado exitosamente al grupo '%s'\n", mkusr.User, mkusr.Grp)

	return nil
}
