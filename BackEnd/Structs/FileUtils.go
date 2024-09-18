package structs

import (
	"backend/utils"
	"fmt" // Importamos fmt para los mensajes de depuración
	"os"
	"strings"
	"time"
)

// createFolderinode crea una carpeta en un inodo específico
func (sb *Superblock) createFileInInode(file *os.File, inodeIndex int32, parentsDir []string, destFile string, fileSize int, fileContent []string) error {
	// Crear un nuevo inodo
	fmt.Printf("Intentando crear archivo '%s' en inodo index %d\n", destFile, inodeIndex) // Depuración
	inode := &Inode{}
	// Deserializar el inodo
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("Error al deserializar inodo %d: %v", inodeIndex, err)
	}

	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		fmt.Printf("El inodo %d es una carpeta, omitiendo.\n", inodeIndex) // Depuración
		return nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			fmt.Printf("El inodo %d no tiene más bloques, saliendo.\n", inodeIndex) // Depuración
			break
		}

		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // posición del bloque
		if err != nil {
			return fmt.Errorf("Error al deserializar bloque %d: %v", blockIndex, err)
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.B_content[indexContent]

			// Si las carpetas padre no están vacías, debemos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				if content.B_inodo == -1 {
					fmt.Printf("No se encontró carpeta padre en el inodo %d, saliendo.\n", inodeIndex) // Depuración
					break
				}

				// Obtener la carpeta padre más cercana
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return err
				}

				// Convertir B_name a string y eliminar los caracteres nulos
				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				// Convertir parentDir a string y eliminar los caracteres nulos
				parentDirName := strings.Trim(parentDir, "\x00 ")

				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					fmt.Printf("Encontrada carpeta padre '%s' en inodo %d\n", parentDirName, content.B_inodo) // Depuración
					// Si son las mismas, entonces entramos al inodo que apunta el bloque
					err := sb.createFileInInode(file, content.B_inodo, utils.RemoveElement(parentsDir, 0), destFile, fileSize, fileContent)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				if content.B_inodo != -1 {
					fmt.Printf("El inodo %d ya está ocupado, continuando.\n", content.B_inodo) // Depuración
					continue
				}

				// Actualizar el contenido del bloque
				copy(content.B_name[:], []byte(destFile))
				content.B_inodo = sb.S_inodes_count

				// Actualizar el bloque
				block.B_content[indexContent] = content

				// Serializar el bloque
				err = block.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return fmt.Errorf("Error al serializar bloque %d: %v", blockIndex, err)
				}

				fmt.Printf("Bloque actualizado para el archivo '%s' en el inodo %d\n", destFile, sb.S_inodes_count) // Depuración

				// Crear el inodo del archivo
				fileInode := &Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  int32(fileSize),
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'1'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				// Crear los bloques del archivo
				for i := 0; i < len(fileContent); i++ {
					fileInode.I_block[i] = sb.S_blocks_count

					// Crear el bloque del archivo
					fileBlock := &FileBlock{
						B_content: [64]byte{},
					}
					copy(fileBlock.B_content[:], fileContent[i])

					// Serializar el bloque
					err = fileBlock.Encode(file, int64(sb.S_first_blo))
					if err != nil {
						return fmt.Errorf("Error al serializar bloque de archivo: %v", err)
					}

					fmt.Printf("Bloque de archivo '%s' serializado correctamente.\n", destFile) // Depuración

					// Actualizar el bitmap de bloques
					err = sb.UpdateBitmapBlock(file, sb.S_blocks_count, true)
					if err != nil {
						return fmt.Errorf("Error al actualizar bitmap de bloque: %v", err)
					}

					// Actualizar el superbloque
					sb.UpdateSuperblockAfterBlockAllocation()
				}

				// Serializar el inodo del archivo
				err = fileInode.Encode(file, int64(sb.S_first_ino))
				if err != nil {
					return fmt.Errorf("Error al serializar inodo del archivo: %v", err)
				}

				fmt.Printf("Inodo del archivo '%s' serializado correctamente.\n", destFile) // Depuración

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInode(file, sb.S_inodes_count, true)
				if err != nil {
					return fmt.Errorf("Error al actualizar bitmap de inodo: %v", err)
				}

				// Actualizar el superbloque
				sb.UpdateSuperblockAfterInodeAllocation()

				fmt.Printf("Archivo '%s' creado correctamente en el inodo %d.\n", destFile, sb.S_inodes_count) // Depuración

				return nil
			}
		}
	}
	return nil
}

// CreateFile crea un archivo en el sistema de archivos
func (sb *Superblock) CreateFile(file *os.File, parentsDir []string, destFile string, size int, cont []string) error {
	fmt.Printf("Creando archivo '%s' con tamaño %d\n", destFile, size) // Depuración

	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		return sb.createFileInInode(file, 0, parentsDir, destFile, size, cont)
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.S_inodes_count; i++ {
		err := sb.createFileInInode(file, i, parentsDir, destFile, size, cont)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Archivo '%s' creado exitosamente.\n", destFile) // Depuración
	return nil
}
