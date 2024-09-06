package commands

import (
	structs "ArchivosP1/Structs"
	globals "ArchivosP1/globals"
	"fmt"
)

// LOGOUT estructura que representa el comando Logout
type LOGOUT struct{}

// ParserLogout inicializa el comando LOGOUT (sin parámetros)
func ParserLogout(tokens []string) (*LOGOUT, error) {
	// El comando Logout no debe recibir parámetros
	if len(tokens) > 1 {
		return nil, fmt.Errorf("el comando Logout no acepta parámetros")
	}

	// Crear una instancia del comando LOGOUT
	cmd := &LOGOUT{}
	err := commandLogout() // Ya no necesitas pasar el parámetro
	if err != nil {
		fmt.Println("Error:", err)
	}
	return cmd, nil
}

// commandLogout ejecuta el comando LOGOUT sin el parámetro innecesario
func commandLogout() error {
	// Verificar si hay una sesión activa
	if globals.UsuarioActual == nil || !globals.UsuarioActual.Status {
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	// Cerrar la sesión
	fmt.Printf("Cerrando sesión de usuario: %s\n", globals.UsuarioActual.Name)

	// Reiniciar la estructura del usuario actual
	globals.UsuarioActual = &structs.User{}

	fmt.Println("Sesión cerrada correctamente.")
	return nil
}
