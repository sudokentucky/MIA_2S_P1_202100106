package commands

import (
	structures "ArchivosP1/Structs"
	utils "ArchivosP1/utils"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Constantes para los valores de unit y fit
const (
	UnitK = "K"
	UnitM = "M"
	FitBF = "BF"
	FitFF = "FF"
	FitWF = "WF"
)

type MkDisk struct {
	size int    // Tamaño del disco
	unit string // Unidad de medida del tamaño (K o M)
	fit  string // Tipo de ajuste (BF, FF, WF)
	path string // Ruta del archivo del disco
}

func ParserMkdisk(tokens []string) (*MkDisk, error) {
	cmd := &MkDisk{}

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando mkdisk
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmM]|-fit=[bBfFwW]{2}|-path="[^"]+"|-path=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-size":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return nil, errors.New("el tamaño debe ser un número entero positivo")
			}
			cmd.size = size
		case "-unit":
			// Verifica que la unidad sea "K" o "M"
			value = strings.ToUpper(value)
			if value != UnitK && value != UnitM {
				return nil, errors.New("la unidad debe ser K o M")
			}
			cmd.unit = value
		case "-fit":
			// Verifica que el ajuste sea "BF", "FF" o "WF"
			value = strings.ToUpper(value)
			if value != FitBF && value != FitFF && value != FitWF {
				return nil, errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return nil, errors.New("el path no puede estar vacío")
			}
			// Asegura que el archivo tenga la extensión .mia
			if !strings.HasSuffix(value, ".mia") {
				return nil, errors.New("el archivo debe tener la extensión .mia")
			}
			cmd.path = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que los parámetros -size y -path hayan sido proporcionados
	if cmd.size == 0 {
		return nil, errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return nil, errors.New("faltan parámetros requeridos: -path")
	}

	// Si no se proporcionó la unidad, se establece por defecto a "M"
	if cmd.unit == "" {
		cmd.unit = UnitM
	}

	// Si no se proporcionó el ajuste, se establece por defecto a "FF"
	if cmd.fit == "" {
		cmd.fit = FitFF
	}

	// Crear el disco con los parámetros proporcionados
	err := commandMkdisk(cmd)
	if err != nil {
		return nil, fmt.Errorf("error al crear el disco: %v", err)
	}

	return cmd, nil // Devuelve el comando MKDISK creado
}

func commandMkdisk(mkdisk *MkDisk) error {
	// Convertir el tamaño a bytes
	fmt.Println("======================== MKDISK ==========================")
	fmt.Println("Creating disk with size:", mkdisk.size, mkdisk.unit)
	sizeBytes, err := utils.ConvertToBytes(mkdisk.size, mkdisk.unit)
	if err != nil {
		fmt.Println("Error converting size:", err)
		return err
	}

	// Crear el disco con el tamaño proporcionado
	err = createDisk(mkdisk, sizeBytes)
	if err != nil {
		fmt.Println("Error creating disk:", err)
		return err
	}

	// Crear el MBR con el tamaño proporcionado
	err = createMBR(mkdisk, sizeBytes)
	if err != nil {
		fmt.Println("Error creating MBR:", err)
		return err
	}

	return nil
}

func createDisk(mkdisk *MkDisk, sizeBytes int) error {
	// Crear las carpetas necesarias
	err := os.MkdirAll(filepath.Dir(mkdisk.path), os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		return err
	}

	// Crear el archivo binario
	file, err := os.Create(mkdisk.path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	// Escribir en el archivo usando un buffer de 1 MB
	buffer := make([]byte, 1024*1024) // Crea un buffer de 1 MB
	for sizeBytes > 0 {
		writeSize := len(buffer)
		if sizeBytes < writeSize {
			writeSize = sizeBytes // Ajusta el tamaño de escritura si es menor que el buffer
		}
		if _, err := file.Write(buffer[:writeSize]); err != nil {
			return err // Devuelve un error si la escritura falla
		}
		sizeBytes -= writeSize // Resta el tamaño escrito del tamaño total
	}
	return nil
}

func createMBR(mkdisk *MkDisk, sizeBytes int) error {
	// Abrir el archivo del disco para escritura
	file, err := os.OpenFile(mkdisk.path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error abriendo el archivo:", err)
		return err
	}
	defer file.Close() // Asegura que el archivo se cierre al final de la función

	// Crear el MBR con los valores proporcionados
	mbr := &structures.MBR{
		MbrSize:          int32(sizeBytes),
		MbrCreacionDate:  float32(time.Now().Unix()),
		MbrDiskSignature: rand.Int31(),
		MbrDiskFit:       [1]byte{mkdisk.fit[0]},
		MbrPartitions: [4]structures.Partition{
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
		},
	}

	// Serializar el MBR en el archivo usando el puntero de archivo `file`
	err = mbr.Encode(file)
	if err != nil {
		fmt.Println("Error serializando el MBR en el archivo:", err)
		return err
	}

	fmt.Println("Disk created successfully:", mkdisk.path)
	// Imprimir el MBR creado
	mbr.Print()
	fmt.Println("===========================================================")

	return nil
}
