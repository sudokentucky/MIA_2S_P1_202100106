package structs

import (
	"backend/utils" // Ajusta el path del paquete "utils" según corresponda
	"fmt"
	"os"
	"strings"
)

// Definimos una constante para el tamaño del bloque
const BlockSize = 64

type FileBlock struct {
	B_content [BlockSize]byte
	// Total: 64 bytes
}

// Encode serializa la estructura FileBlock en un archivo binario en la posición especificada
func (fb *FileBlock) Encode(file *os.File, offset int64) error {
	// Utilizamos la función WriteToFile del paquete utils
	err := utils.WriteToFile(file, offset, fb.B_content)
	if err != nil {
		return fmt.Errorf("error writing FileBlock to file: %w", err)
	}
	return nil
}

// Decode deserializa la estructura FileBlock desde un archivo binario en la posición especificada
func (fb *FileBlock) Decode(file *os.File, offset int64) error {
	// Utilizamos la función ReadFromFile del paquete utils
	err := utils.ReadFromFile(file, offset, &fb.B_content)
	if err != nil {
		return fmt.Errorf("error reading FileBlock from file: %w", err)
	}
	return nil
}

// EspacioUsado calcula el espacio usado en el bloque en bytes
func (fb *FileBlock) EspacioUsado() int {
	content := fb.GetContent()
	return len(content)
}

// GetContent retorna el contenido de B_content como una cadena, eliminando bytes nulos al final
func (fb *FileBlock) GetContent() string {
	content := string(fb.B_content[:])
	// Eliminar bytes nulos al final
	content = strings.TrimRight(content, "\x00")
	return content
}

// SetContent copia una cadena en B_content, asegurando que no exceda el tamaño máximo
func (fb *FileBlock) SetContent(content string) error {
	if len(content) > BlockSize {
		return fmt.Errorf("el tamaño del contenido excede el tamaño del bloque de %d bytes", BlockSize)
	}
	// Limpiar B_content
	fb.ClearContent()
	// Copiar el nuevo contenido
	copy(fb.B_content[:], content)
	return nil
}

// EspacioDisponible retorna la cantidad de bytes disponibles en el bloque
func (fb *FileBlock) EspacioDisponible() int {
	return BlockSize - fb.EspacioUsado()
}

// TieneEspacio verifica si aún queda espacio en el bloque
func (fb *FileBlock) TieneEspacio() bool {
	return fb.EspacioDisponible() > 0
}

// Print imprime el contenido de B_content como una cadena
func (fb *FileBlock) Print() {
	fmt.Print(fb.GetContent())
}

// AppendContent agrega contenido al final de B_content si hay espacio disponible
func (fb *FileBlock) AppendContent(content string) error {
	// Verifica el espacio disponible en el bloque
	espacioDisponible := fb.EspacioDisponible()

	// Si no hay suficiente espacio, retorna un error
	if len(content) > espacioDisponible {
		return fmt.Errorf("no hay suficiente espacio para agregar el contenido, se requieren %d bytes pero solo hay %d bytes disponibles", len(content), espacioDisponible)
	}

	// Obtener el espacio actualmente usado en el bloque
	espacioUsado := fb.EspacioUsado()

	// Copia el nuevo contenido en el espacio disponible en una sola operación
	copy(fb.B_content[espacioUsado:], content)

	return nil
}

// ClearContent limpia el contenido de B_content
func (fb *FileBlock) ClearContent() {
	for i := range fb.B_content {
		fb.B_content[i] = 0
	}
}

// NewFileBlock crea un nuevo FileBlock con contenido opcional
func NewFileBlock(content string) (*FileBlock, error) {
	fb := &FileBlock{}
	err := fb.SetContent(content)
	if err != nil {
		return nil, err
	}
	return fb, nil
}

// SplitContent divide una cadena en bloques de tamaño BlockSize y retorna un slice de FileBlocks
func SplitContent(content string) ([]*FileBlock, error) {
	var blocks []*FileBlock
	for len(content) > 0 {
		end := BlockSize
		if len(content) < BlockSize {
			end = len(content)
		}
		fb, err := NewFileBlock(content[:end])
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, fb)
		content = content[end:]
	}
	return blocks, nil
}
