package commands

import (
	structures "backend/Structs"
	globals "backend/globals"
	utils "backend/utils"
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Mount struct {
	path string
	name string
}

// ParserMount parsea el comando mount y devuelve una instancia de MOUNT junto con un buffer de salida
func ParserMount(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer
	cmd := &Mount{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
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

		// Switch para manejar diferentes parámetros
		switch key {
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-name":
			if value == "" {
				return "", errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	// Ejecutar el comando mount y capturar los mensajes importantes en el buffer
	err := commandMount(cmd, &outputBuffer)
	if err != nil {
		fmt.Println("Error:", err) // Mensaje de depuración en consola
		return "", err
	}

	// Retornar los mensajes importantes al frontend
	return outputBuffer.String(), nil
}

func commandMount(mount *Mount, outputBuffer *bytes.Buffer) error {
	fmt.Fprintln(outputBuffer, "========================== MOUNT ==========================") // Mensaje importante

	// Abrir el archivo del disco
	file, err := os.OpenFile(mount.path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo el archivo del disco: %v", err)
	}
	defer file.Close()

	// Crear una instancia de MBR
	var mbr structures.MBR
	err = mbr.Decode(file)
	if err != nil {
		fmt.Println("Error deserializando el MBR:", err) // Mensaje de depuración
		return err
	}

	// Buscar la partición con el nombre especificado
	partition, indexPartition := mbr.GetPartitionByName(mount.name)
	if partition == nil {
		fmt.Fprintln(outputBuffer, "Error: la partición no existe") // Mensaje importante
		return errors.New("la partición no existe")
	}

	// Verificar si la partición ya está montada
	for id, mountedPath := range globals.MountedPartitions {
		if mountedPath == mount.path && strings.Contains(id, mount.name) {
			return fmt.Errorf("la partición %s ya está montada", mount.name)
		}
	}

	// Generar un id único para la partición
	idPartition, err := GenerateIdPartition(mount, indexPartition)
	if err != nil {
		fmt.Println("Error generando el id de partición:", err) // Mensaje de depuración
		return err
	}

	// Guardar la partición montada en la lista de montajes globales
	globals.MountedPartitions[idPartition] = mount.path
	partition.MountPartition(indexPartition, idPartition)

	// Guardar la partición modificada en el MBR
	mbr.MbrPartitions[indexPartition] = *partition
	err = mbr.Encode(file)
	if err != nil {
		fmt.Println("Error serializando el MBR:", err) // Mensaje de depuración
		return err
	}

	// Mostrar la partición montada y su ID (Mensajes importantes para el usuario)
	fmt.Fprintf(outputBuffer, "Partición %s montada correctamente con ID: %s\n", mount.name, idPartition)
	fmt.Fprintln(outputBuffer, "\n=== Particiones Montadas ===")
	for id, path := range globals.MountedPartitions {
		fmt.Fprintf(outputBuffer, "ID: %s | Path: %s\n", id, path)
	}
	fmt.Fprintln(outputBuffer, "===========================================================")

	return nil
}

// GenerateIdPartition genera un ID único para la partición montada
func GenerateIdPartition(mount *Mount, indexPartition int) (string, error) {
	lastTwoDigits := globals.Carnet[len(globals.Carnet)-2:]
	letter, err := utils.GetLetter(mount.path)
	if err != nil {
		return "", err
	}

	idPartition := fmt.Sprintf("%s%d%s", lastTwoDigits, indexPartition+1, letter)
	return idPartition, nil
}
