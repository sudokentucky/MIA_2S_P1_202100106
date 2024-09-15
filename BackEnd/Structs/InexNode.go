package structs

import (
	"backend/utils" // Asegúrate de ajustar el path del package "utils"
	"fmt"
	"os"
	"time"
)

type Inode struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime float32
	I_ctime float32
	I_mtime float32
	I_block [15]int32 // 12 bloques directos, 1 indirecto simple, 1 indirecto doble, 1 indirecto triple
	I_type  [1]byte
	I_perm  [3]byte
	// Total: 88 bytes
}

func (inode *Inode) Encode(file *os.File, offset int64) error {
	// Utilizamos la función WriteToFile del paquete utils
	err := utils.WriteToFile(file, offset, inode)
	if err != nil {
		return fmt.Errorf("error writing Inode to file: %w", err)
	}
	return nil
}

func (inode *Inode) Decode(file *os.File, offset int64) error {
	// Utilizamos la función ReadFromFile del paquete utils
	err := utils.ReadFromFile(file, offset, inode)
	if err != nil {
		return fmt.Errorf("error reading Inode from file: %w", err)
	}
	return nil
}

// Print imprime los atributos del inodo
func (inode *Inode) Print() {
	atime := time.Unix(int64(inode.I_atime), 0)
	ctime := time.Unix(int64(inode.I_ctime), 0)
	mtime := time.Unix(int64(inode.I_mtime), 0)

	fmt.Printf("I_uid: %d\n", inode.I_uid)
	fmt.Printf("I_gid: %d\n", inode.I_gid)
	fmt.Printf("I_size: %d\n", inode.I_size)
	fmt.Printf("I_atime: %s\n", atime.Format(time.RFC3339))
	fmt.Printf("I_ctime: %s\n", ctime.Format(time.RFC3339))
	fmt.Printf("I_mtime: %s\n", mtime.Format(time.RFC3339))
	fmt.Printf("I_block: %v\n", inode.I_block)
	fmt.Printf("I_type: %s\n", string(inode.I_type[:]))
	fmt.Printf("I_perm: %s\n", string(inode.I_perm[:]))
}
