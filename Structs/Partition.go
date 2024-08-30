package structs

import "fmt"

// Estructura que representa una particion
type Partition struct {
	part_status      [1]byte  //indica si la particion esta activa o no
	part_type        [1]byte  //P = Primaria, E = Extendida
	part_fit         [1]byte  //BF = Best Fit, FF = First Fit, WF = Worst Fit
	part_start       int32    //Byte donde inicia la particion
	part_size        int32    //Tamaño de la particion en bytes
	part_name        [16]byte //Nombre de la particion
	part_correlative int32    // inicialmente 0, se asigna cuando se monta la particion
	part_id          [4]byte  // inicialmente vacio, se asigna cuando se monta la particion
	//Total de bytes: 32 bytes
}

// Metodo que crea una particion
func (p *Partition) CreatePartition(partStart, partSize int, partType, partFit, partName string) {
	//Asignamos el valor status de la particion
	p.part_status[0] = '0' //0 = Inactiva, 1 = Activa
	p.part_start = int32(partStart)
	p.part_size = int32(partSize)

	if len(partType) > 0 {
		p.part_type[0] = partType[0]
	}

	if len(partFit) > 0 {
		p.part_fit[0] = partFit[0]
	}

	copy(p.part_name[:], partName)
}

// Metodo que monta una particion
func (p *Partition) MountPartition(correlative int, id string) error {
	p.part_correlative = int32(correlative)
	copy(p.part_id[:], id)
	return nil
}

// Imprimir los valores de la partición en una sola línea
func (p *Partition) Print() {
	fmt.Printf("Status: %c | Type: %c | Fit: %c | Start: %d | Size: %d | Name: %s | Correlative: %d | ID: %s\n",
		p.part_status[0], p.part_type[0], p.part_fit[0], p.part_start, p.part_size,
		string(p.part_name[:]), p.part_correlative, string(p.part_id[:]))
}
