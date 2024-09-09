package reps

import (
	structs "backend/Structs"
	"backend/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ReportDisk genera un reporte de la estructura del disco y lo guarda en la ruta especificada
func ReportDisk(mbr *structs.MBR, path string, diskPath string) error {
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

	dotContent := `digraph G {
		fontname="Helvetica,Arial,sans-serif"
		node [fontname="Helvetica,Arial,sans-serif"]
		edge [fontname="Helvetica,Arial,sans-serif"]
		concentrate=True;
		rankdir=TB;
		node [shape=record];

		title [label="Reporte DISK" shape=plaintext fontname="Helvetica,Arial,sans-serif"];

		dsk [label="`

	// Calcular el tamaño total del disco y el tamaño usado
	totalSize := mbr.MbrSize
	usedSize := int32(0)

	// Agregar MBR al reporte
	dotContent += "{MBR}"

	// Recorrer las particiones del MBR y generar el contenido DOT
	for _, part := range mbr.MbrPartitions {
		if part.Part_size > 0 {
			// Calcular el porcentaje de uso
			percentage := (float64(part.Part_size) / float64(totalSize)) * 100
			usedSize += part.Part_size

			// Convertir Part_name a string y eliminar los caracteres nulos
			partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
			if part.Part_type[0] == 'P' {
				// Partición primaria
				dotContent += fmt.Sprintf("|{Primaria %s\\n%.2f%%}", partName, percentage)
			} else if part.Part_type[0] == 'E' {
				// Partición extendida
				dotContent += fmt.Sprintf("|{Extendida %.2f%%|{", percentage)
				ebrStart := part.Part_start
				ebrCount := 0
				ebrUsedSize := int32(0)

				for ebrStart != -1 {
					ebr, err := structs.ReadEBR(int32(ebrStart), file)
					if err != nil {
						return fmt.Errorf("error al leer EBR: %v", err)
					}

					ebrName := strings.TrimRight(string(ebr.Ebr_name[:]), "\x00")
					ebrPercentage := (float64(ebr.Ebr_size) / float64(totalSize)) * 100
					ebrUsedSize += ebr.Ebr_size

					// Agregar EBR y partición lógica al reporte
					if ebrCount > 0 {
						dotContent += "|"
					}
					dotContent += fmt.Sprintf("{EBR|Lógica %s\\n%.2f%%}", ebrName, ebrPercentage)

					// Actualizar el inicio para el próximo EBR
					ebrStart = ebr.Ebr_next
					ebrCount++
				}

				// Calcular espacio libre dentro de la partición extendida
				extendedFreeSize := part.Part_size - ebrUsedSize
				if extendedFreeSize > 0 {
					extendedFreePercentage := (float64(extendedFreeSize) / float64(totalSize)) * 100
					dotContent += fmt.Sprintf("|Libre %.2f%%", extendedFreePercentage)
				}

				dotContent += "}}"
			}
		}
	}

	// Calcular espacio libre restante y añadirlo si es necesario
	freeSize := totalSize - usedSize
	if freeSize > 0 {
		freePercentage := (float64(freeSize) / float64(totalSize)) * 100
		dotContent += fmt.Sprintf("|Libre %.2f%%", freePercentage)
	}

	// Cerrar el nodo de disco y completar el DOT
	dotContent += `"];

		title -> dsk [style=invis];
	}`

	// Crear el archivo DOT
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

	// Generar la imagen con Graphviz
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	fmt.Println("Reporte de disco generado:", outputImage)
	return nil
}
