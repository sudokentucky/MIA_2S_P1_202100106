package reps

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	structures "backend/Structs"
	utils "backend/utils"
)

// ReportBMInode genera un reporte del bitmap de inodos y lo guarda en la ruta especificada
func ReportBMInode(superblock *structures.Superblock, diskPath string, outputPath string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(outputPath)
	if err != nil {
		return fmt.Errorf("error creando carpetas padre: %v", err)
	}

	// Abrir el archivo de disco
	file, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}
	defer file.Close()

	// Calcular el número total de inodos
	totalInodes := superblock.S_inodes_count + superblock.S_free_inodes_count
	// Calcular cuántos bytes necesita el bitmap (cada byte tiene 8 bits)
	byteCount := (totalInodes + 7) / 8

	// Variable para almacenar el contenido del reporte del bitmap de inodos
	var bitmapContent strings.Builder

	for byteIndex := int32(0); byteIndex < byteCount; byteIndex++ {
		// Mover el puntero al byte correspondiente en el bitmap de inodos
		_, err := file.Seek(int64(superblock.S_bm_inode_start+byteIndex), 0)
		if err != nil {
			return fmt.Errorf("error al posicionar el archivo: %v", err)
		}

		// Leer un byte del bitmap
		var byteVal byte
		err = binary.Read(file, binary.LittleEndian, &byteVal)
		if err != nil {
			return fmt.Errorf("error al leer el byte del bitmap: %v", err)
		}

		// Procesar cada bit del byte (cada bit representa un inodo)
		for bitOffset := 0; bitOffset < 8; bitOffset++ {
			// Verificar si estamos fuera del rango total de inodos
			if byteIndex*8+int32(bitOffset) >= totalInodes {
				break
			}

			// Si el bit es 1, el inodo está ocupado, si es 0, está libre
			if (byteVal & (1 << bitOffset)) != 0 {
				bitmapContent.WriteByte('1') // Inodo ocupado
			} else {
				bitmapContent.WriteByte('0') // Inodo libre
			}

			// Añadir salto de línea cada 20 inodos
			if (byteIndex*8+int32(bitOffset)+1)%20 == 0 {
				bitmapContent.WriteString("\n")
			}
		}
	}

	// Guardar el reporte en el archivo especificado
	txtFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error al crear el archivo de reporte: %v", err)
	}
	defer txtFile.Close()

	_, err = txtFile.WriteString(bitmapContent.String())
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo de reporte: %v", err)
	}

	fmt.Println("Reporte del bitmap de inodos generado correctamente:", outputPath)
	return nil
}
