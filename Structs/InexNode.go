package structs

import (
	"fmt"
	"os"
	"time"

	utilidades "ArchivosP1/utils" // Importa el paquete utils
)

// Inode representa un nodo de índice en el sistema de archivos
type Inode struct {
	I_uid   int32     // ID del usuario propietario del archivo o carpeta
	I_gid   int32     // ID del grupo al que pertenece el archivo o carpeta
	I_size  int32     // Tamaño del archivo en bytes
	I_atime float32   // Última fecha que se leyó el inodo sin modificarlo "02/01/2006 15:04"
	I_ctime float32   // Fecha en que se creó el inodo "02/01/2006 15:04"
	I_mtime float32   // Última fecha en la que se modifica el inodo "02/01/2006 15:04"
	I_block [15]int32 // -1 si no están usados. Los valores del arreglo son: primeros 12 -> bloques directo; 13 -> bloque simple indirecto; 14 -> bloque doble indirecto; 15 -> bloque triple indirecto
	I_type  [1]byte   // 1 -> archivo; 0 -> carpeta
	I_perm  [3]byte   // Permisos del usuario o carpeta
}

// Encode serializa la estructura Inode en un archivo en la posición especificada
func (inode *Inode) Encode(file *os.File, offset int64) error {
	return utilidades.WriteToFile(file, offset, inode)
}

// Decode deserializa la estructura Inode desde un archivo en la posición especificada
func (inode *Inode) Decode(file *os.File, offset int64) error {
	return utilidades.ReadFromFile(file, offset, inode)
}

// Print imprime los atributos del inodo
func (inode *Inode) Print() {
	atime := time.Unix(int64(inode.I_atime), 0)
	ctime := time.Unix(int64(inode.I_ctime), 0)
	mtime := time.Unix(int64(inode.I_mtime), 0)

	fmt.Println("┌─────────┬──────────────────────────┐")
	fmt.Printf("│ %-7s │ %-24d │\n", "I_uid", inode.I_uid)
	fmt.Println("├─────────┼──────────────────────────┤")
	fmt.Printf("│ %-7s │ %-24d │\n", "I_gid", inode.I_gid)
	fmt.Println("├─────────┼──────────────────────────┤")
	fmt.Printf("│ %-7s │ %-24d │\n", "I_size", inode.I_size)
	fmt.Println("├─────────┼──────────────────────────┤")
	fmt.Printf("│ %-7s │ %-24s │\n", "I_atime", atime.Format(time.RFC3339))
	fmt.Println("├─────────┼──────────────────────────┤")
	fmt.Printf("│ %-7s │ %-24s │\n", "I_ctime", ctime.Format(time.RFC3339))
	fmt.Println("├─────────┼──────────────────────────┤")
	fmt.Printf("│ %-7s │ %-24s │\n", "I_mtime", mtime.Format(time.RFC3339))
	fmt.Println("├─────────┼──────────────────────────┤")
	fmt.Printf("│ %-7s │ %-24v │\n", "I_block", inode.I_block)
	fmt.Println("├─────────┼──────────────────────────┤")
	fmt.Printf("│ %-7s │ %-24s │\n", "I_type", string(inode.I_type[:]))
	fmt.Println("├─────────┼──────────────────────────┤")
	fmt.Printf("│ %-7s │ %-24s │\n", "I_perm", string(inode.I_perm[:]))
	fmt.Println("└─────────┴──────────────────────────┘")
}
