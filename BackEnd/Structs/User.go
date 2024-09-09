package structs

import "fmt"

// User define la estructura para los usuarios del sistema
type User struct {
	Id       string
	Group    string
	Name     string
	Password string
	Status   bool // Indica si el usuario está logueado
}

// NewUser crea un nuevo usuario
func NewUser(id, group, name, password string) *User {
	return &User{id, group, name, password, false}
}

// ToString devuelve una representación en cadena del usuario
func (u *User) ToString() string {
	return fmt.Sprintf("%s,U,%s,%s,%s", u.Id, u.Group, u.Name, u.Password)
}

// ToGroupString devuelve una representación del grupo en cadena
func (u *User) ToGroupString() string {
	return fmt.Sprintf("%s,G,%s", u.Id, u.Group)
}
