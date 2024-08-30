package structs

import (
	"fmt"
	"os"

	utilidades "ArchivosP1/utils" // Importa el paquete utils
)

// FileBlock representa un bloque de archivo en el sistema de archivos
type FileBlock struct {
	B_content [64]byte
}

// Encode serializa la estructura FileBlock en un archivo en la posición especificada
func (fb *FileBlock) Encode(file *os.File, offset int64) error {
	return utilidades.WriteToFile(file, offset, fb)
}

// Decode deserializa la estructura FileBlock desde un archivo en la posición especificada
func (fb *FileBlock) Decode(file *os.File, offset int64) error {
	return utilidades.ReadFromFile(file, offset, fb)
}

// Print prints the content of B_content as a string
func (fb *FileBlock) Print() {
	fmt.Printf("%s", fb.B_content)
}
