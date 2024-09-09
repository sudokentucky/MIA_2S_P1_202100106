package structs

import (
	"fmt"
	"os"

	utilidades "backend/utils" // Importa el paquete utils
)

// FolderBlock representa un bloque de carpeta en el sistema de archivos
type FolderBlock struct {
	B_content [4]FolderContent // 4 * 16 = 64 bytes
}

// FolderContent representa el contenido de un bloque de carpeta
type FolderContent struct {
	B_name  [12]byte
	B_inodo int32
}

// Encode serializa la estructura FolderBlock en un archivo en la posición especificada
func (fb *FolderBlock) Encode(file *os.File, offset int64) error {
	return utilidades.WriteToFile(file, offset, fb)
}

// Decode deserializa la estructura FolderBlock desde un archivo en la posición especificada
func (fb *FolderBlock) Decode(file *os.File, offset int64) error {
	return utilidades.ReadFromFile(file, offset, fb)
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
