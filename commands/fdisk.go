package commands

import (
	structures "ArchivosP1/Structs"
	utils "ArchivosP1/utils"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Fdisk estructura que representa el comando fdisk con sus parámetros
type Fdisk struct {
	size int    // Tamaño de la partición
	unit string // Unidad de medida del tamaño (K o M)
	fit  string // Tipo de ajuste (BF, FF, WF)
	path string // Ruta del archivo del disco
	typ  string // Tipo de partición (P, E, L)
	name string // Nombre de la partición
}

// ParserFdisk parsea el comando fdisk y devuelve una instancia de Fdisk
func ParserFdisk(tokens []string) (*Fdisk, error) {
	cmd := &Fdisk{} // Crea una nueva instancia de Fdisk

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando fdisk
	re := regexp.MustCompile(`-size=\d+|-unit=[bBkKmM]|-fit=[bBfFwfW]{2}|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+`)
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

		// Elimina las comillas si están presentes
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
			// Verifica que la unidad sea "B", "K" o "M"
			value = strings.ToUpper(value)
			if value != "B" && value != "K" && value != "M" {
				return nil, errors.New("la unidad debe ser B, K o M")
			}
			cmd.unit = value
		case "-fit":
			// Verifica que el ajuste sea "BF", "FF" o "WF"
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return nil, errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return nil, errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-type":
			// Verifica que el tipo sea "P", "E" o "L"
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return nil, errors.New("el tipo debe ser P, E o L")
			}
			cmd.typ = value
		case "-name":
			// Verifica que el nombre no esté vacío
			if value == "" {
				return nil, errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que los parámetros -size, -path y -name hayan sido proporcionados
	if cmd.size == 0 {
		return nil, errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return nil, errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return nil, errors.New("faltan parámetros requeridos: -name")
	}

	// Si no se proporcionó la unidad, se establece por defecto a "K"
	if cmd.unit == "" {
		cmd.unit = "K"
	}

	// Si no se proporcionó el ajuste, se establece por defecto a "WF"
	if cmd.fit == "" {
		cmd.fit = "WF"
	}

	// Si no se proporcionó el tipo, se establece por defecto a "P"
	if cmd.typ == "" {
		cmd.typ = "P"
	}

	// Crear la partición con los parámetros proporcionados
	err := commandFdisk(cmd)
	if err != nil {
		return nil, fmt.Errorf("error al crear la partición: %v", err)
	}

	return cmd, nil // Devuelve el comando Fdisk creado
}

// commandFdisk ejecuta el comando fdisk con los parámetros especificados
func commandFdisk(fdisk *Fdisk) error {
	fmt.Println("======================== FDISK ========================")
	fmt.Println("Creating partition with the following parameters:", fdisk.size, fdisk.unit, fdisk.fit, fdisk.path, fdisk.typ, fdisk.name)

	// Abrir el archivo del disco
	file, err := os.OpenFile(fdisk.path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo el archivo del disco: %v", err)
	}
	defer file.Close()

	// Convertir el tamaño a bytes
	sizeBytes, err := utils.ConvertToBytes(fdisk.size, fdisk.unit)
	if err != nil {
		fmt.Println("Error converting size:", err)
		return err
	}

	if fdisk.typ == "P" {
		// Crear partición primaria
		err = createPrimaryPartition(file, fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creando partición primaria:", err)
			return err
		}
	} else if fdisk.typ == "E" {
		fmt.Println("Creando partición extendida...")
		err = createExtendedPartition(file, fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creando partición extendida:", err)
			return err
		}
	} else if fdisk.typ == "L" {
		fmt.Println("Creando partición lógica...")
		err = createLogicalPartition(file, fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creando partición lógica:", err)
			return err
		}
	}

	fmt.Println("Partición creada exitosamente.")
	fmt.Println("=======================================================")

	return nil
}

// Crear una partición primaria
func createPrimaryPartition(file *os.File, fdisk *Fdisk, sizeBytes int) error {
	fmt.Println("Creating primary partition with size:", fdisk.size, fdisk.unit)
	var mbr structures.MBR

	// Deserializar la estructura MBR desde el archivo
	err := mbr.Decode(file)
	if err != nil {
		fmt.Println("Error deserializando el MBR:", err)
		return err
	}

	// Obtener la primera partición disponible
	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		return fmt.Errorf("no hay particiones disponibles")
	}

	// Calcular el espacio disponible antes de crear la partición
	availableSpace, err := mbr.CalculateAvailableSpace()
	if err != nil {
		return err
	}

	if int32(sizeBytes) > availableSpace {
		return fmt.Errorf("no hay suficiente espacio para la partición primaria")
	}

	// Crear la partición con los parámetros proporcionados
	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)

	// Guardar la partición en el MBR
	mbr.MbrPartitions[indexPartition] = *availablePartition

	// Serializar el MBR actualizado en el archivo del disco
	err = mbr.Encode(file)
	if err != nil {
		fmt.Println("Error actualizando el MBR en el disco:", err)
		return err
	}

	fmt.Println("Partición primaria creada exitosamente.")
	return nil
}

// Crear una partición extendida
func createExtendedPartition(file *os.File, fdisk *Fdisk, sizeBytes int) error {
	fmt.Println("Creating extended partition with size:", fdisk.size, fdisk.unit, fdisk.name)
	var mbr structures.MBR

	// Deserializar la estructura MBR desde el archivo
	err := mbr.Decode(file)
	if err != nil {
		return fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	// Verificar si ya existe una partición extendida
	if mbr.HasExtendedPartition() {
		return fmt.Errorf("ya existe una partición extendida en este disco")
	}

	// Calcular el espacio disponible antes de crear la partición
	availableSpace, err := mbr.CalculateAvailableSpace()
	if err != nil {
		return err
	}

	if int32(sizeBytes) > availableSpace {
		return fmt.Errorf("no hay suficiente espacio para la partición extendida")
	}

	// Obtener la primera partición disponible en el MBR
	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		return fmt.Errorf("no hay espacio disponible en el MBR para una nueva partición")
	}

	// Crear la partición extendida
	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)
	mbr.MbrPartitions[indexPartition] = *availablePartition

	// Crear el primer EBR dentro de la partición extendida
	err = structures.CreateAndWriteEBR(int32(startPartition), 0, fdisk.fit[0], fdisk.name, file)
	if err != nil {
		return fmt.Errorf("error al crear el primer EBR en la partición extendida: %v", err)
	}

	// Serializar el MBR actualizado en el archivo del disco
	err = mbr.Encode(file)
	if err != nil {
		return fmt.Errorf("error al actualizar el MBR en el disco: %v", err)
	}

	fmt.Println("Partición extendida creada exitosamente.")
	return nil
}

func createLogicalPartition(file *os.File, fdisk *Fdisk, sizeBytes int) error {
	fmt.Println("Creating logical partition with size:", fdisk.size, fdisk.unit, fdisk.name)
	var mbr structures.MBR

	// Deserializar el MBR desde el archivo del disco
	err := mbr.Decode(file)
	if err != nil {
		return fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	// Verificar si existe una partición extendida
	var extendedPartition *structures.Partition
	for i := range mbr.MbrPartitions {
		if mbr.MbrPartitions[i].Part_type[0] == 'E' {
			extendedPartition = &mbr.MbrPartitions[i]
			break
		}
	}

	if extendedPartition == nil {
		return fmt.Errorf("no se encontró una partición extendida en el disco")
	}

	// Encontrar el último EBR dentro de la partición extendida
	lastEBR, err := structures.FindLastEBR(extendedPartition.Part_start, file)
	if err != nil {
		return fmt.Errorf("error al buscar el último EBR: %v", err)
	}

	// Verificar si es el primer EBR
	if lastEBR.Ebr_size == 0 {
		fmt.Println("Detectado EBR inicial vacío, asignando tamaño a la nueva partición lógica.")
		lastEBR.Ebr_size = int32(sizeBytes)
		copy(lastEBR.Ebr_name[:], fdisk.name)

		err = lastEBR.Encode(file, int64(lastEBR.Ebr_start))
		if err != nil {
			return fmt.Errorf("error al escribir el primer EBR con la nueva partición lógica: %v", err)
		}

		fmt.Println("Primera partición lógica creada exitosamente.")
		return nil
	}

	// Calcular el inicio del nuevo EBR
	newEBRStart, err := lastEBR.CalculateNextEBRStart(extendedPartition.Part_start, extendedPartition.Part_size)
	if err != nil {
		return fmt.Errorf("error calculando el inicio del nuevo EBR: %v", err)
	}

	// Verificar si hay suficiente espacio para el nuevo EBR
	availableSize := extendedPartition.Part_size - (newEBRStart - extendedPartition.Part_start)
	if availableSize < int32(sizeBytes) {
		return fmt.Errorf("no hay suficiente espacio en la partición extendida para una nueva partición lógica")
	}

	// Crear el nuevo EBR
	newEBR := structures.EBR{}
	newEBR.SetEBR(fdisk.fit[0], int32(sizeBytes), newEBRStart, -1, fdisk.name)

	// Escribir el nuevo EBR en el disco
	err = newEBR.Encode(file, int64(newEBRStart))
	if err != nil {
		return fmt.Errorf("error al escribir el nuevo EBR en el disco: %v", err)
	}

	// Actualizar el último EBR para que apunte al nuevo
	lastEBR.SetNextEBR(newEBRStart)
	err = lastEBR.Encode(file, int64(lastEBR.Ebr_start))
	if err != nil {
		return fmt.Errorf("error al actualizar el EBR anterior: %v", err)
	}

	fmt.Println("Partición lógica creada exitosamente.")
	return nil
}
