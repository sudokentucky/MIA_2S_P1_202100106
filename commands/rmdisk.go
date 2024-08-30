package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type rmDisk struct {
	path string // Ruta del archivo del disco
}

// ParserRmdisk parsea el comando rmdisk y devuelve una instancia de RMDISK
func ParserRmdisk(tokens []string) (*rmDisk, error) {
	cmd := &rmDisk{} // Crea una nueva instancia de RMDISK

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar el parámetro del comando rmdisk
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+`)
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

		// Remueve comillas del valor si están presentes
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar el parámetro -path
		switch key {
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return nil, errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que el parámetro -path haya sido proporcionado
	if cmd.path == "" {
		return nil, errors.New("faltan parámetros requeridos: -path")
	}

	// Ejecutar el comando para eliminar el disco
	err := commandRmdisk(cmd)
	if err != nil {
		return nil, fmt.Errorf("error al eliminar el disco: %v", err)
	}

	return cmd, nil // Devuelve el comando RMDISK creado
}

func commandRmdisk(rmdisk *rmDisk) error {
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Eliminando disco en %s...\n", rmdisk.path)
	// Verificar si el archivo existe
	if _, err := os.Stat(rmdisk.path); os.IsNotExist(err) {
		return fmt.Errorf("el archivo %s no existe", rmdisk.path)
	}

	// Confirmación de eliminación
	fmt.Printf("¿Estás seguro de que deseas eliminar el disco en %s? (y/n): ", rmdisk.path)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error al leer la entrada: %v", err)
	}

	// Procesar la respuesta
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Operación cancelada.")
		return nil
	}

	// Eliminar el archivo
	err = os.Remove(rmdisk.path)
	if err != nil {
		return fmt.Errorf("error al eliminar el archivo: %v", err)
	}

	fmt.Printf("Disco en %s eliminado exitosamente.\n", rmdisk.path)
	fmt.Println("--------------------------------------------------")
	return nil
}
