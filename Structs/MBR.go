package structs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Estructura que representa un MBR
//Master Boot Record

type MBR struct {
	mbr_size           int32        //Tamaño del MBR
	mbr_creacion_date  float32      //Fecha de creacion del MBR
	mbr_disk_signature int32        //Numero de serie del disco (random)
	mbr_disk_fit       [1]byte      //BF = Best Fit, FF = First Fit, WF = Worst Fit
	mbr_partitions     [4]Partition //Particiones del MBR (4 particiones)
	//Total de bytes: 92 bytes
}

// Serializar el MBR
func (mbr *MBR) Encode(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	//Escritura de la estructura MBR
	err = binary.Write(file, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	return nil
}

func (mbr *MBR) Decode(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	mbrSize := binary.Size(mbr) // Se obtiene el tamaño de la estructura MBR
	if mbrSize <= 0 {
		return fmt.Errorf("invalid MBR size: %d", mbrSize)
	}

	buffer := make([]byte, mbrSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Decodificar los bytes leídos en la estructura MBR
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	return nil
}

// Método para obtener la primera partición disponible
func (mbr *MBR) GetFirstAvailablePartition() (*Partition, int, int) {
	// Calcular el offset para el start de la partición
	offset := binary.Size(mbr) // Tamaño del MBR en bytes

	// Recorrer las particiones del MBR
	for i := 0; i < len(mbr.mbr_partitions); i++ {
		if mbr.mbr_partitions[i].part_start == -1 { // -1 disponible
			return &mbr.mbr_partitions[i], offset, i
		} else {
			// Calcular el nuevo offset para la siguiente partición, es decir, sumar el tamaño de la partición
			offset += int(mbr.mbr_partitions[i].part_size)
		}
	}
	return nil, -1, -1
}

// Método para obtener una partición por nombre
func (mbr *MBR) GetPartitionByName(name string) (*Partition, int) {
	for i, partition := range mbr.mbr_partitions {
		partitionName := strings.Trim(string(partition.part_name[:]), "\x00 ")
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
	for i := 0; i < len(mbr.mbr_partitions); i++ {
		partitionID := strings.Trim(string(mbr.mbr_partitions[i].part_id[:]), "\x00 ")
		inputID := strings.Trim(id, "\x00 ")
		// Si el nombre de la partición coincide, devolver la partición
		if strings.EqualFold(partitionID, inputID) {
			return &mbr.mbr_partitions[i], nil
		}
	}
	return nil, errors.New("partición no encontrada")
}

// HasExtendedPartition verifica si ya existe una partición extendida en el MBR
func (mbr *MBR) HasExtendedPartition() bool {
	for _, partition := range mbr.mbr_partitions {
		// Verificar si la partición es extendida
		if partition.part_type[0] == 'E' {
			return true // Devolver verdadero si se encuentra una partición extendida
		}
	}
	return false
}

// CalculateAvailableSpace calcula el espacio disponible en el disco.
func (mbr *MBR) CalculateAvailableSpace() (int32, error) {
	totalSize := mbr.mbr_size
	usedSpace := int32(binary.Size(MBR{})) // Tamaño del MBR

	partitions := mbr.mbr_partitions[:] // Obtener todas las particiones
	for _, part := range partitions {
		if part.part_size != 0 { // Si la partición está ocupada
			usedSpace += part.part_size
		}
	}

	if usedSpace >= totalSize {
		return 0, fmt.Errorf("there is no available space on the disk")
	}

	return totalSize - usedSpace, nil
}

// Método para imprimir los valores del MBR
func (mbr *MBR) Print() {
	creationTime := time.Unix(int64(mbr.mbr_creacion_date), 0)
	diskFit := rune(mbr.mbr_disk_fit[0])
	fmt.Printf("MBR Size: %d | Creation Date: %s | Disk Signature: %d | Disk Fit: %c\n",
		mbr.mbr_size, creationTime.Format(time.RFC3339), mbr.mbr_disk_signature, diskFit)
}

// Método para imprimir las particiones del MBR
func (mbr *MBR) PrintPartitions() {
	for i, partition := range mbr.mbr_partitions {
		partStatus := rune(partition.part_status[0])
		partType := rune(partition.part_type[0])
		partFit := rune(partition.part_fit[0])
		partName := strings.TrimSpace(string(partition.part_name[:]))
		partID := strings.TrimSpace(string(partition.part_id[:]))

		// Imprimir en una sola línea la información de cada partición
		fmt.Printf("Partition %d: Status: %c | Type: %c | Fit: %c | Start: %d | Size: %d | Name: %s | Correlative: %d | ID: %s\n",
			i+1, partStatus, partType, partFit, partition.part_start, partition.part_size, partName, partition.part_correlative, partID)
	}
}
