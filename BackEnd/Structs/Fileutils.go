package structs

import (
	"backend/utils"
	"fmt"
	"os"
	"strings"
	"time"
)

func (sb *Superblock) createFolderInInode(file *os.File, inodeIndex int32, parentsDir []string, destDir string) error {
	fmt.Printf("Iniciando la creación de carpeta en inodo %d\n", inodeIndex)
	fmt.Printf("Ruta de destino: %v\n", append(parentsDir, destDir))

	// Cargar el inodo actual
	inode := &Inode{}
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		fmt.Printf("Error al deserializar el inodo %d: %v\n", inodeIndex, err)
		return err
	}
	fmt.Printf("Inodo %d cargado correctamente\n", inodeIndex)

	// Verificar si es una carpeta
	if inode.I_type[0] == '1' { // Si es archivo, salir
		fmt.Printf("El inodo %d es un archivo. Saliendo...\n", inodeIndex)
		return nil
	}
	fmt.Printf("El inodo %d es una carpeta\n", inodeIndex)

	// Iterar sobre los bloques del inodo
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			fmt.Printf("No se encontró más espacio en los bloques del inodo %d. Deteniendo iteración...\n", inodeIndex)
			break
		}
		fmt.Printf("Procesando bloque %d del inodo %d\n", blockIndex, inodeIndex)

		block := &FolderBlock{}
		err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			fmt.Printf("Error al deserializar el bloque %d: %v\n", blockIndex, err)
			return err
		}
		fmt.Printf("Bloque %d del inodo %d cargado correctamente\n", blockIndex, inodeIndex)

		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]
			fmt.Printf("Revisando contenido en la posición %d del bloque %d\n", indexContent, blockIndex)

			// Verificar carpetas padres
			if len(parentsDir) != 0 {
				if content.B_inodo == -1 {
					fmt.Printf("Contenido vacío encontrado en la posición %d. Deteniendo...\n", indexContent)
					break
				}

				parentDir, err := utils.First(parentsDir)
				if err != nil {
					fmt.Printf("Error al obtener la carpeta padre: %v\n", err)
					return err
				}

				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")
				fmt.Printf("Comparando carpeta '%s' con el contenido '%s'\n", parentDirName, contentName)

				if strings.EqualFold(contentName, parentDirName) {
					fmt.Printf("Carpeta padre '%s' encontrada, continuando recursión en el inodo %d\n", parentDirName, content.B_inodo)
					return sb.createFolderInInode(file, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
				}
			} else {
				// Crear la carpeta en el espacio vacío
				if content.B_inodo != -1 {
					fmt.Printf("El inodo %d ya está ocupado. Continuando con el siguiente...\n", content.B_inodo)
					continue
				}

				// Asignar nuevo inodo para la carpeta
				newInodeIndex, err := sb.AssignNewInode(file, inode, indexContent)
				if err != nil {
					fmt.Printf("Error al asignar nuevo inodo: %v\n", err)
					return err
				}
				fmt.Printf("Nuevo inodo asignado: %d\n", newInodeIndex)

				// **IMPORTANTE: No debes usar el mismo índice `indexContent` para asignar un bloque.**

				// En su lugar, asigna un nuevo bloque en el primer índice disponible en `I_block` del inodo padre.
				newBlockIndex, err := sb.AssignNewBlock(file, inode, findFirstFreeBlockIndex(inode))
				if err != nil {
					fmt.Printf("Error al asignar nuevo bloque: %v\n", err)
					return err
				}
				fmt.Printf("Nuevo bloque asignado: %d\n", newBlockIndex)

				// Actualizar el contenido del bloque
				copy(content.B_name[:], destDir)
				content.B_inodo = newInodeIndex
				block.B_content[indexContent] = content
				fmt.Printf("Contenido del bloque actualizado con la carpeta '%s' en la posición %d\n", destDir, indexContent)

				// Guardar el bloque actualizado
				err = block.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					fmt.Printf("Error al guardar el bloque %d: %v\n", blockIndex, err)
					return err
				}
				fmt.Printf("Bloque %d guardado correctamente\n", blockIndex)

				// Inicializar el nuevo bloque de la carpeta con las entradas `.` y `..`
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: newInodeIndex},   // Apunta a sí mismo
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex}, // Apunta al inodo padre
						{B_name: [12]byte{'-'}, B_inodo: -1},              // Vacío
						{B_name: [12]byte{'-'}, B_inodo: -1},              // Vacío
					},
				}

				// Guardar el nuevo bloque de la carpeta
				err = folderBlock.Encode(file, int64(sb.S_block_start+(newBlockIndex*sb.S_block_size)))
				if err != nil {
					fmt.Printf("Error al guardar el nuevo bloque %d: %v\n", newBlockIndex, err)
					return err
				}
				fmt.Printf("Nuevo bloque %d (carpeta) guardado correctamente\n", newBlockIndex)

				// Inicializar el nuevo inodo de la carpeta
				folderInode := &Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{newBlockIndex, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'}, // Carpeta
					I_perm:  [3]byte{'6', '6', '4'},
				}

				// Guardar el nuevo inodo de la carpeta
				inodeOffset := sb.CalculateInodeOffset(newInodeIndex)
				err = WriteInodeToFile(file, inodeOffset, folderInode)
				if err != nil {
					fmt.Printf("Error al serializar el nuevo inodo %d: %v\n", newInodeIndex, err)
					return err
				}
				fmt.Printf("Nuevo inodo %d (carpeta) guardado correctamente en el offset %d\n", newInodeIndex, inodeOffset)

				return nil
			}
		}
	}

	fmt.Println("Finalizando la creación de la carpeta. No se encontró espacio disponible o la carpeta ya existe.")
	return nil
}

// Función auxiliar para encontrar el primer índice de bloque libre en el array I_block
func findFirstFreeBlockIndex(inode *Inode) int {
	for i, block := range inode.I_block {
		if block == -1 {
			return i
		}
	}
	return -1 // Retorna -1 si no hay espacio disponible
}

func (sb *Superblock) createFileInInode(file *os.File, inodeIndex int32, parentsDir []string, destFile string, fileSize int, fileContent []string) error {
	// Crear un nuevo inodo
	inode := &Inode{}
	// Deserializar el inodo
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		return nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			break
		}

		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return err
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.B_content[indexContent]

			// Sí las carpetas padre no están vacías debereamos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				// Si el contenido está vacío, salir
				if content.B_inodo == -1 {
					break
				}

				// Obtenemos la carpeta padre más cercana
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
					return err
				}

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
					// Actualizamos el inodo del archivo
					fileInode.I_block[i] = sb.S_blocks_count

					// Creamos el bloque del archivo
					fileBlock := &FileBlock{
						B_content: [64]byte{},
					}
					// Copiamos el contenido del archivo en el bloque
					copy(fileBlock.B_content[:], fileContent[i])

					// Serializar el bloque del archivo
					err = fileBlock.Encode(file, int64(sb.S_first_blo))
					if err != nil {
						return err
					}

					// Actualizar el bitmap de bloques con la posición del bloque
					err = sb.UpdateBitmapBlock(file, sb.S_blocks_count, true) // 'true' para marcarlo como ocupado
					if err != nil {
						return err
					}

					// Actualizar el superbloque
					sb.S_blocks_count++
					sb.S_free_blocks_count--
					sb.S_first_blo += sb.S_block_size
				}

				// Serializar el inodo del archivo
				err = fileInode.Encode(file, int64(sb.S_first_ino))
				if err != nil {
					return err
				}

				// Actualizar el bitmap de inodos con la posición del inodo
				err = sb.UpdateBitmapInode(file, sb.S_inodes_count, true) // 'true' para marcarlo como ocupado
				if err != nil {
					return err
				}

				// Actualizar el superbloque
				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size

				// Guardar el superbloque actualizado
				err = sb.Encode(file, int64(sb.S_inode_start))
				if err != nil {
					return fmt.Errorf("error al guardar el Superblock después de la creación del inodo: %w", err)
				}

				return nil
			}
		}

	}
	return nil
}

// CreateFolder crea una carpeta en el sistema de archivos
func (sb *Superblock) CreateFolder(file *os.File, parentsDir []string, destDir string) error {
	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		return sb.createFolderInInode(file, 0, parentsDir, destDir)
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.S_inodes_count; i++ {
		err := sb.createFolderInInode(file, i, parentsDir, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateFile crea un archivo en el sistema de archivos
func (sb *Superblock) CreateFile(file *os.File, parentsDir []string, destFile string, size int, cont []string) error {

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

	return nil
}
