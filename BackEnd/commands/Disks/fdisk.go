package commands

import (
	structures "backend/Structs"
	utils "backend/utils"
	"bytes"
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

// ParserFdisk parsea el comando fdisk y devuelve los mensajes generados
func ParserFdisk(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer
	cmd := &Fdisk{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-size=\d+|-unit=[bBkKmM]|-fit=[bBfFwfW]{2}|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+`)
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
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return "", errors.New("el tamaño debe ser un número entero positivo")
			}
			cmd.size = size
		case "-unit":
			value = strings.ToUpper(value)
			if value != "B" && value != "K" && value != "M" {
				return "", errors.New("la unidad debe ser B, K o M")
			}
			cmd.unit = value
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-type":
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return "", errors.New("el tipo debe ser P, E o L")
			}
			cmd.typ = value
		case "-name":
			if value == "" {
				return "", errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que los parámetros -size, -path y -name hayan sido proporcionados
	if cmd.size == 0 {
		return "", errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	if cmd.unit == "" {
		cmd.unit = "K"
	}

	if cmd.fit == "" {
		cmd.fit = "WF"
	}

	if cmd.typ == "" {
		cmd.typ = "P"
	}

	// Ejecutar el comando fdisk y capturar los mensajes en el buffer
	err := commandFdisk(cmd, &outputBuffer)
	if err != nil {
		return "", fmt.Errorf("error al crear la partición: %v", err)
	}

	// Retornar el contenido del buffer, no el objeto Fdisk
	return outputBuffer.String(), nil
}

// commandFdisk ejecuta el comando fdisk con los parámetros especificados
func commandFdisk(fdisk *Fdisk, outputBuffer *bytes.Buffer) error {
	fmt.Fprintf(outputBuffer, "========================== FDISK ==========================\n")
	fmt.Fprintf(outputBuffer, "Creando partición con nombre '%s' y tamaño %d %s...\n", fdisk.name, fdisk.size, fdisk.unit)
	fmt.Println("Detalles internos de la creación de partición:", fdisk.size, fdisk.unit, fdisk.fit, fdisk.path, fdisk.typ, fdisk.name)

	// Abrir el archivo del disco
	file, err := os.OpenFile(fdisk.path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo el archivo del disco: %v", err)
	}
	defer file.Close()

	sizeBytes, err := utils.ConvertToBytes(fdisk.size, fdisk.unit)
	if err != nil {
		fmt.Println("Error converting size:", err) // Mensaje de depuración
		return err
	}

	if fdisk.typ == "P" {
		err = createPrimaryPartition(file, fdisk, sizeBytes, outputBuffer)
		if err != nil {
			fmt.Println("Error creando partición primaria:", err) // Mensaje de depuración
			return err
		}
	} else if fdisk.typ == "E" {
		fmt.Println("Creando partición extendida...") // Mensaje de depuración
		err = createExtendedPartition(file, fdisk, sizeBytes, outputBuffer)
		if err != nil {
			fmt.Println("Error creando partición extendida:", err) // Mensaje de depuración
			return err
		}
	} else if fdisk.typ == "L" {
		fmt.Println("Creando partición lógica...") // Mensaje de depuración
		err = createLogicalPartition(file, fdisk, sizeBytes, outputBuffer)
		if err != nil {
			fmt.Println("Error creando partición lógica:", err) // Mensaje de depuración
			return err
		}
	}

	fmt.Fprintln(outputBuffer, "Partición creada exitosamente.") // Mensaje importante para el usuario
	fmt.Fprintln(outputBuffer, "===========================================================")
	return nil
}

// Crear una partición primaria
func createPrimaryPartition(file *os.File, fdisk *Fdisk, sizeBytes int, outputBuffer *bytes.Buffer) error {
	fmt.Fprintf(outputBuffer, "Creando partición primaria con tamaño %d %s...\n", fdisk.size, fdisk.unit) // Mensaje importante

	// Proceso interno (mensajes de depuración)
	var mbr structures.MBR
	err := mbr.Decode(file)
	if err != nil {
		fmt.Println("Error deserializando el MBR:", err) // Mensaje de depuración
		return err
	}

	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		return errors.New("no hay espacio disponible en el MBR para una nueva partición")
	}

	availableSpace, err := mbr.CalculateAvailableSpace()
	if err != nil {
		return err
	}

	if int32(sizeBytes) > availableSpace {
		return errors.New("no hay suficiente espacio para la partición primaria")
	}

	// Creación de la partición primaria
	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)
	mbr.MbrPartitions[indexPartition] = *availablePartition

	err = mbr.Encode(file)
	if err != nil {
		fmt.Println("Error actualizando el MBR en el disco:", err) // Mensaje de depuración
		return err
	}

	fmt.Fprintln(outputBuffer, "Partición primaria creada exitosamente.")
	return nil
}

// Crear una partición extendida
func createExtendedPartition(file *os.File, fdisk *Fdisk, sizeBytes int, outputBuffer *bytes.Buffer) error {
	fmt.Fprintf(outputBuffer, "Creando partición extendida con tamaño %d %s...\n", fdisk.size, fdisk.unit)
	var mbr structures.MBR

	// Deserializar la estructura MBR desde el archivo
	err := mbr.Decode(file)
	if err != nil {
		return fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	// Verificar si ya existe una partición extendida
	if mbr.HasExtendedPartition() {
		return errors.New("ya existe una partición extendida en este disco")
	}

	availableSpace, err := mbr.CalculateAvailableSpace()
	if err != nil {
		return err
	}

	if int32(sizeBytes) > availableSpace {
		return errors.New("no hay suficiente espacio para la partición extendida")
	}

	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		return errors.New("no hay espacio disponible en el MBR para una nueva partición")
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

	fmt.Fprintln(outputBuffer, "Partición extendida creada exitosamente.") // Mensaje importante
	return nil
}

// Crear una partición lógica
func createLogicalPartition(file *os.File, fdisk *Fdisk, sizeBytes int, outputBuffer *bytes.Buffer) error {
	fmt.Fprintf(outputBuffer, "Creando partición lógica con tamaño %d %s...\n", fdisk.size, fdisk.unit) // Mensaje importante
	var mbr structures.MBR

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
		return errors.New("no se encontró una partición extendida en el disco")
	}

	lastEBR, err := structures.FindLastEBR(extendedPartition.Part_start, file)
	if err != nil {
		return fmt.Errorf("error al buscar el último EBR: %v", err)
	}

	// Verificar si es el primer EBR
	if lastEBR.Ebr_size == 0 {
		fmt.Println("Detectado EBR inicial vacío, asignando tamaño a la nueva partición lógica.") // Mensaje de depuración
		lastEBR.Ebr_size = int32(sizeBytes)
		copy(lastEBR.Ebr_name[:], fdisk.name)

		err = lastEBR.Encode(file, int64(lastEBR.Ebr_start))
		if err != nil {
			return fmt.Errorf("error al escribir el primer EBR con la nueva partición lógica: %v", err)
		}

		fmt.Fprintln(outputBuffer, "Primera partición lógica creada exitosamente.") // Mensaje importante
		return nil
	}

	// Calcular el inicio del nuevo EBR
	newEBRStart, err := lastEBR.CalculateNextEBRStart(extendedPartition.Part_start, extendedPartition.Part_size)
	if err != nil {
		return fmt.Errorf("error calculando el inicio del nuevo EBR: %v", err)
	}

	availableSize := extendedPartition.Part_size - (newEBRStart - extendedPartition.Part_start)
	if availableSize < int32(sizeBytes) {
		return errors.New("no hay suficiente espacio en la partición extendida para una nueva partición lógica")
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

	fmt.Fprintln(outputBuffer, "Partición lógica creada exitosamente.") // Mensaje importante
	return nil
}
