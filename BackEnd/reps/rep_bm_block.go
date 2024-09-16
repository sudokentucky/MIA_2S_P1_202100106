package reps

import (
	structures "backend/Structs"
	utils "backend/utils"
	"fmt"
	"os"
	"strings"
)

// ReportBMInode genera un reporte del bitmap de inodos y lo guarda en la ruta especificada
func ReportBMBlock(superblock *structures.Superblock, diskPath string, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	// Abrir el archivo de disco
	file, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}
	defer file.Close()

	// Calcular el número total de inodos
	totalBlocks := superblock.S_blocks_count + superblock.S_free_blocks_count

	// Obtener el contenido del bitmap de inodos
	var bitmapContent strings.Builder

	for i := int32(0); i < totalBlocks; i++ {
		// Establecer el puntero
		_, err := file.Seek(int64(superblock.S_bm_inode_start+i), 0)
		if err != nil {
			return fmt.Errorf("error al establecer el puntero en el archivo: %v", err)
		}

		// En el bitmap de bloques se usan los caracteres 'O' y 'X'
		char := make([]byte, 1)
		_, err = file.Read(char)
		if err != nil {
			return fmt.Errorf("error al leer el byte del archivo: %v", err)
		}

		// Agregar el carácter al contenido del bitmap
		bitmapContent.WriteByte(char[0])

		// Agregar un carácter de nueva línea cada 20 caracteres (20 inodos)
		if (i+1)%20 == 0 {
			bitmapContent.WriteString("\n")
		}
	}

	txtFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error al crear el archivo TXT: %v", err)
	}
	defer txtFile.Close()
	_, err = txtFile.WriteString(bitmapContent.String())
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo TXT: %v", err)
	}

	fmt.Println("Archivo del bitmap de inodos generado:", path)
	return nil
}
