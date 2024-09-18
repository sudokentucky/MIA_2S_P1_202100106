package reps

import (
	structs "backend/Structs"
	"backend/utils"
	"fmt"
	"html"
	"os"
	"os/exec"
	"strings"
)

// ReportBlockConnections genera un reporte visual de bloques usando Graphviz
func ReportBlock(superblock *structs.Superblock, diskPath string, path string) error {
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

	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)

	// Inicio del Dot
	dotContent := initDotGraph()

	// Generar los bloques y sus conexiones
	dotContent, connections, err := generateBlockGraph(dotContent, superblock, file)
	if err != nil {
		return err
	}

	dotContent += connections // Agregar conexiones fuera de las definiciones de nodos
	dotContent += "}"         // Fin del Dot

	// Crear el archivo DOT
	err = writeDotFile(dotFileName, dotContent)
	if err != nil {
		return err
	}

	// Ejecutar Graphviz para generar la imagen
	err = GenerateblockImage(dotFileName, outputImage)
	if err != nil {
		return err
	}

	fmt.Println("Imagen de los bloques generada:", outputImage)
	return nil
}

// GenerateblockImage genera una imagen a partir del archivo DOT usando Graphviz
func GenerateblockImage(dotFileName string, outputImage string) error {
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr // Esto imprimirá los errores en la consola
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	return nil
}

// generateBlockGraph genera el contenido del grafo de bloques en formato DOT
func generateBlockGraph(dotContent string, superblock *structs.Superblock, file *os.File) (string, string, error) {
	visitedBlocks := make(map[int32]bool)
	var connections string

	for i := int32(0); i < superblock.S_inodes_count; i++ {
		inode := &structs.Inode{}
		err := inode.Decode(file, int64(superblock.S_inode_start+(i*superblock.S_inode_size)))
		if err != nil {
			return "", "", fmt.Errorf("error al deserializar el inodo %d: %v", i, err)
		}

		if inode.I_uid == -1 || inode.I_uid == 0 {
			continue
		}

		// Recorrer los bloques asociados al inodo
		for _, block := range inode.I_block {
			if block != -1 {
				if !visitedBlocks[block] {
					dotContent, connections, err = generateBlockLabel(dotContent, connections, block, inode, superblock, file, visitedBlocks)
					if err != nil {
						return "", "", err
					}
					visitedBlocks[block] = true
				}
			}
		}
	}
	return dotContent, connections, nil
}

func generateBlockLabel(dotContent, connections string, blockIndex int32, inode *structs.Inode, superblock *structs.Superblock, file *os.File, visitedBlocks map[int32]bool) (string, string, error) {
	blockOffset := int64(superblock.S_block_start + (blockIndex * superblock.S_block_size))

	if inode.I_type[0] == '0' { // Bloque de carpeta
		folderBlock := &structs.FolderBlock{}
		err := folderBlock.Decode(file, blockOffset)
		if err != nil {
			return "", "", fmt.Errorf("error al decodificar bloque de carpeta %d: %w", blockIndex, err)
		}

		// Generar la etiqueta del bloque de carpeta con bordes
		label := fmt.Sprintf("BLOQUE DE CARPETA %d", blockIndex)
		hasValidConnections := false // Verificar si el bloque tiene conexiones válidas

		for i, content := range folderBlock.B_content {
			name := cleanBlockName(content.B_name)

			// Usar html.EscapeString para evitar que caracteres especiales rompan el DOT
			name = html.EscapeString(name)

			// Evitar conexiones internas (.) y (..)
			if content.B_inodo != -1 && !(i == 0 || i == 1) {
				// Añadir conexiones a otros inodos
				label += fmt.Sprintf("\\nContenido %d: %s (Inodo %d)", i+1, name, content.B_inodo)
				// Conectar solo si es un bloque de archivo válido
				if content.B_inodo != blockIndex {
					connections += fmt.Sprintf("block%d -> block%d [color=\"#FF7043\"];\n", blockIndex, content.B_inodo)
				}
				hasValidConnections = true
			} else {
				if i > 1 { // Evitamos mostrar las referencias internas en la etiqueta
					label += fmt.Sprintf("\\nContenido %d: %s (Inodo no asignado)", i+1, name)
				}
			}
		}

		// Solo agregar al contenido DOT si el bloque tiene conexiones válidas o contenido significativo
		if hasValidConnections {
			dotContent += fmt.Sprintf("block%d [label=\"%s\", shape=box, style=filled, fillcolor=\"#FFFDE7\", color=\"#EEEEEE\"];\n", blockIndex, label)
		}

	} else if inode.I_type[0] == '1' { // Bloque de archivo
		fileBlock := &structs.FileBlock{}
		err := fileBlock.Decode(file, blockOffset)
		if err != nil {
			return "", "", fmt.Errorf("error al decodificar bloque de archivo %d: %w", blockIndex, err)
		}

		content := cleanBlockContent(fileBlock.GetContent())

		// Solo generar la tabla si hay contenido
		if len(strings.TrimSpace(content)) > 0 {
			label := fmt.Sprintf("BLOQUE DE ARCHIVO %d\\n%s", blockIndex, content)
			dotContent += fmt.Sprintf("block%d [label=\"%s\", shape=box, style=filled, fillcolor=\"#FFFDE7\", color=\"#EEEEEE\"];\n", blockIndex, label)

			// Conectar con el siguiente bloque de archivo si existe
			nextBlock := findNextValidBlock(inode, blockIndex)
			if nextBlock != -1 {
				connections += fmt.Sprintf("block%d -> block%d [color=\"#FF7043\"];\n", blockIndex, nextBlock)
			}
		}
	}

	// Agregar referencia al bloque padre si existe
	parentBlock := findParentBlock(inode, blockIndex)
	if parentBlock != -1 {
		connections += fmt.Sprintf("block%d -> block%d [color=\"#FF7043\"];\n", parentBlock, blockIndex)
	}

	return dotContent, connections, nil
}

// findParentBlock busca el bloque padre del bloque actual
func findParentBlock(inode *structs.Inode, currentBlock int32) int32 {
	for i := 0; i < len(inode.I_block); i++ {
		if inode.I_block[i] == currentBlock && i > 0 {
			return inode.I_block[i-1]
		}
	}
	return -1 // No hay bloque padre
}

// findNextValidBlock busca el siguiente bloque válido en el array I_block del inodo
func findNextValidBlock(inode *structs.Inode, currentBlock int32) int32 {
	for i := 0; i < len(inode.I_block); i++ {
		if inode.I_block[i] == currentBlock {
			for j := i + 1; j < len(inode.I_block); j++ {
				if inode.I_block[j] != -1 {
					return inode.I_block[j]
				}
			}
		}
	}
	return -1 // No hay más bloques válidos
}

// cleanBlockName limpia el nombre del bloque, eliminando los caracteres nulos
func cleanBlockName(nameArray [12]byte) string {
	return strings.TrimRight(string(nameArray[:]), "\x00")
}

// cleanBlockContent limpia el contenido del bloque
func cleanBlockContent(content string) string {
	return strings.ReplaceAll(content, "\n", "\\n")
}
