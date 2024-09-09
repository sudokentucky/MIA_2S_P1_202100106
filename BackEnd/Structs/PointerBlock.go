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

// assignPointerBlock asigna un nuevo bloque de datos a un PointerBlock existente
func assignPointerBlock(file *os.File, sb *Superblock, pb *PointerBlock, blockOffset int64) (int64, error) {
	// Buscar un apuntador libre en el bloque de apuntadores
	freeIndex, err := pb.FindFreePointer()
	if err != nil {
		return -1, fmt.Errorf("error encontrando apuntador libre: %w", err)
	}

	// Obtener un nuevo bloque de datos
	newBlock, err := sb.GetNextFreeBlock(file)
	if err != nil {
		return -1, fmt.Errorf("error obteniendo un nuevo bloque: %w", err)
	}

	// Asignar el nuevo bloque en el bloque de apuntadores
	pb.B_pointers[freeIndex] = int64(newBlock) // Conversión a int64

	// Guardar el bloque de apuntadores actualizado en el archivo
	err = pb.Encode(file, blockOffset)
	if err != nil {
		return -1, fmt.Errorf("error guardando el bloque de apuntadores: %w", err)
	}

	return int64(newBlock), nil // Conversión a int64 en el retorno también
}
