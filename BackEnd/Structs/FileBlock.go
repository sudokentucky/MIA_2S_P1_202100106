package structs

import (
	"backend/utils" // Asegúrate de ajustar el path del package "utils"
	"fmt"
	"os"
)

type FileBlock struct {
	B_content [64]byte
	// Total: 64 bytes
}

// Encode serializa la estructura FileBlock en un archivo binario en la posición especificada
func (fb *FileBlock) Encode(file *os.File, offset int64) error {
	// Utilizamos la función WriteToFile del paquete utils
	err := utils.WriteToFile(file, offset, fb)
	if err != nil {
		return fmt.Errorf("error writing FileBlock to file: %w", err)
	}
	return nil
}

// Decode deserializa la estructura FileBlock desde un archivo binario en la posición especificada
func (fb *FileBlock) Decode(file *os.File, offset int64) error {
	// Utilizamos la función ReadFromFile del paquete utils
	err := utils.ReadFromFile(file, offset, fb)
	if err != nil {
		return fmt.Errorf("error reading FileBlock from file: %w", err)
	}
	return nil
}

// PrintContent imprime el contenido de B_content como una cadena
func (fb *FileBlock) Print() {
	fmt.Printf("%s", fb.B_content)
}
