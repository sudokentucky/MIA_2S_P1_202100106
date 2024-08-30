package structs

import (
	utilidades "ArchivosP1/utils" // Importa el paquete utils
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

const (
	BlockSize = 64 // Tamaño de los bloques
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

// CreateUsersFile crea el archivo user.txt que contendrá la información de los usuarios
// CreateUsersFile crea el archivo users.txt que contendrá la información de los usuarios
func (sb *Superblock) CreateUsersFile(file *os.File) error {
	// Crear el inodo raíz
	rootInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Serializar el inodo raíz
	err := utilidades.WriteToFile(file, int64(sb.S_first_ino), rootInode)
	if err != nil {
		return fmt.Errorf("failed to encode root inode: %w", err)
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInode(file)
	if err != nil {
		return fmt.Errorf("failed to update inode bitmap: %w", err)
	}

	// Actualizar el superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += int32(binary.Size(Inode{}))

	// Crear el bloque de carpeta raíz
	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},
			{B_name: [12]byte{'-'}, B_inodo: -1},
			{B_name: [12]byte{'-'}, B_inodo: -1},
		},
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(file)
	if err != nil {
		return fmt.Errorf("failed to update block bitmap: %w", err)
	}

	// Codificar el bloque de carpeta raíz
	err = utilidades.WriteToFile(file, int64(sb.S_first_blo), rootBlock)
	if err != nil {
		return fmt.Errorf("failed to encode root block: %w", err)
	}

	// Actualizar el superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += int32(binary.Size(FolderBlock{}))

	// Verificar el inodo raíz
	fmt.Println("\nInodo Raíz:")
	rootInode.Print()

	// Verificar el bloque de carpeta raíz
	fmt.Println("\nBloque de Carpeta Raíz:")
	rootBlock.Print()

	// Crear el archivo /users.txt
	usersText := "1,G,root\n1,U,root,123\n"

	// Deserializar el inodo raíz
	err = utilidades.ReadFromFile(file, int64(sb.S_inode_start+0), rootInode)
	if err != nil {
		return fmt.Errorf("failed to decode root inode: %w", err)
	}

	// Actualizar el inodo raíz
	rootInode.I_atime = float32(time.Now().Unix())

	// Serializar el inodo raíz
	err = utilidades.WriteToFile(file, int64(sb.S_inode_start+0), rootInode)
	if err != nil {
		return fmt.Errorf("failed to re-encode root inode: %w", err)
	}

	// Deserializar el bloque de carpeta raíz
	err = utilidades.ReadFromFile(file, int64(sb.S_block_start+0), rootBlock)
	if err != nil {
		return fmt.Errorf("failed to decode root block: %w", err)
	}

	// Actualizar el bloque de carpeta raíz
	rootBlock.B_content[2] = FolderContent{B_name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: sb.S_inodes_count}

	// Serializar el bloque de carpeta raíz
	err = utilidades.WriteToFile(file, int64(sb.S_block_start+0), rootBlock)
	if err != nil {
		return fmt.Errorf("failed to re-encode root block: %w", err)
	}

	// Crear el inodo para users.txt
	usersInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(usersText)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInode(file)
	if err != nil {
		return fmt.Errorf("failed to update inode bitmap for users.txt: %w", err)
	}

	// Serializar el inodo users.txt
	err = utilidades.WriteToFile(file, int64(sb.S_first_ino), usersInode)
	if err != nil {
		return fmt.Errorf("failed to encode users inode: %w", err)
	}

	// Actualizar el superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += int32(binary.Size(Inode{}))

	// Crear el bloque para users.txt
	usersBlock := &FileBlock{
		B_content: [64]byte{},
	}
	// Copiar el texto de usuarios en el bloque
	copy(usersBlock.B_content[:], usersText)

	// Serializar el bloque de users.txt
	err = utilidades.WriteToFile(file, int64(sb.S_first_blo), usersBlock)
	if err != nil {
		return fmt.Errorf("failed to encode users block: %w", err)
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(file)
	if err != nil {
		return fmt.Errorf("failed to update block bitmap for users.txt: %w", err)
	}

	// Actualizar el superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += int32(binary.Size(FileBlock{}))

	// Verificar el inodo raíz actualizado
	fmt.Println("\nInodo Raíz Actualizado:")
	rootInode.Print()

	// Verificar el bloque de carpeta raíz actualizado
	fmt.Println("\nBloque de Carpeta Raíz Actualizado:")
	rootBlock.Print()

	// Verificar el inodo users.txt
	fmt.Println("\nInodo users.txt:")
	usersInode.Print()

	// Verificar el bloque de users.txt
	fmt.Println("\nBloque de users.txt:")
	usersBlock.Print()

	return nil
}

// UpdateBitmapInode actualiza el Bitmap de inodos
func (sb *Superblock) UpdateBitmapInode(file *os.File) error {
	return utilidades.WriteToFile(file, int64(sb.S_bm_inode_start)+int64(sb.S_inodes_count), []byte{'1'})
}

// UpdateBitmapBlock actualiza el Bitmap de bloques
func (sb *Superblock) UpdateBitmapBlock(file *os.File) error {
	return utilidades.WriteToFile(file, int64(sb.S_bm_block_start)+int64(sb.S_blocks_count), []byte{'X'})
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
				err := utilidades.ReadFromFile(file, int64(sb.S_block_start+(blockIndex*BlockSize)), block)
				if err != nil {
					return fmt.Errorf("failed to decode folder block %d: %w", blockIndex, err)
				}
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
			} else if inode.I_type[0] == '1' {
				block := &FileBlock{}
				err := utilidades.ReadFromFile(file, int64(sb.S_block_start+(blockIndex*BlockSize)), block)
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
