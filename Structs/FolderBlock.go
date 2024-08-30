package structs

import (
	"fmt"
	"os"

	utilidades "ArchivosP1/utils" // Importa el paquete utils
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
	fmt.Println("┌──────────┬────────────────┬─────────────┐")
	fmt.Println("│ Índice   │ B_name         │ B_inodo     │")
	fmt.Println("├──────────┼────────────────┼─────────────┤")
	for i, content := range fb.B_content {
		name := string(content.B_name[:])
		fmt.Printf("│ %-8d │ %-14s │ %-11d │\n", i+1, name, content.B_inodo)
		fmt.Println("├──────────┼────────────────┼─────────────┤")
	}
	fmt.Println("└──────────┴────────────────┴─────────────┘")
}
