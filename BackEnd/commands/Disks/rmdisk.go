package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type rmDisk struct {
	path string // Ruta del archivo del disco
}

// ParserRmdisk parsea el comando rmdisk y devuelve una instancia de RMDISK junto con un buffer de salida
func ParserRmdisk(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer // Buffer para capturar la salida

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
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
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
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que el parámetro -path haya sido proporcionado
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// Ejecutar el comando para eliminar el disco y capturar la salida en el buffer
	err := commandRmdisk(cmd, &outputBuffer)
	if err != nil {
		return "", fmt.Errorf("error al eliminar el disco: %v", err)
	}

	// Retornar el contenido del buffer como string
	return outputBuffer.String(), nil
}

func commandRmdisk(rmdisk *rmDisk, outputBuffer *bytes.Buffer) error {
	// Redirigir las salidas de fmt a outputBuffer
	fmt.Fprintln(outputBuffer, "============================= RMDISK ===============================")
	fmt.Fprintf(outputBuffer, "Eliminando disco en %s...\n", rmdisk.path)

	// Verificar si el archivo existe
	if _, err := os.Stat(rmdisk.path); os.IsNotExist(err) {
		return fmt.Errorf("el archivo %s no existe", rmdisk.path)
	}

	// Eliminar el archivo inmediatamente, sin preguntar
	err := os.Remove(rmdisk.path)
	if err != nil {
		return fmt.Errorf("error al eliminar el archivo: %v", err)
	}

	fmt.Fprintf(outputBuffer, "Disco en %s eliminado exitosamente.\n", rmdisk.path)
	fmt.Fprintln(outputBuffer, "========================================================================")
	return nil
}
