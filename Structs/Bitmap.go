package structs

import (
	utilidades "ArchivosP1/utils" // Importa el paquete utils
	"fmt"
	"os"
)

// CreateBitMaps crea los Bitmaps de inodos y bloques en el archivo especificado
func (sb *Superblock) CreateBitMaps(file *os.File) error {
	// Crear un buffer de n '0' para el bitmap de inodos
	inodeBuffer := make([]byte, sb.S_free_inodes_count)
	for i := range inodeBuffer {
		inodeBuffer[i] = '0'
	}

	// Escribir el buffer de bitmap de inodos en el archivo
	err := utilidades.WriteToFile(file, int64(sb.S_bm_inode_start), inodeBuffer)
	if err != nil {
		return fmt.Errorf("failed to write inode bitmap to file: %w", err)
	}

	// Crear un buffer de n 'O' para el bitmap de bloques
	blockBuffer := make([]byte, sb.S_free_blocks_count)
	for i := range blockBuffer {
		blockBuffer[i] = 'O'
	}

	// Escribir el buffer de bitmap de bloques en el archivo
	err = utilidades.WriteToFile(file, int64(sb.S_bm_block_start), blockBuffer)
	if err != nil {
		return fmt.Errorf("failed to write block bitmap to file: %w", err)
	}

	return nil
}
