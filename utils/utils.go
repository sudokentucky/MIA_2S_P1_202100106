package utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// ConvertToBytes convierte un tamaño y una unidad a bytes
func ConvertToBytes(size int, unit string) (int, error) {
	switch unit {
	case "K":
		return size * 1024, nil // Convierte kilobytes a bytes
	case "M":
		return size * 1024 * 1024, nil // Convierte megabytes a bytes
	default:
		return 0, errors.New("invalid unit") // Devuelve un error si la unidad es inválida
	}
}

// Lista con todo el abecedario
var alphabet = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
	"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

// Mapa para almacenar la asignación de letras a los diferentes paths
var pathToLetter = make(map[string]string)

// Índice para la siguiente letra disponible en el abecedario
var nextLetterIndex = 0

// GetLetter obtiene la letra asignada a un path
func GetLetter(path string) (string, error) {
	if _, exists := pathToLetter[path]; !exists {
		if nextLetterIndex < len(alphabet) {
			pathToLetter[path] = alphabet[nextLetterIndex]
			nextLetterIndex++
		} else {
			return "", errors.New("no hay más letras disponibles para asignar")
		}
	}

	return pathToLetter[path], nil
}

// RemoveLetter elimina la letra asignada a un path
func RemoveLetter(path string) {
	delete(pathToLetter, path)
}

// readFromFile lee datos desde un archivo binario en la posición especificada
func ReadFromFile(file *os.File, offset int64, data interface{}) error {
	_, err := file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to offset %d: %w", offset, err)
	}

	err = binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		return fmt.Errorf("failed to read data from file: %w", err)
	}

	return nil
}

// writeToFile escribe datos a un archivo binario en la posición especificada
func WriteToFile(file *os.File, offset int64, data interface{}) error {
	_, err := file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to offset %d: %w", offset, err)
	}

	err = binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	return nil
}
