package reps

import (
	structs "backend/Structs"
	"backend/utils"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// ReportInode genera un reporte de los inodos y lo guarda en la ruta especificada
func ReportInode(superblock *structs.Superblock, diskPath string, path string) error {
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

	// Si no hay inodos, devolver un error
	if superblock.S_inodes_count == 0 {
		return fmt.Errorf("no hay inodos en el sistema")
	}

	// Generar los inodos y sus conexiones
	dotContent, err = generateInodeGraph(dotContent, superblock, file)
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
	err = generateInodeImage(dotFileName, outputImage)
	if err != nil {
		return err
	}

	fmt.Println("Imagen de los inodos generada:", outputImage)
	return nil
}

// initDotGraph inicializa el contenido básico del archivo DOT
func initDotGraph() string {
	return `digraph G {
		fontname="Helvetica,Arial,sans-serif"
		node [fontname="Helvetica,Arial,sans-serif", shape=plain, fontsize=12];
		edge [fontname="Helvetica,Arial,sans-serif", color="#FF7043", arrowsize=0.8];
		rankdir=LR;
		bgcolor="#FAFAFA";
		node [shape=plaintext];
		inodeHeaderColor="#4CAF50"; 
		blockHeaderColor="#FF9800"; 
		cellBackgroundColor="#FFFDE7";
		cellBorderColor="#EEEEEE";
		textColor="#263238";
	`
}

// generateInodeGraph genera el contenido del grafo de inodos en formato DOT
func generateInodeGraph(dotContent string, superblock *structs.Superblock, file *os.File) (string, error) {
	for i := int32(0); i < superblock.S_inodes_count; i++ {
		inode := &structs.Inode{}
		err := inode.Decode(file, int64(superblock.S_inode_start+(i*superblock.S_inode_size)))
		if err != nil {
			return "", fmt.Errorf("error al deserializar el inodo %d: %v", i, err)
		}

		// Verificar si el inodo está en uso
		if inode.I_uid == -1 || inode.I_uid == 0 {
			continue
		}

		// Generar la tabla del inodo
		dotContent += generateInodeTable(i, inode)

		// Conexión entre inodos
		if i < superblock.S_inodes_count-1 {
			dotContent += fmt.Sprintf("inode%d -> inode%d [color=\"#FF7043\"];\n", i, i+1)
		}
	}
	return dotContent, nil
}

// generateInodeTable genera la tabla con los atributos y bloques del inodo en formato DOT
func generateInodeTable(inodeIndex int32, inode *structs.Inode) string {
	// Convertir tiempos a string
	atime := time.Unix(int64(inode.I_atime), 0).Format(time.RFC3339)
	ctime := time.Unix(int64(inode.I_ctime), 0).Format(time.RFC3339)
	mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)

	// Generar la tabla del inodo
	table := fmt.Sprintf(`inode%d [label=<
		<table border="0" cellborder="1" cellspacing="0" cellpadding="4" bgcolor="#FFFDE7" style="rounded">
			<tr><td colspan="2" bgcolor="#4CAF50" align="center"><b>INODO %d</b></td></tr>
			<tr><td><b>i_uid</b></td><td>%d</td></tr>
			<tr><td><b>i_gid</b></td><td>%d</td></tr>
			<tr><td><b>i_size</b></td><td>%d</td></tr>
			<tr><td><b>i_atime</b></td><td>%s</td></tr>
			<tr><td><b>i_ctime</b></td><td>%s</td></tr>
			<tr><td><b>i_mtime</b></td><td>%s</td></tr>
			<tr><td><b>i_type</b></td><td>%c</td></tr>
			<tr><td><b>i_perm</b></td><td>%s</td></tr>
			<tr><td colspan="2" bgcolor="#FF9800"><b>BLOQUES DIRECTOS</b></td></tr>
	`, inodeIndex, inodeIndex, inode.I_uid, inode.I_gid, inode.I_size, atime, ctime, mtime, rune(inode.I_type[0]), string(inode.I_perm[:]))

	// Agregar los bloques directos
	for j, block := range inode.I_block[:12] {
		if block != -1 { // Bloques usados
			table += fmt.Sprintf("<tr><td><b>%d</b></td><td>%d</td></tr>", j+1, block)
		}
	}

	// Agregar bloques indirectos (si existen)
	table += generateIndirectBlocks(inode)

	table += "</table>>];"
	return table
}

// generateIndirectBlocks agrega los bloques indirectos al inodo
func generateIndirectBlocks(inode *structs.Inode) string {
	result := ""
	// Agregar bloque indirecto simple
	if inode.I_block[12] != -1 {
		result += fmt.Sprintf(`
			<tr><td colspan="2" bgcolor="#FF9800"><b>BLOQUE INDIRECTO SIMPLE</b></td></tr>
			<tr><td><b>13</b></td><td>%d</td></tr>
		`, inode.I_block[12])
	}

	// Agregar bloque indirecto doble
	if inode.I_block[13] != -1 {
		result += fmt.Sprintf(`
			<tr><td colspan="2" bgcolor="#FF9800"><b>BLOQUE INDIRECTO DOBLE</b></td></tr>
			<tr><td><b>14</b></td><td>%d</td></tr>
		`, inode.I_block[13])
	}

	// Agregar bloque indirecto triple
	if inode.I_block[14] != -1 {
		result += fmt.Sprintf(`
			<tr><td colspan="2" bgcolor="#FF9800"><b>BLOQUE INDIRECTO TRIPLE</b></td></tr>
			<tr><td><b>15</b></td><td>%d</td></tr>
		`, inode.I_block[14])
	}

	return result
}

// writeDotFile escribe el contenido DOT en un archivo
func writeDotFile(dotFileName string, dotContent string) error {
	dotFile, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear el archivo DOT: %v", err)
	}
	defer dotFile.Close()

	// Escribir el contenido DOT en el archivo
	_, err = dotFile.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo DOT: %v", err)
	}

	return nil
}

// generateInodeImage genera una imagen a partir del archivo DOT usando Graphviz
func generateInodeImage(dotFileName string, outputImage string) error {
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	return nil
}
