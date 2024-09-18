package structs

import (
	"backend/utils"
	"fmt"
	"os"
	"strings"
	"time"
)

// createFolderInInode crea una carpeta en un inodo específico
func (sb *Superblock) createFolderInInode(file *os.File, inodeIndex int32, parentsDir []string, destDir string) error {
	// Crear un nuevo inodo
	inode := &Inode{}
	fmt.Printf("Deserializando inodo %d\n", inodeIndex) // Depuración

	// Deserializar el inodo
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al deserializar inodo %d: %v", inodeIndex, err)
	}
	fmt.Printf("Inodo %d deserializado. Tipo: %c\n", inodeIndex, inode.I_type[0]) // Depuración

	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] != '0' {
		fmt.Printf("Inodo %d no es una carpeta, es de tipo: %c\n", inodeIndex, inode.I_type[0]) // Depuración
		return nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			fmt.Printf("Inodo %d no tiene más bloques asignados, terminando la búsqueda.\n", inodeIndex) // Depuración
			break
		}

		fmt.Printf("Deserializando bloque %d del inodo %d\n", blockIndex, inodeIndex) // Depuración
		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // Calcular la posición del bloque
		if err != nil {
			return fmt.Errorf("error al deserializar bloque %d: %v", blockIndex, err)
		}
		fmt.Printf("Bloque %d del inodo %d deserializado correctamente\n", blockIndex, inodeIndex) // Depuración

		// Iterar sobre cada contenido del bloque, desde el índice 2 (evitamos . y ..)
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]
			fmt.Printf("Verificando contenido en índice %d del bloque %d\n", indexContent, blockIndex) // Depuración

			// Si hay más carpetas padres en la ruta
			if len(parentsDir) != 0 {
				// Si el contenido está vacío, salir
				if content.B_inodo == -1 {
					fmt.Printf("No se encontró carpeta padre en inodo %d en la posición %d, terminando.\n", inodeIndex, indexContent) // Depuración
					break
				}

				// Obtener la carpeta padre más cercana
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return err
				}

				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")
				fmt.Printf("Comparando '%s' con el nombre de la carpeta padre '%s'\n", contentName, parentDirName) // Depuración

				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					fmt.Printf("Carpeta padre '%s' encontrada en inodo %d. Recursion para crear el siguiente directorio.\n", parentDirName, content.B_inodo) // Depuración
					// Llamada recursiva para seguir creando carpetas
					err := sb.createFolderInInode(file, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
					if err != nil {
						return err
					}
					return nil
				}
			} else { // Cuando llegamos al directorio destino (destDir)
				if content.B_inodo != -1 {
					fmt.Printf("El inodo %d ya está ocupado con otro contenido, saltando al siguiente.\n", content.B_inodo) // Depuración
					continue
				}

				fmt.Printf("Asignando el nombre del directorio '%s' al bloque en la posición %d\n", destDir, indexContent) // Depuración
				// Actualizar el contenido del bloque con el nuevo directorio
				copy(content.B_name[:], destDir)
				content.B_inodo = sb.S_inodes_count

				// Actualizar el bloque con el nuevo contenido
				block.B_content[indexContent] = content

				// Serializar el bloque
				err = block.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return fmt.Errorf("error al serializar el bloque %d: %v", blockIndex, err)
				}
				fmt.Printf("Bloque %d actualizado con éxito.\n", blockIndex) // Depuración

				// Crear el inodo de la nueva carpeta
				folderInode := &Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'}, // Tipo carpeta
					I_perm:  [3]byte{'6', '6', '4'},
				}

				fmt.Printf("Serializando el inodo de la carpeta '%s' (inodo %d)\n", destDir, sb.S_inodes_count) // Depuración
				// Serializar el inodo de la nueva carpeta
				err = folderInode.Encode(file, int64(sb.S_first_ino))
				if err != nil {
					return fmt.Errorf("error al serializar el inodo del directorio '%s': %v", destDir, err)
				}

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInode(file, sb.S_inodes_count, true)
				if err != nil {
					return fmt.Errorf("error al actualizar el bitmap de inodos para el directorio '%s': %v", destDir, err)
				}

				// Actualizar el superbloque con los nuevos valores de inodos
				sb.UpdateSuperblockAfterInodeAllocation()

				// Crear el bloque para la nueva carpeta
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: content.B_inodo},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				fmt.Printf("Serializando el bloque de la carpeta '%s'\n", destDir) // Depuración
				// Serializar el bloque de la carpeta
				err = folderBlock.Encode(file, int64(sb.S_first_blo))
				if err != nil {
					return fmt.Errorf("error al serializar el bloque del directorio '%s': %v", destDir, err)
				}

				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(file, sb.S_blocks_count, true)
				if err != nil {
					return fmt.Errorf("error al actualizar el bitmap de bloques para el directorio '%s': %v", destDir, err)
				}

				// Actualizar el superbloque con los nuevos valores de bloques
				sb.UpdateSuperblockAfterBlockAllocation()

				fmt.Printf("Directorio '%s' creado correctamente en inodo %d.\n", destDir, sb.S_inodes_count) // Depuración
				return nil
			}
		}
	}

	fmt.Printf("No se encontraron bloques disponibles para crear la carpeta '%s' en inodo %d\n", destDir, inodeIndex) // Depuración
	return nil
}

// CreateFolder crea una carpeta en el sistema de archivos
func (sb *Superblock) CreateFolder(file *os.File, parentsDir []string, destDir string) error {
	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		return sb.createFolderInInode(file, 0, parentsDir, destDir)
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.S_inodes_count; i++ { //Desde el inodo 0
		err := sb.createFolderInInode(file, i, parentsDir, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}
