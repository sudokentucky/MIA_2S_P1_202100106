package reps

import (
	structs "backend/Structs"
	"backend/utils"
	"fmt"
	"os"
	"path/filepath"
)

// ReportFile genera un reporte que contiene el nombre y el contenido de un archivo específico
func ReportFile(superblock *structs.Superblock, diskPath string, path string, filePath string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	// Abrir el archivo de disco
	file, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}
	defer file.Close()

	// Buscar el inodo del archivo especificado en filePath
	inodeIndex, err := findFileInode(superblock, file, filePath)
	if err != nil {
		return fmt.Errorf("error al buscar el inodo del archivo: %v", err)
	}

	// Leer el contenido del archivo
	fileContent, err := readFileContent(superblock, file, inodeIndex)
	if err != nil {
		return fmt.Errorf("error al leer el contenido del archivo: %v", err)
	}

	// Crear el archivo de salida para el reporte
	reportFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error al crear el archivo de reporte: %v", err)
	}
	defer reportFile.Close()

	// Escribir el nombre y el contenido del archivo en el archivo de reporte
	_, fileName := filepath.Split(filePath)
	reportContent := fmt.Sprintf("Nombre del archivo: %s\n\nContenido del archivo:\n%s", fileName, fileContent)

	_, err = reportFile.WriteString(reportContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo de reporte: %v", err)
	}

	fmt.Println("Reporte del archivo generado:", path)
	return nil
}

// findFileInode busca el inodo del archivo especificado a través de su ruta
func findFileInode(superblock *structs.Superblock, diskFile *os.File, filePath string) (int32, error) {
	// Asumimos que partimos del inodo raíz
	currentInodeIndex := int32(0) // Inodo raíz

	// Dividimos el path en directorios y nombre de archivo
	directories, fileName := utils.GetParentDirectories(filePath)

	// Navegar por cada directorio para encontrar el inodo final
	for _, dir := range directories {
		inode, err := readInode(superblock, diskFile, currentInodeIndex)
		if err != nil {
			return -1, fmt.Errorf("error al leer el inodo: %v", err)
		}

		// Buscar el directorio en el bloque del inodo actual
		found, nextInodeIndex := findInodeInDirectory(inode, diskFile, dir, superblock)
		if !found {
			return -1, fmt.Errorf("directorio '%s' no encontrado", dir)
		}

		currentInodeIndex = nextInodeIndex
	}

	// Ahora buscar el archivo en el último directorio
	inode, err := readInode(superblock, diskFile, currentInodeIndex)
	if err != nil {
		return -1, fmt.Errorf("error al leer el inodo del directorio final: %v", err)
	}

	// Buscar el archivo en el bloque del inodo actual
	found, fileInodeIndex := findInodeInDirectory(inode, diskFile, fileName, superblock)
	if !found {
		return -1, fmt.Errorf("archivo '%s' no encontrado", fileName)
	}

	return fileInodeIndex, nil
}

// readFileContent lee el contenido de un archivo dado su inodo
func readFileContent(superblock *structs.Superblock, diskFile *os.File, inodeIndex int32) (string, error) {
	inode, err := readInode(superblock, diskFile, inodeIndex)
	if err != nil {
		return "", fmt.Errorf("error al leer el inodo del archivo: %v", err)
	}

	// Concatenar el contenido de los bloques
	var content string
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}

		// Leer el bloque de archivo
		block, err := readFileBlock(superblock, diskFile, blockIndex)
		if err != nil {
			return "", fmt.Errorf("error al leer el bloque de archivo: %v", err)
		}

		content += string(block.B_content[:])
	}

	return content, nil
}

// readInode lee el inodo en la posición dada
func readInode(superblock *structs.Superblock, diskFile *os.File, inodeIndex int32) (*structs.Inode, error) {
	inode := &structs.Inode{}
	offset := int64(superblock.S_inode_start + inodeIndex*superblock.S_inode_size)
	err := inode.Decode(diskFile, offset)
	if err != nil {
		return nil, fmt.Errorf("error al decodificar el inodo: %v", err)
	}
	return inode, nil
}

// readFileBlock lee un bloque de archivo en la posición dada
func readFileBlock(superblock *structs.Superblock, diskFile *os.File, blockIndex int32) (*structs.FileBlock, error) {
	block := &structs.FileBlock{}
	offset := int64(superblock.S_block_start + blockIndex*superblock.S_block_size)
	err := block.Decode(diskFile, offset)
	if err != nil {
		return nil, fmt.Errorf("error al decodificar el bloque de archivo: %v", err)
	}
	return block, nil
}

// findInodeInDirectory busca un inodo dentro de un bloque de directorio dado
func findInodeInDirectory(inode *structs.Inode, diskFile *os.File, name string, superblock *structs.Superblock) (bool, int32) {
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}

		// Leer el bloque de carpeta
		block := &structs.FolderBlock{}
		offset := int64(superblock.S_block_start + blockIndex*superblock.S_block_size)
		err := block.Decode(diskFile, offset)
		if err != nil {
			continue
		}

		// Buscar el nombre dentro del bloque de carpeta
		for _, content := range block.B_content {
			if string(content.B_name[:]) == name {
				return true, content.B_inodo
			}
		}
	}
	return false, -1
}
