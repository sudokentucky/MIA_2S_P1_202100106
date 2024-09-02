package commands

import (
	global "ArchivosP1/globals"
	reports "ArchivosP1/reps"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// REP estructura que representa el comando rep con sus parámetros
type REP struct {
	id           string // ID del disco
	path         string // Ruta del archivo del disco
	name         string // Nombre del reporte
	path_file_ls string // Ruta del archivo ls (opcional)
}

func ParserRep(tokens []string) (*REP, error) {
	cmd := &REP{} // Crea una nueva instancia de REP
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[^\s]+|-path="[^"]+"|-path=[^\s]+|-name=[^\s]+|-path_file_ls="[^"]+"|-path_file_ls=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-id":
			if value == "" {
				return nil, errors.New("el id no puede estar vacío")
			}
			cmd.id = value
		case "-path":
			if value == "" {
				return nil, errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-name":
			validNames := []string{"mbr", "disk", "inode", "block", "bm_inode", "bm_block", "sb", "file", "ls"}
			if !contains(validNames, value) {
				return nil, errors.New("nombre inválido, debe ser uno de los siguientes: mbr, disk, inode, block, bm_inode, bm_block, sb, file, ls")
			}
			cmd.name = value
		case "-path_file_ls":
			cmd.path_file_ls = value
		default:
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.id == "" || cmd.path == "" || cmd.name == "" {
		return nil, errors.New("faltan parámetros requeridos: -id, -path, -name")
	}

	// Aquí se puede agregar la lógica para ejecutar el comando rep con los parámetros proporcionados
	err := commandRep(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return cmd, nil // Devuelve el comando REP creado
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func commandRep(rep *REP) error {
	mountedMbr, mountedSb, mountedDiskPath, err := global.GetMountedPartitionRep(rep.id)
	if err != nil {
		return err
	}

	file, err := os.Open(mountedDiskPath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}
	defer file.Close()

	// Switch para manejar diferentes tipos de reportes
	switch rep.name {
	case "mbr":
		// Reporte del MBR
		err = reports.ReportMBR(mountedMbr, rep.path, file)
		if err != nil {
			fmt.Printf("Error generando reporte MBR: %v\n", err)
			return err
		}
	case "disk":
		// Reporte del Disco
		err = reports.ReportDisk(mountedMbr, rep.path, mountedDiskPath)
		if err != nil {
			fmt.Printf("Error generando reporte del disco: %v\n", err)
			return err
		}
	case "inode":
		// Reporte de Inodos
		err = reports.ReportInode(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error generando reporte de inodos: %v\n", err)
			return err
		}
	case "bm_inode":
		// Reporte del Bitmap de Inodos
		err = reports.ReportBMInode(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error generando reporte de bitmap de inodos: %v\n", err)
			return err
		}
	// Agrega más casos para otros tipos de reportes
	default:
		return fmt.Errorf("tipo de reporte no soportado: %s", rep.name)
	}

	return nil
}
