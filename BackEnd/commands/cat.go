package commands

import (
	structs "backend/Structs"
	global "backend/globals"
	utils "backend/utils"
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// CAT estructura que representa el comando CAT con sus parámetros
type CAT struct {
	files []string // Lista de archivos a leer
}

// ParserCat parsea el comando cat y devuelve una instancia de CAT
func ParserCat(tokens []string) (string, error) {
	cmd := &CAT{}                 // Crea una nueva instancia de CAT
	var outputBuffer bytes.Buffer // Buffer para capturar mensajes importantes

	// Expresión regular para capturar los archivos pasados como parámetros -file1, -file2, etc.
	re := regexp.MustCompile(`-file\d+=("[^"]+"|[^\s]+)`)
	matches := re.FindAllString(strings.Join(tokens, " "), -1)

	// Verificar si no se encontraron archivos
	if len(matches) == 0 {
		return "", errors.New("no se especificaron archivos para leer")
	}

	// Iterar sobre cada coincidencia y extraer los archivos
	for _, match := range matches {
		// Separar el parámetro en clave y valor (ej. "-file1=/home/user/a.txt")
		kv := strings.SplitN(match, "=", 2)
		if len(kv) == 2 {
			filePath := kv[1]
			// Si el valor está entre comillas, eliminarlas
			if strings.HasPrefix(filePath, "\"") && strings.HasSuffix(filePath, "\"") {
				filePath = strings.Trim(filePath, "\"")
			}
			cmd.files = append(cmd.files, filePath)
		}
	}

	// Ejecutar el comando CAT
	err := commandCat(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	return outputBuffer.String(), nil
}

func commandCat(cat *CAT, outputBuffer *bytes.Buffer) error {
	fmt.Fprint(outputBuffer, "======================= CAT =======================\n")
	// Verificar si hay un usuario logueado
	if !global.IsLoggedIn() {
		return fmt.Errorf("no hay un usuario logueado")
	}

	// Obtener el ID de la partición desde el usuario logueado
	idPartition := global.UsuarioActual.Id

	// Obtener la partición montada asociada al usuario logueado
	_, _, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Abrir el archivo de partición para operar sobre él
	file, err := os.OpenFile(partitionPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de partición: %w", err)
	}
	defer file.Close() // Cerrar el archivo cuando ya no sea necesario

	// Leer y mostrar el contenido de cada archivo
	for _, filePath := range cat.files {
		fmt.Fprintf(outputBuffer, "Leyendo archivo: %s\n", filePath)

		// Leer el contenido del archivo
		content, err := readFileContent(filePath)
		if err != nil {
			fmt.Fprintf(outputBuffer, "Error al leer el archivo %s: %v\n", filePath, err)
			continue
		}

		// Concatenar el contenido del archivo al buffer de salida
		outputBuffer.WriteString(content)
		outputBuffer.WriteString("\n") // Separar el contenido de los archivos con un salto de línea
		fmt.Fprint(outputBuffer, "===========================================\n")
	}

	return nil
}

// readFileContent busca el archivo en el sistema de archivos y lee su contenido
func readFileContent(filePath string) (string, error) {
	// Obtener el Superblock y la partición montada asociada
	idPartition := global.UsuarioActual.Id
	partitionSuperblock, _, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "", fmt.Errorf("error al obtener la partición montada: %v", err)
	}

	// Abrir el archivo de partición para leer
	file, err := os.OpenFile(partitionPath, os.O_RDONLY, 0666)
	if err != nil {
		return "", fmt.Errorf("error al abrir el archivo de partición: %v", err)
	}
	defer file.Close()

	// Convertir el path del archivo en un array de carpetas
	parentDirs, fileName := utils.GetParentDirectories(filePath)

	// Buscar el archivo en el sistema de archivos
	inodeIndex, err := findFileInode(file, partitionSuperblock, parentDirs, fileName)
	if err != nil {
		return "", fmt.Errorf("error al encontrar el archivo: %v", err)
	}

	// Leer el contenido del archivo
	content, err := readFileFromInode(file, partitionSuperblock, inodeIndex)
	if err != nil {
		return "", fmt.Errorf("error al leer el contenido del archivo: %v", err)
	}

	return content, nil
}

// directoryExists verifica si un directorio o archivo ya existe en el inodo dado
func directoryExists(sb *structs.Superblock, file *os.File, inodeIndex int32, dirName string) (bool, int32, error) {
	fmt.Printf("Verificando si el directorio o archivo '%s' existe en el inodo %d\n", dirName, inodeIndex) // Depuración

	// Deserializar el inodo correspondiente
	inode := &structs.Inode{}
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return false, -1, fmt.Errorf("error al deserializar inodo %d: %v", inodeIndex, err)
	}

	// Verificar si el inodo es de tipo carpeta (I_type == '0') para continuar
	if inode.I_type[0] != '0' {
		return false, -1, fmt.Errorf("el inodo %d no es una carpeta", inodeIndex)
	}

	// Iterar sobre los bloques del inodo para buscar el directorio o archivo
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break // Si no hay más bloques asignados, terminamos la búsqueda
		}

		// Deserializar el bloque de directorio
		block := &structs.FolderBlock{}
		err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return false, -1, fmt.Errorf("error al deserializar bloque %d: %v", blockIndex, err)
		}

		// Iterar sobre los contenidos del bloque para verificar si el nombre coincide
		for _, content := range block.B_content {
			contentName := strings.Trim(string(content.B_name[:]), "\x00 ") // Convertir el nombre y eliminar los caracteres nulos
			if strings.EqualFold(contentName, dirName) && content.B_inodo != -1 {
				fmt.Printf("Directorio o archivo '%s' encontrado en inodo %d\n", dirName, content.B_inodo) // Depuración
				return true, content.B_inodo, nil                                                          // Devolver true si el directorio/archivo fue encontrado
			}
		}
	}

	fmt.Printf("Directorio o archivo '%s' no encontrado en inodo %d\n", dirName, inodeIndex) // Depuración
	return false, -1, nil                                                                    // No se encontró el directorio/archivo
}

// findFileInode busca el inodo de un archivo dado el path
func findFileInode(file *os.File, sb *structs.Superblock, parentsDir []string, fileName string) (int32, error) {
	// Empezar buscando en el inodo raíz
	inodeIndex := int32(0)

	// Navegar por los directorios padres para llegar al archivo
	for len(parentsDir) > 0 {
		dirName := parentsDir[0]
		found, newInodeIndex, err := directoryExists(sb, file, inodeIndex, dirName)
		if err != nil {
			return -1, err
		}
		if !found {
			return -1, fmt.Errorf("directorio '%s' no encontrado", dirName)
		}
		inodeIndex = newInodeIndex
		parentsDir = parentsDir[1:]
	}

	// Buscar el archivo en el último directorio
	found, fileInodeIndex, err := directoryExists(sb, file, inodeIndex, fileName)
	if err != nil {
		return -1, err
	}
	if !found {
		return -1, fmt.Errorf("archivo '%s' no encontrado", fileName)
	}

	return fileInodeIndex, nil
}

// readFileFromInode lee el contenido de un archivo desde su inodo
func readFileFromInode(file *os.File, sb *structs.Superblock, inodeIndex int32) (string, error) {
	inode := &structs.Inode{}
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return "", fmt.Errorf("error al deserializar el inodo %d: %v", inodeIndex, err)
	}

	if inode.I_type[0] != '1' {
		return "", fmt.Errorf("el inodo %d no corresponde a un archivo", inodeIndex)
	}

	// Concatenar los bloques de contenido del archivo
	var contentBuilder strings.Builder
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		fileBlock := &structs.FileBlock{}
		err := fileBlock.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return "", fmt.Errorf("error al deserializar el bloque %d: %v", blockIndex, err)
		}

		contentBuilder.WriteString(string(fileBlock.B_content[:]))
	}

	return contentBuilder.String(), nil
}
