package structs

import (
	"encoding/binary"
	"fmt"
	"os"
)

// EBR representa el Extended Boot Record
type EBR struct {
	Ebr_mount [1]byte  //Indica si la particion esta montada o no
	Ebr_fit   [1]byte  //BF = Best Fit, FF = First Fit, WF = Worst Fit
	Ebr_start int32    //Byte donde inicia la particion
	Ebr_size  int32    //Tamaño de la particion en bytes
	Ebr_next  int32    //Byte donde inicia el siguiente EBR, -1 si no hay siguiente
	Ebr_name  [16]byte //Nombre de la particion
}

// Codificar el EBR en un archivo binario
func (e *EBR) Encode(path string, position int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Seek(position, 0)
	if err != nil {
		return err
	}
	err = binary.Write(file, binary.LittleEndian, e)
	if err != nil {
		return err
	}
	fmt.Printf("coding EBR in position %d with success.\n", position)
	return nil
}

// DEcodificar el EBR desde un archivo binario
func Decode(path string, position int64) (*EBR, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = file.Seek(position, 0)
	if err != nil {
		return nil, err
	}
	ebr := &EBR{}

	err = binary.Read(file, binary.LittleEndian, ebr)
	if err != nil {
		return nil, err
	}

	fmt.Printf("EBR decoded from position %d with success.\n", position)
	return ebr, nil
}

// Establecer los valores del EBR
func (e *EBR) SetEBR(fit byte, size int32, start int32, next int32, name string) {
	fmt.Println("Estableciendo valores del EBR:")
	fmt.Printf("Fit: %c | Size: %d | Start: %d | Next: %d | Name: %s\n", fit, size, start, next, name)

	e.Ebr_mount[0] = '1' // Created
	e.Ebr_fit[0] = fit
	e.Ebr_start = start
	e.Ebr_size = size
	e.Ebr_next = next

	nameBytes := []byte(name)
	if len(nameBytes) > 16 {
		copy(e.Ebr_name[:], nameBytes[:16]) // Truncar si el nombre es mayor a 16 caracteres
	} else {
		copy(e.Ebr_name[:], nameBytes)
	}
}

// ReadEBR lee un EBR desde el disco en una ubicación específica.
func ReadEBR(start int32, diskPath string) (*EBR, error) {
	fmt.Printf("Leyendo EBR desde el disco en la posición: %d\n", start)
	return Decode(diskPath, int64(start))
}

// createAndWriteEBR crea un nuevo EBR y lo escribe en el archivo de disco
func CreateAndWriteEBR(start int32, size int32, fit byte, name string, diskPath string) error {
	fmt.Printf("Creando y escribiendo EBR en la posición: %d\n", start)

	ebr := &EBR{}
	ebr.SetEBR(fit, size, start, -1, name) // Establecer los valores del EBR

	return ebr.Encode(diskPath, int64(start))
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

	if nextStart <= 0 {
		return -1, fmt.Errorf("error calculando la posición del próximo EBR, resultado negativo o cero")
	}
	if nextStart >= extendedPartitionStart+extendedPartitionSize {
		return -1, fmt.Errorf("error: el siguiente EBR está fuera de los límites de la partición extendida")
	}

	fmt.Printf("Inicio del siguiente EBR calculado con éxito: %d\n", nextStart)
	return nextStart, nil
}

// FindLastEBR busca el último EBR en la lista enlazada de EBRs
func FindLastEBR(start int32, diskPath string) (*EBR, error) {
	fmt.Printf("Buscando el último EBR a partir de la posición: %d\n", start)

	currentEBR, err := ReadEBR(start, diskPath)
	if err != nil {
		return nil, err
	}

	for currentEBR.Ebr_next != -1 {
		fmt.Printf("EBR encontrado - Start: %d, Next: %d\n", currentEBR.Ebr_start, currentEBR.Ebr_next)

		nextEBR, err := ReadEBR(currentEBR.Ebr_next, diskPath)
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
