package structs

import "fmt"

// Group define la estructura para los grupos del sistema
type Group struct {
	GID   string // Identificador único del grupo, si es 0 está eliminado
	Tipo  string // Tipo de entidad, en este caso "G" para grupos
	Group string // Nombre del grupo
}

// NewGroup crea un nuevo grupo
func NewGroup(gid, group string) *Group {
	return &Group{gid, "G", group}
}

// ToString devuelve una representación en cadena del grupo
func (g *Group) ToString() string {
	return fmt.Sprintf("%s,%s,%s", g.GID, g.Tipo, g.Group)
}

// Elimina el grupo (cambia el GID a "0")
func (g *Group) Eliminar() {
	g.GID = "0"
}
