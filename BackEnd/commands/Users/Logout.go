package commands

import (
	structs "backend/Structs"
	globals "backend/globals"
	"bytes"
	"fmt"
)

// LOGOUT estructura que representa el comando Logout
type LOGOUT struct{}

// ParserLogout inicializa el comando LOGOUT (sin parámetros) y captura los mensajes importantes
func ParserLogout(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer // Buffer para capturar los mensajes importantes para el usuario

	// El comando Logout no debe recibir parámetros
	if len(tokens) > 1 {
		return "", fmt.Errorf("el comando Logout no acepta parámetros")
	}

	// Ejecutar el comando logout y capturar los mensajes
	err := commandLogout(&outputBuffer)
	if err != nil {
		fmt.Println("Error:", err) // Mensaje de depuración en consola
		return "", err
	}

	// Retornar los mensajes importantes al frontend
	return outputBuffer.String(), nil
}

// commandLogout ejecuta el comando LOGOUT, y captura los mensajes importantes en un buffer
func commandLogout(outputBuffer *bytes.Buffer) error {
	// Verificar si hay una sesión activa
	if globals.UsuarioActual == nil || !globals.UsuarioActual.Status {
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	// Mensaje importante para el usuario
	fmt.Fprintf(outputBuffer, "Cerrando sesión de usuario: %s\n", globals.UsuarioActual.Name)

	// Cerrar la sesión
	fmt.Printf("Cerrando sesión de usuario: %s\n", globals.UsuarioActual.Name) // Mensaje de depuración

	// Reiniciar la estructura del usuario actual
	globals.UsuarioActual = &structs.User{}

	// Mensaje de éxito importante para el usuario
	fmt.Fprintln(outputBuffer, "Sesión cerrada correctamente.")
	fmt.Println("Sesión cerrada correctamente.") // Mensaje de depuración

	return nil
}
