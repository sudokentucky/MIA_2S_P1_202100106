package reps

import (
	structs "backend/Structs"
	"backend/utils"
	"fmt"
	"os"
)

// ReportBlockConnections imprime los bloques y sus conexiones directamente en la consola
func ReportBlock(superblock *structs.Superblock, diskPath string, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	// Abrir el archivo de disco
	file, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}
	defer file.Close()

	// En lugar de crear el archivo DOT y generar una imagen, solo imprimimos los bloques
	fmt.Println("Recorriendo inodos y bloques...")

	// Recorrer todos los inodos para obtener los bloques asociados
	err = printBlockConnections(superblock, file)
	if err != nil {
		return err
	}

	fmt.Println("Finalizado.")
	return nil
}

// printBlockConnections recorre los inodos y los bloques asociados, imprimiéndolos
func printBlockConnections(superblock *structs.Superblock, file *os.File) error {
	for i := int32(0); i < superblock.S_inodes_count; i++ {
		inode := &structs.Inode{}
		err := inode.Decode(file, int64(superblock.S_inode_start+(i*superblock.S_inode_size)))
		if err != nil {
			return fmt.Errorf("error al deserializar el inodo %d: %v", i, err)
		}

		// Verificar si el inodo está en uso
		if inode.I_uid == -1 || inode.I_uid == 0 {
			continue
		}

		// Imprimir información del inodo
		fmt.Printf("Inodo %d:\n", i)
		inode.Print()

		// Recorrer los bloques asociados al inodo
		for _, block := range inode.I_block {
			if block != -1 { // Bloques asignados
				// Imprimir el bloque dependiendo de si es de archivo o de carpeta
				printBlock(block, inode, superblock, file)
			}
		}
	}
	return nil
}

// printBlock imprime la información de un bloque específico dependiendo de su tipo (archivo o carpeta)
func printBlock(blockIndex int32, inode *structs.Inode, superblock *structs.Superblock, file *os.File) error {
	// Obtener el desplazamiento en el archivo donde se encuentra el bloque
	blockOffset := int64(superblock.S_block_start + (blockIndex * superblock.S_block_size))

	// Dependiendo del tipo de inodo, leer e imprimir el bloque correspondiente
	if inode.I_type[0] == '0' { // Bloque de carpeta
		folderBlock := &structs.FolderBlock{}
		err := folderBlock.Decode(file, blockOffset)
		if err != nil {
			return fmt.Errorf("error al decodificar bloque de carpeta %d: %w", blockIndex, err)
		}
		fmt.Printf("\nBloque de carpeta %d:\n", blockIndex)
		folderBlock.Print()
	} else if inode.I_type[0] == '1' { // Bloque de archivo
		fileBlock := &structs.FileBlock{}
		err := fileBlock.Decode(file, blockOffset)
		if err != nil {
			return fmt.Errorf("error al decodificar bloque de archivo %d: %w", blockIndex, err)
		}
		fmt.Printf("\nBloque de archivo %d:\n", blockIndex)
		fileBlock.Print()
	}
	return nil
}
