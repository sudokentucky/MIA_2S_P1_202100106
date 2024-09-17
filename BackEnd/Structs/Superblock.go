package structs

import (
	utilidades "backend/utils" // Importa el paquete utils
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

type Superblock struct {
	S_filesystem_type   int32   // Número que identifica el sistema de archivos usado
	S_inodes_count      int32   // Número total de inodos creados
	S_blocks_count      int32   // Número total de bloques creados
	S_free_blocks_count int32   // Número de bloques libres
	S_free_inodes_count int32   // Número de inodos libres
	S_mtime             float64 // Última fecha en que el sistema fue montado
	S_umtime            float64 // Última fecha en que el sistema fue desmontado
	S_mnt_count         int32   // Número de veces que se ha montado el sistema
	S_magic             int32   // Valor que identifica el sistema de archivos
	S_inode_size        int32   // Tamaño de la estructura inodo
	S_block_size        int32   // Tamaño de la estructura bloque
	S_first_ino         int32   // Primer inodo libre
	S_first_blo         int32   // Primer bloque libre
	S_bm_inode_start    int32   // Inicio del bitmap de inodos
	S_bm_block_start    int32   // Inicio del bitmap de bloques
	S_inode_start       int32   // Inicio de la tabla de inodos
	S_block_start       int32   // Inicio de la tabla de bloques
}

// Encode codifica la estructura Superblock en un archivo
func (sb *Superblock) Encode(file *os.File, offset int64) error {
	return utilidades.WriteToFile(file, offset, sb)
}

// Decode decodifica la estructura Superblock desde un archivo
func (sb *Superblock) Decode(file *os.File, offset int64) error {
	return utilidades.ReadFromFile(file, offset, sb)
}

// CreateUsersFile crea el archivo users.txt que contendrá la información de los usuarios
func (sb *Superblock) CreateUsersFile(file *os.File) error {
	// ----------- Crear Inodo Raíz -----------
	rootInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'}, // Tipo carpeta
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Escribir el inodo raíz
	err := utilidades.WriteToFile(file, int64(sb.S_first_ino), rootInode)
	if err != nil {
		return fmt.Errorf("error al escribir el inodo raíz: %w", err)
	}

	// Actualizar bitmap de inodos
	err = sb.UpdateBitmapInode(file)
	if err != nil {
		return fmt.Errorf("error al actualizar bitmap de inodos: %w", err)
	}

	// Actualizar el contador de inodos y el puntero al primer inodo libre
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Guardar el Superblock actualizado
	err = sb.Encode(file, int64(sb.S_first_ino))
	if err != nil {
		return fmt.Errorf("error al guardar el Superblock: %w", err)
	}

	// ----------- Crear Bloque Raíz (/ carpeta) -----------
	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},                                                         // Apunta a sí mismo
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},                                                    // Apunta al padre
			{B_name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: sb.S_inodes_count}, // Apunta a users.txt
			{B_name: [12]byte{'-'}, B_inodo: -1},                                                        // Vacío
		},
	}

	// Escribir el bloque raíz
	err = utilidades.WriteToFile(file, int64(sb.S_first_blo), rootBlock)
	if err != nil {
		return fmt.Errorf("error al escribir el bloque raíz: %w", err)
	}

	// Actualizar bitmap de bloques
	err = sb.UpdateBitmapBlock(file)
	if err != nil {
		return fmt.Errorf("error al actualizar el bitmap de bloques: %w", err)
	}

	// Actualizar el contador de bloques y el puntero al primer bloque libre
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// Guardar el Superblock actualizado
	err = sb.Encode(file, int64(sb.S_first_blo))
	if err != nil {
		return fmt.Errorf("error al guardar el Superblock después de la creación del bloque raíz: %w", err)
	}

	// ----------- Crear Inodo para /users.txt -----------
	// Crear usuarios y grupos utilizando la estructura User
	// Crear grupo root
	rootGroup := NewGroup("1", "root")
	// Crear usuario root
	rootUser := NewUser("1", "root", "root", "123")

	// Concatenar la información de usuarios y grupos
	usersText := fmt.Sprintf("%s\n%s\n", rootGroup.ToString(), rootUser.ToString())

	usersInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(usersText)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // Apunta al bloque de users.txt
		I_type:  [1]byte{'1'},                                                                         // Tipo archivo
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Escribir el inodo de users.txt
	err = utilidades.WriteToFile(file, int64(sb.S_first_ino), usersInode)
	if err != nil {
		return fmt.Errorf("error al escribir el inodo de users.txt: %w", err)
	}

	// Actualizar bitmap de inodos
	err = sb.UpdateBitmapInode(file)
	if err != nil {
		return fmt.Errorf("error al actualizar bitmap de inodos para users.txt: %w", err)
	}

	// Actualizar el contador de inodos y el puntero al primer inodo libre
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Guardar el Superblock actualizado
	err = sb.Encode(file, int64(sb.S_first_ino))
	if err != nil {
		return fmt.Errorf("error al guardar el Superblock después de la creación del inodo de users.txt: %w", err)
	}

	// ----------- Crear Bloque para users.txt -----------
	usersBlock := &FileBlock{}
	copy(usersBlock.B_content[:], usersText)

	// Escribir el bloque de users.txt
	err = utilidades.WriteToFile(file, int64(sb.S_first_blo), usersBlock)
	if err != nil {
		return fmt.Errorf("error al escribir el bloque de users.txt: %w", err)
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(file)
	if err != nil {
		return fmt.Errorf("error al actualizar el bitmap de bloques para users.txt: %w", err)
	}

	// Actualizar el contador de bloques y el puntero al primer bloque libre
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// Guardar el Superblock actualizado
	err = sb.Encode(file, int64(sb.S_first_blo))
	if err != nil {
		return fmt.Errorf("error al guardar el Superblock después de la creación del bloque de users.txt: %w", err)
	}

	fmt.Println("Archivo users.txt creado correctamente.")
	fmt.Println("Superbloque después de la creación de users.txt:")
	sb.Print()
	fmt.Println("\nBloques:")
	sb.PrintBlocks(file.Name())
	fmt.Println("\nInodos:")
	sb.PrintInodes(file.Name())
	return nil
}

// Print imprime los valores de la estructura SuperBlock
func (sb *Superblock) Print() {
	fmt.Printf("%-25s %-10s\n", "Campo", "Valor")
	fmt.Printf("%-25s %-10s\n", "-------------------------", "----------")
	fmt.Printf("%-25s %-10d\n", "S_filesystem_type:", sb.S_filesystem_type)
	fmt.Printf("%-25s %-10d\n", "S_inodes_count:", sb.S_inodes_count)
	fmt.Printf("%-25s %-10d\n", "S_blocks_count:", sb.S_blocks_count)
	fmt.Printf("%-25s %-10d\n", "S_free_blocks_count:", sb.S_free_blocks_count)
	fmt.Printf("%-25s %-10d\n", "S_free_inodes_count:", sb.S_free_inodes_count)
	fmt.Printf("%-25s %-10s\n", "S_mtime:", time.Unix(int64(sb.S_mtime), 0).Format("02/01/2006 15:04"))
	fmt.Printf("%-25s %-10s\n", "S_umtime:", time.Unix(int64(sb.S_umtime), 0).Format("02/01/2006 15:04"))
	fmt.Printf("%-25s %-10d\n", "S_mnt_count:", sb.S_mnt_count)
	fmt.Printf("%-25s %-10x\n", "S_magic:", sb.S_magic)
	fmt.Printf("%-25s %-10d\n", "S_inode_size:", sb.S_inode_size)
	fmt.Printf("%-25s %-10d\n", "S_block_size:", sb.S_block_size)
	fmt.Printf("%-25s %-10d\n", "S_first_ino:", sb.S_first_ino)
	fmt.Printf("%-25s %-10d\n", "S_first_blo:", sb.S_first_blo)
	fmt.Printf("%-25s %-10d\n", "S_bm_inode_start:", sb.S_bm_inode_start)
	fmt.Printf("%-25s %-10d\n", "S_bm_block_start:", sb.S_bm_block_start)
	fmt.Printf("%-25s %-10d\n", "S_inode_start:", sb.S_inode_start)
	fmt.Printf("%-25s %-10d\n", "S_block_start:", sb.S_block_start)
}

// PrintInodes imprime los inodos desde el archivo
func (sb *Superblock) PrintInodes(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	fmt.Println("\nInodos\n----------------")
	inodes := make([]Inode, sb.S_inodes_count)

	// Deserializar todos los inodos en memoria
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &inodes[i]
		err := utilidades.ReadFromFile(file, int64(sb.S_inode_start+(i*int32(binary.Size(Inode{})))), inode)
		if err != nil {
			return fmt.Errorf("failed to decode inode %d: %w", i, err)
		}
	}

	// Imprimir los inodos
	for i, inode := range inodes {
		fmt.Printf("\nInodo %d:\n", i)
		inode.Print()
	}

	return nil
}

// PrintBlocks imprime los bloques desde el archivo
func (sb *Superblock) PrintBlocks(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	fmt.Println("\nBloques\n----------------")
	inodes := make([]Inode, sb.S_inodes_count)

	// Deserializar todos los inodos en memoria
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &inodes[i]
		err := utilidades.ReadFromFile(file, int64(sb.S_inode_start+(i*int32(binary.Size(Inode{})))), inode)
		if err != nil {
			return fmt.Errorf("failed to decode inode %d: %w", i, err)
		}
	}

	// Imprimir los bloques
	for _, inode := range inodes {
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				break
			}
			if inode.I_type[0] == '0' {
				block := &FolderBlock{}
				err := utilidades.ReadFromFile(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)), block)
				if err != nil {
					return fmt.Errorf("failed to decode folder block %d: %w", blockIndex, err)
				}
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
			} else if inode.I_type[0] == '1' {
				block := &FileBlock{}
				err := utilidades.ReadFromFile(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)), block)
				if err != nil {
					return fmt.Errorf("failed to decode file block %d: %w", blockIndex, err)
				}
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
			}
		}
	}

	return nil
}

func (sb *Superblock) FindNextFreeBlock(file *os.File) (int32, error) {
	// Mover el puntero al inicio del bitmap de bloques
	_, err := file.Seek(int64(sb.S_bm_block_start), 0)
	if err != nil {
		return -1, fmt.Errorf("error buscando el inicio del bitmap de bloques: %w", err)
	}

	// Leer el bitmap de bloques en memoria
	buffer := make([]byte, sb.S_free_blocks_count)
	_, err = file.Read(buffer)
	if err != nil {
		return -1, fmt.Errorf("error leyendo el bitmap de bloques: %w", err)
	}

	// Buscar el primer bloque libre ('0')
	for i, bit := range buffer {
		if bit == FreeBlock {
			// Bloque libre encontrado
			buffer[i] = OccupiedBlock

			// Actualizar el bitmap en el archivo
			if err := sb.updateBitmap(file, sb.S_bm_block_start, int32(i), OccupiedBlock); err != nil {
				return -1, fmt.Errorf("error actualizando el bitmap de bloques: %w", err)
			}

			// Log opcional para confirmar el bloque asignado
			fmt.Printf("Bloque asignado: %d\n", i)

			// Devolver el índice del bloque libre encontrado
			return int32(i), nil
		}
	}

	// Si no se encuentran bloques libres, devolver un error
	return -1, fmt.Errorf("no hay bloques disponibles")
}

func (sb *Superblock) CalculateInodeOffset(inodeIndex int32) int64 {
	// Calcula el desplazamiento en el archivo basado en el índice del inodo
	return int64(sb.S_inode_start) + int64(inodeIndex)*int64(sb.S_inode_size)
}
