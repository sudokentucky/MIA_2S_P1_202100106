package structs

import "fmt"

// User define la estructura para los usuarios del sistema
type User struct {
	Id       string // Identificador único del usuario, si es 0 está eliminado
	Tipo     string // Tipo de entidad, en este caso "U" para usuarios
	Group    string // Grupo al que pertenece el usuario
	Name     string // Nombre del usuario
	Password string // Contraseña del usuario
	Status   bool   // Indica si el usuario está activo o eliminado
}

// NewUser crea un nuevo usuario
func NewUser(id, group, name, password string) *User {
	return &User{id, "U", group, name, password, true} // El usuario se crea como activo
}

// ToString devuelve una representación en cadena del usuario
func (u *User) ToString() string {
	return fmt.Sprintf("%s,%s,%s,%s,%s", u.Id, u.Tipo, u.Group, u.Name, u.Password)
}

// Elimina el usuario (cambia el ID a "0" y desactiva el estado)
func (u *User) Eliminar() {
	u.Id = "0"
	u.Status = false
}
