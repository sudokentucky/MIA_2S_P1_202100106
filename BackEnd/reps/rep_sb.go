package reps

import (
	structs "backend/Structs"
	"backend/utils"
	"fmt"
	"os/exec"
	"time"
)

// ReportSuperblock genera un reporte del Superbloque en formato de tabla
func ReportSuperblock(superblock *structs.Superblock, diskPath string, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)

	// Inicio del Dot
	dotContent := initDotGraphForSuperblock(superblock)

	// Crear el archivo DOT
	err = writeDotFile(dotFileName, dotContent)
	if err != nil {
		return err
	}

	// Ejecutar Graphviz para generar la imagen
	err = generateSuperblockImage(dotFileName, outputImage)
	if err != nil {
		return err
	}

	fmt.Println("Imagen del Superbloque generada:", outputImage)
	return nil
}

// generateSuperblockImage genera una imagen a partir del archivo DOT usando Graphviz
func generateSuperblockImage(dotFileName string, outputImage string) error {
	// Crear el comando para ejecutar Graphviz
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)

	// Ejecutar el comando y esperar a que termine
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz para generar la imagen del Superbloque: %v", err)
	}

	return nil
}

// initDotGraphForSuperblock inicializa el contenido básico del archivo DOT para el Superbloque
func initDotGraphForSuperblock(superblock *structs.Superblock) string {
	// Convertir tiempos a formato legible
	mtime := time.Unix(int64(superblock.S_mtime), 0).Format(time.RFC3339)
	umtime := time.Unix(int64(superblock.S_umtime), 0).Format(time.RFC3339)

	// Generar la tabla del Superbloque
	dotContent := `
		digraph G {
			fontname="Helvetica,Arial,sans-serif"
			node [fontname="Helvetica,Arial,sans-serif", shape=plain, fontsize=12];
			edge [fontname="Helvetica,Arial,sans-serif", color="#FF7043", arrowsize=0.8];
			bgcolor="#FAFAFA";
			rankdir=TB;

			superblockTable [label=<
				<table border="0" cellborder="1" cellspacing="0" cellpadding="10" bgcolor="#FFF9C4" style="rounded">
					<tr><td colspan="2" bgcolor="#4CAF50" align="center"><b>REPORTE DEL SUPERBLOQUE</b></td></tr>
					<tr><td><b>Cantidad de Inodos</b></td><td>%d</td></tr>
					<tr><td><b>Cantidad de Bloques</b></td><td>%d</td></tr>
					<tr><td><b>Inodos Libres</b></td><td>%d</td></tr>
					<tr><td><b>Bloques Libres</b></td><td>%d</td></tr>
					<tr><td><b>Tamaño de Inodo</b></td><td>%d bytes</td></tr>
					<tr><td><b>Tamaño de Bloque</b></td><td>%d bytes</td></tr>
					<tr><td><b>Primer Inodo Libre</b></td><td>%d</td></tr>
					<tr><td><b>Primer Bloque Libre</b></td><td>%d</td></tr>
					<tr><td><b>Inicio Bitmap de Inodos</b></td><td>%d</td></tr>
					<tr><td><b>Inicio Bitmap de Bloques</b></td><td>%d</td></tr>
					<tr><td><b>Última Modificación</b></td><td>%s</td></tr>
					<tr><td><b>Último Montaje</b></td><td>%s</td></tr>
				</table>>];
		}
	`

	// Formatear el contenido con los datos del superbloque
	dotContent = fmt.Sprintf(dotContent,
		superblock.S_inodes_count,
		superblock.S_blocks_count,
		superblock.S_free_inodes_count,
		superblock.S_free_blocks_count,
		superblock.S_inode_size,
		superblock.S_block_size,
		superblock.S_first_ino,
		superblock.S_first_blo,
		superblock.S_bm_inode_start,
		superblock.S_bm_block_start,
		mtime,
		umtime,
	)

	return dotContent
}
