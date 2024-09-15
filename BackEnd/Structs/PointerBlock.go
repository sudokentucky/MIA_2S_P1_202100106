package structs

import (
	"encoding/binary"
	"fmt"
	"os"
)

// PointerBlock : Estructura para guardar los bloques de apuntadores
type PointerBlock struct {
	B_pointers [16]int64 // Apuntadores a bloques de carpetas o datos
}

// FindFreePointer busca el primer apuntador libre en un bloque de apuntadores y devuelve su índice
func (pb *PointerBlock) FindFreePointer() (int, error) {
	for i, pointer := range pb.B_pointers {
		if pointer == -1 || pointer == 0 { // Usamos -1 o 0 para indicar apuntadores no asignados
			return i, nil
		}
	}
	return -1, fmt.Errorf("no hay apuntadores libres en el bloque de apuntadores")
}

// Encode serializa el PointerBlock en el archivo en la posición dada
func (pb *PointerBlock) Encode(file *os.File, offset int64) error {
	// Mover el cursor del archivo a la posición deseada
	_, err := file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("error buscando la posición en el archivo: %w", err)
	}

	// Escribir la estructura PointerBlock en el archivo
	err = binary.Write(file, binary.BigEndian, pb)
	if err != nil {
		return fmt.Errorf("error escribiendo el PointerBlock: %w", err)
	}
	return nil
}

// Decode deserializa el PointerBlock desde el archivo en la posición dada
func (pb *PointerBlock) Decode(file *os.File, offset int64) error {
	// Mover el cursor del archivo a la posición deseada
	_, err := file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("error buscando la posición en el archivo: %w", err)
	}

	// Leer la estructura PointerBlock desde el archivo
	err = binary.Read(file, binary.BigEndian, pb)
	if err != nil {
		return fmt.Errorf("error leyendo el PointerBlock: %w", err)
	}
	return nil
}
