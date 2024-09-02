package reps

import (
	structures "ArchivosP1/Structs"
	utils "ArchivosP1/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ReportMBR genera un reporte del MBR y lo guarda en la ruta especificada
func ReportMBR(mbr *structures.MBR, path string, file *os.File) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)

	// Definir la paleta de colores
	primaryColor := "#FFDDC1"     // Color para particiones primarias
	extendedColor := "#C1E1C1"    // Color para particiones extendidas
	logicalColor := "#C1D1FF"     // Color para particiones lógicas
	ebrColor := "#FFD1DC"         // Color para EBR
	unallocatedColor := "#FFFFFF" // Color para espacios no asignados

	// Definir el contenido DOT con una tabla
	dotContent := fmt.Sprintf(`digraph G {
        node [shape=plaintext]
        tabla [label=<
            <table border="0" cellborder="1" cellspacing="0">
                <tr><td colspan="2" bgcolor="#F8D7DA"><b>REPORTE MBR</b></td></tr>
                <tr><td bgcolor="#F5B7B1">mbr_tamano</td><td bgcolor="#F5B7B1">%d</td></tr>
                <tr><td bgcolor="#F5B7B1">mrb_fecha_creacion</td><td bgcolor="#F5B7B1">%s</td></tr>
                <tr><td bgcolor="#F5B7B1">mbr_disk_signature</td><td bgcolor="#F5B7B1">%d</td></tr>
            `, mbr.MbrSize, time.Unix(int64(mbr.MbrCreacionDate), 0), mbr.MbrDiskSignature)

	// Calcular el tamaño total del disco y mantener un seguimiento del espacio no asignado
	totalSize := mbr.MbrSize
	allocatedSize := int32(0)

	// Agregar las particiones a la tabla
	for i, part := range mbr.MbrPartitions {
		// Solo incluir particiones válidas
		if part.Part_size > 0 && part.Part_start > 0 {
			// Calcular el espacio no asignado antes de esta partición
			if part.Part_start > allocatedSize {
				unallocatedSize := part.Part_start - allocatedSize
				dotContent += fmt.Sprintf(`
                    <tr><td colspan="2" bgcolor="%s"><b>ESPACIO NO ASIGNADO (Tamaño: %d bytes)</b></td></tr>
                `, unallocatedColor, unallocatedSize)
				allocatedSize += unallocatedSize
			}

			// Convertir Part_name a string y eliminar los caracteres nulos
			partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
			// Convertir Part_status, Part_type y Part_fit a char
			partStatus := rune(part.Part_status[0])
			partType := rune(part.Part_type[0])
			partFit := rune(part.Part_fit[0])

			// Definir el color de fondo dependiendo del tipo de partición
			rowColor := ""
			switch partType {
			case 'P':
				rowColor = primaryColor
			case 'E':
				rowColor = extendedColor
			}

			// Agregar la partición a la tabla
			dotContent += fmt.Sprintf(`
                <tr><td colspan="2" bgcolor="%s"><b>PARTICIÓN %d</b></td></tr>
                <tr><td bgcolor="%s">part_status</td><td bgcolor="%s">%c</td></tr>
                <tr><td bgcolor="%s">part_type</td><td bgcolor="%s">%c</td></tr>
                <tr><td bgcolor="%s">part_fit</td><td bgcolor="%s">%c</td></tr>
                <tr><td bgcolor="%s">part_start</td><td bgcolor="%s">%d</td></tr>
                <tr><td bgcolor="%s">part_size</td><td bgcolor="%s">%d</td></tr>
                <tr><td bgcolor="%s">part_name</td><td bgcolor="%s">%s</td></tr>
            `, rowColor, i+1,
				rowColor, rowColor, partStatus,
				rowColor, rowColor, partType,
				rowColor, rowColor, partFit,
				rowColor, rowColor, part.Part_start,
				rowColor, rowColor, part.Part_size,
				rowColor, rowColor, partName)

			allocatedSize += part.Part_size

			// Si es una partición extendida, mostrar EBRs y particiones lógicas
			if partType == 'E' {
				ebrStart := part.Part_start
				dotContent += fmt.Sprintf(`
                    <tr><td colspan="2" bgcolor="%s"><b>PART. EXTENDIDA (Inicio: %d)</b></td></tr>
                `, extendedColor, ebrStart)

				// Iterar sobre los EBRs
				for ebrStart != -1 {
					ebr, err := structures.ReadEBR(int32(ebrStart), file)
					if err != nil {
						return fmt.Errorf("error al leer EBR: %v", err)
					}

					ebrName := strings.TrimRight(string(ebr.Ebr_name[:]), "\x00")
					ebrFit := rune(ebr.Ebr_fit[0])

					// Mostrar información del EBR
					dotContent += fmt.Sprintf(`
                        <tr><td colspan="2" bgcolor="%s"><b>EBR (Inicio: %d)</b></td></tr>
                        <tr><td bgcolor="%s">ebr_fit</td><td bgcolor="%s">%c</td></tr>
                        <tr><td bgcolor="%s">ebr_start</td><td bgcolor="%s">%d</td></tr>
                        <tr><td bgcolor="%s">ebr_size</td><td bgcolor="%s">%d</td></tr>
                        <tr><td bgcolor="%s">ebr_next</td><td bgcolor="%s">%d</td></tr>
                        <tr><td bgcolor="%s">ebr_name</td><td bgcolor="%s">%s</td></tr>
                    `, ebrColor, ebrStart,
						ebrColor, ebrColor, ebrFit,
						ebrColor, ebrColor, ebr.Ebr_start,
						ebrColor, ebrColor, ebr.Ebr_size,
						ebrColor, ebrColor, ebr.Ebr_next,
						ebrColor, ebrColor, ebrName)

					// Si hay una partición lógica después del EBR
					if ebr.Ebr_size > 0 {
						dotContent += fmt.Sprintf(`
                            <tr><td colspan="2" bgcolor="%s"><b>PART. LÓGICA (Inicio: %d)</b></td></tr>
                        `, logicalColor, ebr.Ebr_start)
					}

					allocatedSize += ebr.Ebr_size
					ebrStart = int32(ebr.Ebr_next)
				}
			}
		}
	}

	// Calcular espacio no asignado restante al final del disco
	if allocatedSize < totalSize {
		unallocatedSize := totalSize - allocatedSize
		dotContent += fmt.Sprintf(`
            <tr><td colspan="2" bgcolor="%s"><b>ESPACIO NO ASIGNADO (Tamaño: %d bytes)</b></td></tr>
        `, unallocatedColor, unallocatedSize)
	}

	// Cerrar la tabla y el contenido DOT
	dotContent += "</table>>] }"

	// Guardar el contenido DOT en un archivo
	file, err = os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}

	// Ejecutar el comando Graphviz para generar la imagen
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	fmt.Println("Imagen de la tabla generada:", outputImage)
	return nil
}
