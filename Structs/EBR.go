package structs

import (
	utilidades "ArchivosP1/utils" // Importa el paquete utils
	"fmt"
	"os"
)

// EBR representa el Extended Boot Record
type EBR struct {
	Ebr_mount [1]byte  // Indica si la partición está montada o no
	Ebr_fit   [1]byte  // BF = Best Fit, FF = First Fit, WF = Worst Fit
	Ebr_start int32    // Byte donde inicia la partición
	Ebr_size  int32    // Tamaño de la partición en bytes
	Ebr_next  int32    // Byte donde inicia el siguiente EBR, -1 si no hay siguiente
	Ebr_name  [16]byte // Nombre de la partición
}

// Encode serializa la estructura EBR en un archivo en la posición especificada
func (e *EBR) Encode(file *os.File, position int64) error {
	return utilidades.WriteToFile(file, position, e)
}

// Decode deserializa la estructura EBR desde un archivo en la posición especificada
func Decode(file *os.File, position int64) (*EBR, error) {
	ebr := &EBR{}

	// Verificar que la posición no sea negativa y esté dentro del rango del archivo
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error al obtener información del archivo: %v", err)
	}
	if position < 0 || position >= fileInfo.Size() {
		return nil, fmt.Errorf("posición inválida para EBR: %d", position)
	}

	err = utilidades.ReadFromFile(file, position, ebr)
	if err != nil {
		return nil, err
	}

	fmt.Printf("EBR decoded from position %d with success.\n", position)
	return ebr, nil
}

// SetEBR establece los valores del EBR
func (e *EBR) SetEBR(fit byte, size int32, start int32, next int32, name string) {
	fmt.Println("Estableciendo valores del EBR:")
	fmt.Printf("Fit: %c | Size: %d | Start: %d | Next: %d | Name: %s\n", fit, size, start, next, name)

	e.Ebr_mount[0] = '1' // Created
	e.Ebr_fit[0] = fit
	e.Ebr_start = start
	e.Ebr_size = size
	e.Ebr_next = next

	// Copiar el nombre al array Ebr_name y rellenar el resto con ceros
	copy(e.Ebr_name[:], name)
	for i := len(name); i < len(e.Ebr_name); i++ {
		e.Ebr_name[i] = 0 // Rellenar con ceros
	}
}

// ReadEBR lee un EBR desde el archivo en una ubicación específica
func ReadEBR(start int32, file *os.File) (*EBR, error) {
	fmt.Printf("Leyendo EBR desde el archivo en la posición: %d\n", start)
	return Decode(file, int64(start))
}

// CreateAndWriteEBR crea un nuevo EBR y lo escribe en el archivo de disco
func CreateAndWriteEBR(start int32, size int32, fit byte, name string, file *os.File) error {
	fmt.Printf("Creando y escribiendo EBR en la posición: %d\n", start)

	ebr := &EBR{}
	ebr.SetEBR(fit, size, start, -1, name) // Establecer los valores del EBR

	return ebr.Encode(file, int64(start))
}

// Print imprime los valores del EBR en una sola línea
func (e *EBR) Print() {
	fmt.Printf("Mount: %c | Fit: %c | Start: %d | Size: %d | Next: %d | Name: %s\n",
		e.Ebr_mount[0], e.Ebr_fit[0], e.Ebr_start, e.Ebr_size, e.Ebr_next, string(e.Ebr_name[:]))
}

// CalculateNextEBRStart calcula la posición de inicio del próximo EBR
func (e *EBR) CalculateNextEBRStart(extendedPartitionStart int32, extendedPartitionSize int32) (int32, error) {
	fmt.Printf("Calculando el inicio del siguiente EBR...\nEBR Actual - Start: %d, Size: %d, Next: %d\n",
		e.Ebr_start, e.Ebr_size, e.Ebr_next)

	if e.Ebr_size <= 0 {
		return -1, fmt.Errorf("EBR size is invalid or zero")
	}

	if e.Ebr_start < extendedPartitionStart {
		return -1, fmt.Errorf("EBR start position is invalid")
	}

	nextStart := e.Ebr_start + e.Ebr_size

	// Asegurarse de que nextStart esté dentro del rango de la partición extendida
	if nextStart <= e.Ebr_start || nextStart >= extendedPartitionStart+extendedPartitionSize {
		return -1, fmt.Errorf("error: el siguiente EBR está fuera de los límites de la partición extendida")
	}

	fmt.Printf("Inicio del siguiente EBR calculado con éxito: %d\n", nextStart)
	return nextStart, nil
}

// FindLastEBR busca el último EBR en la lista enlazada de EBRs
func FindLastEBR(start int32, file *os.File) (*EBR, error) {
	fmt.Printf("Buscando el último EBR a partir de la posición: %d\n", start)

	currentEBR, err := ReadEBR(start, file)
	if err != nil {
		return nil, err
	}

	for currentEBR.Ebr_next != -1 {
		if currentEBR.Ebr_next < 0 {
			// Evitar leer una posición negativa
			return currentEBR, nil
		}
		fmt.Printf("EBR encontrado - Start: %d, Next: %d\n", currentEBR.Ebr_start, currentEBR.Ebr_next)

		nextEBR, err := ReadEBR(currentEBR.Ebr_next, file)
		if err != nil {
			return nil, err
		}
		currentEBR = nextEBR
	}

	fmt.Printf("Último EBR encontrado en la posición: %d\n", currentEBR.Ebr_start)
	return currentEBR, nil
}

// SetNextEBR establece el apuntador al siguiente EBR en la lista enlazada de EBRs
func (e *EBR) SetNextEBR(newNext int32) {
	fmt.Printf("Estableciendo el siguiente EBR: Actual Start: %d, Nuevo Next: %d\n", e.Ebr_start, newNext)
	e.Ebr_next = newNext
}
