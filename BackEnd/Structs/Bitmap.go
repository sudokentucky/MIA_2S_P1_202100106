package structs

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	FreeBlockBit     = 0
	OccupiedBlockBit = 1
)

// CreateBitMaps crea los Bitmaps de inodos y bloques en el archivo especificado
func (sb *Superblock) CreateBitMaps(file *os.File) error {
	// Crear el bitmap de inodos
	err := sb.createBitmap(file, sb.S_bm_inode_start, sb.S_inodes_count+sb.S_free_inodes_count, false)
	if err != nil {
		return fmt.Errorf("error creando bitmap de inodos: %w", err)
	}

	// Crear el bitmap de bloques
	err = sb.createBitmap(file, sb.S_bm_block_start, sb.S_blocks_count+sb.S_free_blocks_count, false)
	if err != nil {
		return fmt.Errorf("error creando bitmap de bloques: %w", err)
	}

	return nil
}

// createBitmap es una función auxiliar que escribe un bitmap en el archivo
// Cada bloque o inodo está representado por un bit
func (sb *Superblock) createBitmap(file *os.File, start int32, count int32, occupied bool) error {
	_, err := file.Seek(int64(start), 0)
	if err != nil {
		return fmt.Errorf("error buscando el inicio del bitmap: %w", err)
	}

	// Calcular el número de bytes necesarios (cada byte tiene 8 bits)
	byteCount := (count + 7) / 8

	// Crear el buffer de bytes con todos los bits en 0 (libres) o 1 (ocupados)
	fillByte := byte(0x00) // 00000000 (todos los bloques libres)
	if occupied {
		fillByte = 0xFF // 11111111 (todos los bloques ocupados)
	}

	buffer := make([]byte, byteCount)
	for i := range buffer {
		buffer[i] = fillByte
	}

	// Escribir el buffer en el archivo
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return fmt.Errorf("error escribiendo el bitmap: %w", err)
	}

	return nil
}

// UpdateBitmapInode actualiza el bitmap de inodos
func (sb *Superblock) UpdateBitmapInode(file *os.File, position int32, occupied bool) error {
	return sb.updateBitmap(file, sb.S_bm_inode_start, position, occupied)
}

// UpdateBitmapBlock actualiza el bitmap de bloques
func (sb *Superblock) UpdateBitmapBlock(file *os.File, position int32, occupied bool) error {
	return sb.updateBitmap(file, sb.S_bm_block_start, position, occupied)
}

// updateBitmap es una función auxiliar que actualiza un bit en un bitmap
func (sb *Superblock) updateBitmap(file *os.File, start int32, position int32, occupied bool) error {
	// Calcular el byte y el bit dentro de ese byte
	byteIndex := position / 8
	bitOffset := position % 8

	// Mover el puntero al byte correspondiente
	_, err := file.Seek(int64(start)+int64(byteIndex), 0)
	if err != nil {
		return fmt.Errorf("error buscando la posición en el bitmap: %w", err)
	}

	// Leer el byte actual
	var byteVal byte
	err = binary.Read(file, binary.LittleEndian, &byteVal)
	if err != nil {
		return fmt.Errorf("error leyendo el byte del bitmap: %w", err)
	}

	// Actualizar el bit correspondiente dentro del byte
	if occupied {
		byteVal |= (1 << bitOffset) // Poner el bit a 1 (ocupado)
	} else {
		byteVal &= ^(1 << bitOffset) // Poner el bit a 0 (libre)
	}

	// Mover el puntero de nuevo al byte correspondiente
	_, err = file.Seek(int64(start)+int64(byteIndex), 0)
	if err != nil {
		return fmt.Errorf("error buscando la posición en el bitmap para escribir: %w", err)
	}

	// Escribir el byte actualizado de vuelta en el archivo
	err = binary.Write(file, binary.LittleEndian, &byteVal)
	if err != nil {
		return fmt.Errorf("error escribiendo el byte actualizado del bitmap: %w", err)
	}

	return nil
}

// isBlockFree verifica si un bloque en el bitmap está libre
func (sb *Superblock) isBlockFree(file *os.File, start int32, position int32) (bool, error) {
	// Calcular el byte y el bit dentro del byte
	byteIndex := position / 8
	bitOffset := position % 8

	// Mover el puntero al byte correspondiente
	_, err := file.Seek(int64(start)+int64(byteIndex), 0)
	if err != nil {
		return false, fmt.Errorf("error buscando la posición en el bitmap: %w", err)
	}

	// Leer el byte actual
	var byteVal byte
	err = binary.Read(file, binary.LittleEndian, &byteVal)
	if err != nil {
		return false, fmt.Errorf("error leyendo el byte del bitmap: %w", err)
	}

	// Verificar si el bit está libre (0) o ocupado (1)
	return (byteVal & (1 << bitOffset)) == 0, nil
}

// isInodeFree verifica si un inodo en el bitmap está libre
func (sb *Superblock) isInodeFree(file *os.File, start int32, position int32) (bool, error) {
	byteIndex := position / 8 // Calcular el byte dentro del bitmap
	bitOffset := position % 8 // Calcular el bit dentro del byte

	// Leer el byte que contiene el bit correspondiente al inodo
	_, err := file.Seek(int64(start)+int64(byteIndex), 0)
	if err != nil {
		return false, fmt.Errorf("error buscando el byte en el bitmap de inodos: %w", err)
	}

	var byteVal byte
	err = binary.Read(file, binary.LittleEndian, &byteVal)
	if err != nil {
		return false, fmt.Errorf("error leyendo el byte del bitmap de inodos: %w", err)
	}

	// Verificar si el bit correspondiente está en 0 (libre)
	return (byteVal & (1 << bitOffset)) == 0, nil
}
