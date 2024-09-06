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

type LOGIN struct {
	User string
	Pass string
	ID   string
}

// ParserLogin analiza los tokens y crea una instancia del comando LOGIN
func ParserLogin(tokens []string) (*LOGIN, error) {
	cmd := &LOGIN{}
	args := strings.Join(tokens, " ")

	// Expresión regular para encontrar los parámetros del comando login
	re := regexp.MustCompile(`-user=[^\s]+|-pass=[^\s]+|-id=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		switch key {
		case "-user":
			cmd.User = value
		case "-pass":
			cmd.Pass = value
		case "-id":
			cmd.ID = value
		default:
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Validar que se hayan proporcionado todos los parámetros
	if cmd.User == "" || cmd.Pass == "" || cmd.ID == "" {
		return nil, fmt.Errorf("faltan parámetros requeridos: -user, -pass, -id")
	}

	// Ejecutar el comando login
	err := commandLogin(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return cmd, nil
}

// Lógica para ejecutar el login
func commandLogin(login *LOGIN) error {
	fmt.Println("===== INICIO DE LOGIN =====")
	fmt.Printf("Login ID: %s, User: %s\n", login.ID, login.User)

	// 1. Validar si ya hay una sesión activa
	if globals.UsuarioActual != nil && globals.UsuarioActual.Status {
		return fmt.Errorf("ya hay un usuario logueado, debe cerrar sesión primero")
	}

	// 2. Verificar que la partición esté montada
	_, path, err := globals.GetMountedPartition(login.ID)
	if err != nil {
		return fmt.Errorf("no se puede encontrar la partición: %v", err)
	}
	fmt.Printf("Partición montada en: %s\n", path)

	// 3. Cargar el Superblock de la partición montada
	_, sb, _, err := globals.GetMountedPartitionRep(login.ID)
	if err != nil {
		return fmt.Errorf("no se pudo cargar el Superblock: %v", err)
	}
	fmt.Println("Superblock cargado correctamente")

	// 4. Acceder al inodo del archivo users.txt (inodo 1)
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("no se puede abrir el archivo de partición: %v", err)
	}
	defer file.Close()

	// Leer el inodo 1 (que contiene el archivo users.txt)
	var usersInode structs.Inode
	inodeOffset := int64(sb.S_inode_start + int32(binary.Size(usersInode)))
	fmt.Printf("Leyendo inodo users.txt en la posición: %d\n", inodeOffset)

	err = usersInode.Decode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error leyendo inodo de users.txt: %v", err)
	}
	fmt.Println("Inodo de users.txt leído correctamente")
	usersInode.Print()

	// 5. Leer el contenido de los bloques asociados al archivo users.txt
	var contenido string
	for _, blockIndex := range usersInode.I_block {
		if blockIndex == -1 {
			// Si el bloque no está asignado, lo ignoramos
			continue
		}

		blockOffset := int64(sb.S_block_start + blockIndex*int32(binary.Size(structs.FileBlock{})))
		fmt.Printf("Leyendo bloque en la posición: %d (índice de bloque: %d)\n", blockOffset, blockIndex)

		var fileBlock structs.FileBlock
		err = fileBlock.Decode(file, blockOffset)
		if err != nil {
			return fmt.Errorf("error leyendo bloque de users.txt: %v", err)
		}

		contenido += string(fileBlock.B_content[:])
		fmt.Printf("Contenido del bloque leído: %s\n", string(fileBlock.B_content[:]))
	}

	fmt.Println("Contenido total de users.txt:")
	fmt.Println(contenido)

	// 6. Validar el usuario y contraseña
	encontrado := false
	lineas := strings.Split(strings.TrimSpace(contenido), "\n")
	for _, linea := range lineas {
		if linea == "" {
			continue
		}

		datos := strings.Split(linea, ",")
		if len(datos) == 5 && datos[1] == "U" {
			// Crear un objeto User a partir de la línea
			usuario := structs.NewUser(datos[0], datos[2], datos[3], datos[4])

			// Comparar usuario y contraseña
			if usuario.Name == login.User && usuario.Password == login.Pass {
				encontrado = true
				globals.UsuarioActual = usuario
				globals.UsuarioActual.Status = true
				fmt.Printf("Bienvenido %s, inicio de sesión exitoso.\n", usuario.Name)
				//guardar el ID de la particion montada
				globals.UsuarioActual.Id = login.ID
				break
			}
		}
	}

	if !encontrado {
		return fmt.Errorf("usuario o contraseña incorrectos")
	}

	fmt.Println("===== FIN DE LOGIN =====")
	return nil
}
