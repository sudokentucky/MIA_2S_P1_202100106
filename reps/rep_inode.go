package reps

import (
	structs "ArchivosP1/Structs"
	"ArchivosP1/utils"
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
	defer file.Close() // Asegurar el cierre del archivo al final

	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)

	// Inico del Dot
	dotContent := `digraph G {
		fontname="Helvetica,Arial,sans-serif"
		node [fontname="Helvetica,Arial,sans-serif"]
		edge [fontname="Helvetica,Arial,sans-serif"]
		rankdir=LR;
		node [shape=plaintext, fontsize=12];
		edge [color="#A8A8A8", arrowsize=0.8];

		// Definir una paleta de colores
		bgcolor="#F4F4F9"
		highlight="#FFCC80"
		light="#FFF3E0"
		dark="#E65100"
		border="#E0E0E0"
		text="#263238"
`

	// Iterar sobre cada inodo
	for i := int32(0); i < superblock.S_inodes_count; i++ {
		inode := &structs.Inode{}
		err := inode.Decode(file, int64(superblock.S_inode_start+(i*superblock.S_inode_size)))
		if err != nil {
			return fmt.Errorf("error al deserializar el inodo %d: %v", i, err)
		}

		// Verificar si el inodo está en uso (suponiendo que el valor por defecto cuando no está en uso es -1)
		if inode.I_uid == -1 {
			continue
		}

		// Convertir tiempos a string
		atime := time.Unix(int64(inode.I_atime), 0).Format(time.RFC3339)
		ctime := time.Unix(int64(inode.I_ctime), 0).Format(time.RFC3339)
		mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)
		dotContent += fmt.Sprintf(`inode%d [label=<
            <table border="1" cellborder="0" cellspacing="0" cellpadding="4" bgcolor="#FFF3E0" style="rounded">
                <tr><td bgcolor="#FFCC80" colspan="2"><b>REPORTE INODO %d</b></td></tr>
                <tr><td align="left"><b>i_uid</b></td><td align="left">%d</td></tr>
                <tr><td align="left"><b>i_gid</b></td><td align="left">%d</td></tr>
                <tr><td align="left"><b>i_size</b></td><td align="left">%d</td></tr>
                <tr><td align="left"><b>i_atime</b></td><td align="left">%s</td></tr>
                <tr><td align="left"><b>i_ctime</b></td><td align="left">%s</td></tr>
                <tr><td align="left"><b>i_mtime</b></td><td align="left">%s</td></tr>
                <tr><td align="left"><b>i_type</b></td><td align="left">%c</td></tr>
                <tr><td align="left"><b>i_perm</b></td><td align="left">%s</td></tr>
                <tr><td colspan="2" bgcolor="#FFCC80"><b>BLOQUES DIRECTOS</b></td></tr>
            `, i, i, inode.I_uid, inode.I_gid, inode.I_size, atime, ctime, mtime, rune(inode.I_type[0]), string(inode.I_perm[:]))
		for j, block := range inode.I_block[:12] {
			if block == -1 { // Si el bloque no está usado, continuar
				continue
			}
			dotContent += fmt.Sprintf("<tr><td align='left'><b>%d</b></td><td align='left'>%d</td></tr>", j+1, block)
		}

		if inode.I_block[12] != -1 {
			dotContent += fmt.Sprintf(`
                <tr><td colspan="2" bgcolor="#FFCC80"><b>BLOQUE INDIRECTO</b></td></tr>
                <tr><td align='left'><b>%d</b></td><td align='left'>%d</td></tr>
            `, 13, inode.I_block[12])
		}

		if inode.I_block[13] != -1 {
			dotContent += fmt.Sprintf(`
                <tr><td colspan="2" bgcolor="#FFCC80"><b>BLOQUE INDIRECTO DOBLE</b></td></tr>
                <tr><td align='left'><b>%d</b></td><td align='left'>%d</td></tr>
            `, 14, inode.I_block[13])
		}

		if inode.I_block[14] != -1 {
			dotContent += fmt.Sprintf(`
                <tr><td colspan="2" bgcolor="#FFCC80"><b>BLOQUE INDIRECTO TRIPLE</b></td></tr>
                <tr><td align='left'><b>%d</b></td><td align='left'>%d</td></tr>
            `, 15, inode.I_block[14])
		}

		dotContent += `</table>>];`

		if i < superblock.S_inodes_count-1 {
			dotContent += fmt.Sprintf("inode%d -> inode%d [color=\"#E65100\"];\n", i, i+1)
		}
	}

	dotContent += "}" // Fin del Dot
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

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	fmt.Println("Imagen de los inodos generada:", outputImage)
	return nil
}
