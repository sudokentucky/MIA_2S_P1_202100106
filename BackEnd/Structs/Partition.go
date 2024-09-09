package structs

import "fmt"

// Estructura que representa una particion
type Partition struct {
	Part_status      [1]byte  //indica si la particion esta activa o no
	Part_type        [1]byte  //P = Primaria, E = Extendida
	Part_fit         [1]byte  //BF = Best Fit, FF = First Fit, WF = Worst Fit
	Part_start       int32    //Byte donde inicia la particion
	Part_size        int32    //Tamaño de la particion en bytes
	Part_name        [16]byte //Nombre de la particion
	Part_correlative int32    // inicialmente 0, se asigna cuando se monta la particion
	Part_id          [4]byte  // inicialmente vacio, se asigna cuando se monta la particion
	//Total de bytes: 32 bytes
}

// Metodo que crea una particion
func (p *Partition) CreatePartition(partStart, partSize int, partType, partFit, partName string) {
	//Asignamos el valor status de la particion
	p.Part_status[0] = '0' //0 = Inactiva, 1 = Activa
	p.Part_start = int32(partStart)
	p.Part_size = int32(partSize)

	if len(partType) > 0 {
		p.Part_type[0] = partType[0]
	}

	if len(partFit) > 0 {
		p.Part_fit[0] = partFit[0]
	}

	copy(p.Part_name[:], partName)
}

// Metodo que monta una particion
func (p *Partition) MountPartition(correlative int, id string) error {
	p.Part_correlative = int32(correlative)
	copy(p.Part_id[:], id)
	return nil
}

// Imprimir los valores de la partición en una sola línea
func (p *Partition) Print() {
	fmt.Printf("Status: %c | Type: %c | Fit: %c | Start: %d | Size: %d | Name: %s | Correlative: %d | ID: %s\n",
		p.Part_status[0], p.Part_type[0], p.Part_fit[0], p.Part_start, p.Part_size,
		string(p.Part_name[:]), p.Part_correlative, string(p.Part_id[:]))
}
