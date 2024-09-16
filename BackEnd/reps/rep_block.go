package reps

import (
	structs "backend/Structs"
	"backend/utils"
	"fmt"
	"os"
	"os/exec"
)

// ReportBlock genera un reporte de los bloques utilizados en el sistema de archivos
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
	dotContent := initDotGraphForBlocks()

	// Si no hay bloques, devolver un error
	if superblock.S_blocks_count == 0 {
		return fmt.Errorf("no hay bloques en el sistema")
	}

	// Generar los bloques y sus conexiones
	dotContent, err = generateBlockGraph(dotContent, superblock, file)
	if err != nil {
		return err
	}

	dotContent += "}" // Fin del Dot

	// Crear el archivo DOT
	err = writeDotFile(dotFileName, dotContent)
	if err != nil {
		return err
	}

	// Ejecutar Graphviz para generar la imagen
	err = generateBlockImage(dotFileName, outputImage)
	if err != nil {
		return err
	}

	fmt.Println("Imagen de los bloques generada:", outputImage)
	return nil
}

// initDotGraphForBlocks inicializa el contenido básico del archivo DOT para los bloques
func initDotGraphForBlocks() string {
	return `digraph G {
		fontname="Helvetica,Arial,sans-serif"
		node [fontname="Helvetica,Arial,sans-serif", shape=plain, fontsize=12];
		edge [fontname="Helvetica,Arial,sans-serif", color="#FF7043", arrowsize=0.8];
		rankdir=LR;
		bgcolor="#FAFAFA";
		node [shape=plaintext];
		blockHeaderColor="#FF9800"; 
		cellBackgroundColor="#FFFDE7";
		cellBorderColor="#EEEEEE";
		textColor="#263238";
	`
}

// generateBlockGraph genera el contenido del grafo de bloques en formato DOT
func generateBlockGraph(dotContent string, superblock *structs.Superblock, file *os.File) (string, error) {
	// Iterar sobre cada bloque
	for i := int32(0); i < superblock.S_blocks_count; i++ {
		block := &structs.FolderBlock{}
		offset := int64(superblock.S_block_start + (i * superblock.S_block_size))

		// Intentar decodificar como bloque de carpeta
		err := block.Decode(file, offset)
		if err == nil {
			// Es un bloque de carpeta
			dotContent += generateFolderBlockTable(i, block)
			continue
		}

		// Intentar decodificar como bloque de archivo
		fileBlock := &structs.FileBlock{}
		err = fileBlock.Decode(file, offset)
		if err == nil {
			// Es un bloque de archivo
			dotContent += generateFileBlockTable(i, fileBlock)
			continue
		}

		// Intentar decodificar como bloque de apuntadores (indirectos)
		// Se puede implementar lógica para bloques indirectos aquí
	}

	return dotContent, nil
}

// generateFolderBlockTable genera una tabla en DOT para los bloques de carpeta
func generateFolderBlockTable(blockIndex int32, block *structs.FolderBlock) string {
	table := fmt.Sprintf(`block%d [label=<
		<table border="0" cellborder="1" cellspacing="0" cellpadding="4" bgcolor="#FFFDE7" style="rounded">
			<tr><td colspan="2" bgcolor="#FF9800" align="center"><b>BLOQUE CARPETA %d</b></td></tr>
	`, blockIndex, blockIndex)

	// Añadir contenido de la carpeta (nombre e inodo)
	for i, content := range block.B_content {
		if content.B_inodo != -1 { // Bloque usado
			name := string(content.B_name[:])
			table += fmt.Sprintf("<tr><td><b>Contenido %d - Inodo %d</b></td><td>%s</td></tr>", i+1, content.B_inodo, name)
		}
	}

	table += "</table>>];"
	return table
}

// generateFileBlockTable genera una tabla en DOT para los bloques de archivo
func generateFileBlockTable(blockIndex int32, block *structs.FileBlock) string {
	content := string(block.B_content[:])
	table := fmt.Sprintf(`block%d [label=<
		<table border="0" cellborder="1" cellspacing="0" cellpadding="4" bgcolor="#FFFDE7" style="rounded">
			<tr><td colspan="2" bgcolor="#FF9800" align="center"><b>BLOQUE ARCHIVO %d</b></td></tr>
			<tr><td colspan="2">%s</td></tr>
		</table>>];
	`, blockIndex, blockIndex, content)

	return table
}

// generateBlockImage genera una imagen a partir del archivo DOT usando Graphviz
func generateBlockImage(dotFileName string, outputImage string) error {
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	return nil
}
