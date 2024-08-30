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

// MOUNT estructura que representa el comando mount con sus parámetros
type Mount struct {
	path string
	name string
}

/*
	mount -path=/home/Disco1.mia -name=Part1 #id=341a
	mount -path=/home/Disco2.mia -name=Part1 #id=342a
	mount -path=/home/Disco3.mia -name=Part2 #id=343a
*/

// CommandMount parsea el comando mount y devuelve una instancia de MOUNT
func ParserMount(tokens []string) (*Mount, error) {
	cmd := &Mount{} // Crea una nueva instancia de MOUNT

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando mount
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
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
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return nil, errors.New("el path no puede estar vacío")
			}
			cmd.path = value
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

	// Verifica que los parámetros -path y -name hayan sido proporcionados
	if cmd.path == "" {
		return nil, errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return nil, errors.New("faltan parámetros requeridos: -name")
	}

	// Montamos la partición
	err := commandMount(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return cmd, nil // Devuelve el comando MOUNT creado
}

func commandMount(mount *Mount) error {
	// Abrir el archivo del disco
	file, err := os.OpenFile(mount.path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo el archivo del disco: %v", err)
	}
	defer file.Close()

	// Crear una instancia de MBR
	var mbr structures.MBR

	// Deserializar la estructura MBR desde el archivo
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

	// Generar un id único para la partición
	idPartition, err := GenerateIdPartition(mount, indexPartition)
	if err != nil {
		fmt.Println("Error generando el id de partición:", err)
		return err
	}

	// Guardar la partición montada en la lista de montajes globales
	globals.MountedPartitions[idPartition] = mount.path

	// Modificar la partición para indicar que está montada
	partition.MountPartition(indexPartition, idPartition)

	// Guardar la partición modificada en el MBR
	mbr.MbrPartitions[indexPartition] = *partition

	// Serializar la estructura MBR en el archivo
	err = mbr.Encode(file)
	if err != nil {
		fmt.Println("Error serializando el MBR:", err)
		return err
	}

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
