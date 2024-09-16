package structs

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	FreeBlock     = '0'
	OccupiedBlock = '1'
)

// CreateBitMaps crea los Bitmaps de inodos y bloques en el archivo especificado
func (sb *Superblock) CreateBitMaps(file *os.File) error {
	// Crear el bitmap de inodos
	err := sb.createBitmap(file, sb.S_bm_inode_start, sb.S_free_inodes_count, FreeBlock)
	if err != nil {
		return fmt.Errorf("error creating inode bitmap: %w", err)
	}

	// Crear el bitmap de bloques
	err = sb.createBitmap(file, sb.S_bm_block_start, sb.S_free_blocks_count, FreeBlock)
	if err != nil {
		return fmt.Errorf("error creating block bitmap: %w", err)
	}

	return nil
}

// createBitmap es una función auxiliar que escribe un bitmap en el archivo
func (sb *Superblock) createBitmap(file *os.File, start int32, count int32, fillByte byte) error {
	_, err := file.Seek(int64(start), 0)
	if err != nil {
		return fmt.Errorf("error seeking to bitmap start: %w", err)
	}

	buffer := make([]byte, count)
	for i := range buffer {
		buffer[i] = fillByte
	}

	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return fmt.Errorf("error writing bitmap: %w", err)
	}

	return nil
}

// UpdateBitmapInode actualiza el bitmap de inodos
func (sb *Superblock) UpdateBitmapInode(file *os.File) error {
	return sb.updateBitmap(file, sb.S_bm_inode_start, sb.S_inodes_count, OccupiedBlock)
}

// UpdateBitmapBlock actualiza el bitmap de bloques
func (sb *Superblock) UpdateBitmapBlock(file *os.File) error {
	return sb.updateBitmap(file, sb.S_bm_block_start, sb.S_blocks_count, OccupiedBlock)
}

// updateBitmap es una función auxiliar que actualiza un bit en un bitmap
func (sb *Superblock) updateBitmap(file *os.File, start int32, position int32, newByte byte) error {
	_, err := file.Seek(int64(start)+int64(position), 0)
	if err != nil {
		return fmt.Errorf("error seeking to bitmap position: %w", err)
	}

	_, err = file.Write([]byte{newByte})
	if err != nil {
		return fmt.Errorf("error updating bitmap: %w", err)
	}

	return nil
}
