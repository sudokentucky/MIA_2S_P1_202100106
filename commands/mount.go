package commands

import (
	structures "ArchivosP1/Structs"
	globals "ArchivosP1/globals"
	utils "ArchivosP1/utils"
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

// CommandMount parsea el comando mount y devuelve una instancia de MOUNT
func ParserMount(tokens []string) (*Mount, error) {
	cmd := &Mount{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-path":
			if value == "" {
				return nil, errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-name":
			if value == "" {
				return nil, errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return nil, errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return nil, errors.New("faltan parámetros requeridos: -name")
	}

	err := commandMount(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return cmd, nil // Devuelve el comando MOUNT creado
}

func commandMount(mount *Mount) error {
	fmt.Println("========================== MOUNT ==========================")
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
		fmt.Println("Error deserializando el MBR:", err)
		return err
	}

	// Buscar la partición con el nombre especificado
	partition, indexPartition := mbr.GetPartitionByName(mount.name)
	if partition == nil {
		fmt.Println("Error: la partición no existe")
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
		fmt.Println("Error generando el id de partición:", err)
		return err
	}

	// Guardar la partición montada en la lista de montajes globales
	globals.MountedPartitions[idPartition] = mount.path
	partition.MountPartition(indexPartition, idPartition)

	// Guardar la partición modificada en el MBR
	mbr.MbrPartitions[indexPartition] = *partition
	err = mbr.Encode(file)
	if err != nil {
		fmt.Println("Error serializando el MBR:", err)
		return err
	}

	// Mostrar la partición montada y su ID
	fmt.Printf("Partición %s montada correctamente con ID: %s\n", mount.name, idPartition)
	fmt.Println("\n=== Particiones Montadas ===")
	for id, path := range globals.MountedPartitions {
		fmt.Printf("ID: %s | Path: %s\n", id, path)
	}
	fmt.Println("===========================================================")

	return nil
}

func GenerateIdPartition(mount *Mount, indexPartition int) (string, error) {
	lastTwoDigits := globals.Carnet[len(globals.Carnet)-2:]
	letter, err := utils.GetLetter(mount.path)
	if err != nil {
		return "", err
	}

	idPartition := fmt.Sprintf("%s%d%s", lastTwoDigits, indexPartition+1, letter)
	return idPartition, nil
}
