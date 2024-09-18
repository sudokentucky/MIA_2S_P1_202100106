package commands

import (
	structures "backend/Structs"
	global "backend/globals"
	utils "backend/utils"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// MKFILE estructura que representa el comando mkfile con sus parámetros
type MKFILE struct {
	path string // Ruta del archivo
	r    bool   // Opción recursiva
	size int    // Tamaño del archivo
	cont string // Contenido del archivo
}

// ParserMkfile parsea el comando mkfile y devuelve una instancia de MKFILE
func ParserMkfile(tokens []string) (string, error) {
	cmd := &MKFILE{}              // Crea una nueva instancia de MKFILE
	var outputBuffer bytes.Buffer // Buffer para capturar mensajes importantes

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-r|-size=\d+|-cont="[^"]+"|-cont=[^\s]+`)
	matches := re.FindAllString(args, -1)

	if len(matches) != len(tokens) {
		for _, token := range tokens {
			if !re.MatchString(token) {
				return "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])
		var value string
		if len(kv) == 2 {
			value = kv[1]
		}

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-r":
			cmd.r = true // Habilitar la opción recursiva
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size < 0 {
				return "", errors.New("el tamaño debe ser un número entero no negativo")
			}
			cmd.size = size
		case "-cont":
			if value == "" {
				return "", errors.New("el contenido no puede estar vacío")
			}
			cmd.cont = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	if cmd.size == 0 {
		cmd.size = 0
	}

	if cmd.cont == "" {
		cmd.cont = ""
	}

	// Crear el archivo con los parámetros proporcionados
	err := commandMkfile(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	// Retorna el contenido del buffer con los mensajes importantes
	return outputBuffer.String(), nil
}

func commandMkfile(mkfile *MKFILE, outputBuffer *bytes.Buffer) error {
	// Verificar si hay un usuario logueado
	if !global.IsLoggedIn() {
		return fmt.Errorf("no hay un usuario logueado")
	}

	// Obtener el ID de la partición desde el usuario logueado
	idPartition := global.UsuarioActual.Id

	// Obtener la partición montada asociada al usuario logueado
	partitionSuperblock, mountedPartition, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Generar el contenido del archivo si no se proporcionó
	if mkfile.cont == "" {
		mkfile.cont = generateContent(mkfile.size)
	}

	// Abrir el archivo de partición para operar sobre él
	file, err := os.OpenFile(partitionPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de partición: %w", err)
	}
	defer file.Close() // Cerrar el archivo cuando ya no sea necesario

	// Capturar mensajes importantes en el buffer
	fmt.Fprintln(outputBuffer, "======================= MKFILE =======================")
	fmt.Fprintf(outputBuffer, "Creando archivo: %s\n", mkfile.path)

	// Obtener los directorios y el nombre del archivo
	dirPath, _ := GetDirectoryAndFile(mkfile.path)

	// Verificar solo la existencia del directorio (sin incluir el archivo)
	fmt.Fprintf(outputBuffer, "Verificando la existencia del directorio: %s\n", dirPath)
	exists, _, err := directoryExists(partitionSuperblock, file, 0, dirPath) // Usamos el inodo raíz (0) para empezar la búsqueda
	if err != nil {
		return fmt.Errorf("error al verificar directorio: %w", err)
	}

	// Si -r NO está habilitado y el directorio no existe, devolvemos un error
	if !mkfile.r && !exists {
		return fmt.Errorf("el directorio '%s' no existe y no se puede crear sin el parámetro -r", dirPath)
	}

	// Si -r está habilitado y el directorio no existe, creamos los directorios intermedios
	if mkfile.r && !exists {
		err = createDirectory(dirPath, mkfile.r, partitionSuperblock, file, mountedPartition)
		if err != nil {
			return fmt.Errorf("error al crear directorios intermedios: %w", err)
		}
	}

	// Crear el archivo usando el archivo de partición abierto
	err = createFile(mkfile.path, mkfile.size, mkfile.cont, partitionSuperblock, file, mountedPartition, outputBuffer)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %w", err)
	}

	fmt.Fprintf(outputBuffer, "Archivo %s creado exitosamente\n", mkfile.path)
	fmt.Fprintln(outputBuffer, "=====================================================")

	return nil
}

// generateContent genera una cadena de números del 0 al 9 hasta cumplir el tamaño ingresado
func generateContent(size int) string {
	content := ""
	for len(content) < size {
		content += "0123456789"
	}
	return content[:size] // Recorta la cadena al tamaño exacto
}

// createFile ahora usa el archivo de partición ya abierto
func createFile(filePath string, size int, content string, sb *structures.Superblock, file *os.File, mountedPartition *structures.Partition, outputBuffer *bytes.Buffer) error {
	fmt.Fprintf(outputBuffer, "Creando archivo en la ruta: %s\n", filePath)

	// Obtener los directorios padres y el destino
	parentDirs, destDir := utils.GetParentDirectories(filePath)
	// Obtener contenido por chunks
	chunks := utils.SplitStringIntoChunks(content)
	fmt.Fprintf(outputBuffer, "Contenido generado: %v\n", chunks)

	// Crear el archivo en el sistema de archivos
	err := sb.CreateFile(file, parentDirs, destDir, size, chunks)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %w", err)
	}

	// Serializar el superbloque
	err = sb.Encode(file, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	//Mostrar estructuras en consola de depuración
	fmt.Println("\nInodos:")
	sb.PrintInodes(file.Name())
	fmt.Println("\nBloques de datos:")
	sb.PrintBlocks(file.Name())

	return nil
}

// Función para obtener la carpeta (directorio) y el nombre del archivo
func GetDirectoryAndFile(path string) (string, string) {
	// Obtener la carpeta donde se creará el archivo
	dir := filepath.Dir(path)
	// Obtener el nombre del archivo
	file := filepath.Base(path)
	return dir, file
}
