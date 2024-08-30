package structs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	utilidades "ArchivosP1/utils" // Importa el paquete utils
)

// Estructura que representa un MBR
type MBR struct {
	MbrSize          int32        // Tamaño del MBR
	MbrCreacionDate  float32      // Fecha de creación del MBR
	MbrDiskSignature int32        // Número de serie del disco (random)
	MbrDiskFit       [1]byte      // BF = Best Fit, FF = First Fit, WF = Worst Fit
	MbrPartitions    [4]Partition // Particiones del MBR (4 particiones)
}

// Encode serializa la estructura MBR en un archivo
func (mbr *MBR) Encode(file *os.File) error {
	return utilidades.WriteToFile(file, 0, mbr) // Escribe el MBR en el inicio del archivo
}

// Decode deserializa la estructura MBR desde un archivo
func (mbr *MBR) Decode(file *os.File) error {
	return utilidades.ReadFromFile(file, 0, mbr) // Lee el MBR desde el inicio del archivo
}

// Método para obtener la primera partición disponible
func (mbr *MBR) GetFirstAvailablePartition() (*Partition, int, int) {
	// Calcular el offset para el start de la partición
	offset := binary.Size(mbr) // Tamaño del MBR en bytes

	// Recorrer las particiones del MBR
	for i := 0; i < len(mbr.MbrPartitions); i++ {
		if mbr.MbrPartitions[i].Part_start == -1 { // -1 disponible
			return &mbr.MbrPartitions[i], offset, i
		} else {
			// Calcular el nuevo offset para la siguiente partición, es decir, sumar el tamaño de la partición
			offset += int(mbr.MbrPartitions[i].Part_size)
		}
	}
	return nil, -1, -1
}

// Método para obtener una partición por nombre
func (mbr *MBR) GetPartitionByName(name string) (*Partition, int) {
	for i, partition := range mbr.MbrPartitions {
		partitionName := strings.Trim(string(partition.Part_name[:]), "\x00 ")
		inputName := strings.Trim(name, "\x00 ")
		// Si el nombre de la partición coincide, devolver la partición y el índice
		if strings.EqualFold(partitionName, inputName) {
			return &partition, i
		}
	}
	return nil, -1
}

// Función para obtener una partición por ID
func (mbr *MBR) GetPartitionByID(id string) (*Partition, error) {
	for i := 0; i < len(mbr.MbrPartitions); i++ {
		partitionID := strings.Trim(string(mbr.MbrPartitions[i].Part_id[:]), "\x00 ")
		inputID := strings.Trim(id, "\x00 ")
		// Si el nombre de la partición coincide, devolver la partición
		if strings.EqualFold(partitionID, inputID) {
			return &mbr.MbrPartitions[i], nil
		}
	}
	return nil, errors.New("partición no encontrada")
}

// HasExtendedPartition verifica si ya existe una partición extendida en el MBR
func (mbr *MBR) HasExtendedPartition() bool {
	for _, partition := range mbr.MbrPartitions {
		// Verificar si la partición es extendida
		if partition.Part_type[0] == 'E' {
			return true // Devolver verdadero si se encuentra una partición extendida
		}
	}
	return false
}

// CalculateAvailableSpace calcula el espacio disponible en el disco.
func (mbr *MBR) CalculateAvailableSpace() (int32, error) {
	totalSize := mbr.MbrSize
	usedSpace := int32(binary.Size(MBR{})) // Tamaño del MBR

	partitions := mbr.MbrPartitions[:] // Obtener todas las particiones
	for _, part := range partitions {
		if part.Part_size != 0 { // Si la partición está ocupada
			usedSpace += part.Part_size
		}
	}

	if usedSpace >= totalSize {
		return 0, fmt.Errorf("there is no available space on the disk")
	}

	return totalSize - usedSpace, nil
}

// Método para imprimir los valores del MBR
func (mbr *MBR) Print() {
	creationTime := time.Unix(int64(mbr.MbrCreacionDate), 0)
	diskFit := rune(mbr.MbrDiskFit[0])
	fmt.Printf("MBR Size: %d | Creation Date: %s | Disk Signature: %d | Disk Fit: %c\n",
		mbr.MbrSize, creationTime.Format(time.RFC3339), mbr.MbrDiskSignature, diskFit)
}

// Método para imprimir las particiones del MBR
func (mbr *MBR) PrintPartitions() {
	for i, partition := range mbr.MbrPartitions {
		partStatus := rune(partition.Part_status[0])
		partType := rune(partition.Part_type[0])
		partFit := rune(partition.Part_fit[0])
		partName := strings.TrimSpace(string(partition.Part_name[:]))
		partID := strings.TrimSpace(string(partition.Part_id[:]))

		// Imprimir en una sola línea la información de cada partición
		fmt.Printf("Partition %d: Status: %c | Type: %c | Fit: %c | Start: %d | Size: %d | Name: %s | Correlative: %d | ID: %s\n",
			i+1, partStatus, partType, partFit, partition.Part_start, partition.Part_size, partName, partition.Part_correlative, partID)
	}
}
