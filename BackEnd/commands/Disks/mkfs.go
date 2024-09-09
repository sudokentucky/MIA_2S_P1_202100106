package commands

import (
	structures "backend/Structs"
	global "backend/globals"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"time"
)

// MKFS estructura que representa el comando mkfs con sus parámetros
type MKFS struct {
	id  string // ID del disco
	typ string // Tipo de formato (full)
}

func ParserMkfs(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer // Buffer para capturar los mensajes importantes para el usuario
	cmd := &MKFS{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[^\s]+|-type=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-id":
			if value == "" {
				return "", errors.New("el id no puede estar vacío")
			}
			cmd.id = value
		case "-type":
			if value != "full" {
				return "", errors.New("el tipo debe ser full")
			}
			cmd.typ = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -id")
	}

	if cmd.typ == "" {
		cmd.typ = "full"
	}

	err := commandMkfs(cmd, &outputBuffer)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	return outputBuffer.String(), nil
}

func commandMkfs(mkfs *MKFS, outputBuffer *bytes.Buffer) error {
	fmt.Fprintf(outputBuffer, "========================== MKFS ==========================\n")
	mountedPartition, partitionPath, err := global.GetMountedPartition(mkfs.id)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(partitionPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo el archivo de la partición: %v", err)
	}
	defer file.Close() // Asegura que el archivo se cierre al final de la función

	// Mensaje importante para el usuario
	fmt.Fprintf(outputBuffer, "Partición montada con éxito en %s.\n", partitionPath)

	// Verificar la partición montada (solo en consola para depuración)
	fmt.Println("\nPartición montada:")
	mountedPartition.Print()

	// Calcular el valor de n
	n := calculateN(mountedPartition)

	// Mensaje de depuración
	fmt.Println("\nValor de n:", n)

	// Inicializar un nuevo superbloque
	superBlock := createSuperBlock(mountedPartition, n)

	// Verificar el superbloque (solo en consola)
	fmt.Println("\nSuperBlock:")
	superBlock.Print()

	// Crear los bitmaps
	err = superBlock.CreateBitMaps(file)
	if err != nil {
		return err
	}

	// Mensaje importante para el usuario
	fmt.Fprintln(outputBuffer, "Bitmaps creados correctamente.")

	// Crear archivo users.txt
	err = superBlock.CreateUsersFile(file)
	if err != nil {
		return err
	}

	// Mensaje importante para el usuario
	fmt.Fprintln(outputBuffer, "Archivo users.txt creado correctamente.")

	// Serializar el superbloque
	err = superBlock.Encode(file, int64(mountedPartition.Part_start))
	if err != nil {
		return err
	}

	// Mensaje importante para el usuario
	fmt.Fprintln(outputBuffer, "Superbloque escrito correctamente en el disco.")
	fmt.Fprintln(outputBuffer, "===========================================================")

	return nil
}

func calculateN(partition *structures.Partition) int32 {
	// Calcular el valor de n para el superbloque
	numerator := int(partition.Part_size) - binary.Size(structures.Superblock{})
	denominator := 4 + binary.Size(structures.Inode{}) + 3*binary.Size(structures.FileBlock{}) // Todos los bloques tienen el mismo tamaño
	n := math.Floor(float64(numerator) / float64(denominator))

	// Mensaje de depuración
	fmt.Println("Valor de n calculado:", int32(n))
	return int32(n)
}

func createSuperBlock(partition *structures.Partition, n int32) *structures.Superblock {
	// Calcular punteros de las estructuras
	bm_inode_start := partition.Part_start + int32(binary.Size(structures.Superblock{}))
	bm_block_start := bm_inode_start + n
	inode_start := bm_block_start + (3 * n) // Puntero para los inodos
	block_start := inode_start + (int32(binary.Size(structures.Inode{})) * n)

	// Crear un nuevo superbloque
	superBlock := &structures.Superblock{
		S_filesystem_type:   2,
		S_inodes_count:      0,
		S_blocks_count:      0,
		S_free_inodes_count: int32(n),
		S_free_blocks_count: int32(n * 3),
		S_mtime:             float64(time.Now().Unix()),
		S_umtime:            float64(time.Now().Unix()),
		S_mnt_count:         1,
		S_magic:             0xEF53,
		S_inode_size:        int32(binary.Size(structures.Inode{})),
		S_block_size:        int32(binary.Size(structures.FileBlock{})),
		S_first_ino:         inode_start,
		S_first_blo:         block_start,
		S_bm_inode_start:    bm_inode_start,
		S_bm_block_start:    bm_block_start,
		S_inode_start:       inode_start,
		S_block_start:       block_start,
	}

	// Mensaje de depuración
	fmt.Println("Superbloque creado con éxito.")
	return superBlock
}
