package structs

import (
	"backend/utils" // Asegúrate de ajustar el path del package "utils"
	"fmt"
	"os"
)

// FolderBlock representa un bloque de carpeta con 4 contenidos
type FolderBlock struct {
	B_content [4]FolderContent // 4 * 16 = 64 bytes
	// Total: 64 bytes
}

// FolderContent representa el contenido dentro de un bloque de carpeta
type FolderContent struct {
	B_name  [12]byte
	B_inodo int32
	// Total: 16 bytes
}

// Encode serializa la estructura FolderBlock en un archivo binario en la posición especificada
func (fb *FolderBlock) Encode(file *os.File, offset int64) error {
	// Utilizamos la función WriteToFile del paquete utils
	err := utils.WriteToFile(file, offset, fb)
	if err != nil {
		return fmt.Errorf("error writing FolderBlock to file: %w", err)
	}
	return nil
}

// Decode deserializa la estructura FolderBlock desde un archivo binario en la posición especificada
func (fb *FolderBlock) Decode(file *os.File, offset int64) error {
	// Utilizamos la función ReadFromFile del paquete utils
	err := utils.ReadFromFile(file, offset, fb)
	if err != nil {
		return fmt.Errorf("error reading FolderBlock from file: %w", err)
	}
	return nil
}

// Print imprime los atributos del bloque de carpeta
func (fb *FolderBlock) Print() {
	for i, content := range fb.B_content {
		name := string(content.B_name[:])
		fmt.Printf("Content %d:\n", i+1)
		fmt.Printf("  B_name: %s\n", name)
		fmt.Printf("  B_inodo: %d\n", content.B_inodo)
	}
}
