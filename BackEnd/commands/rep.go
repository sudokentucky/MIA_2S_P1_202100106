package commands

import (
	global "backend/globals"
	reports "backend/reps"
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// REP estructura que representa el comando rep con sus parámetros
type REP struct {
	id           string // ID del disco
	path         string // Ruta del archivo de salida
	name         string // Nombre del reporte
	path_file_ls string // Ruta del archivo ls (opcional)
}

func ParserRep(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer // Buffer para capturar los mensajes importantes

	cmd := &REP{} // Crea una nueva instancia de REP
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[^\s]+|-path="[^"]+"|-path=[^\s]+|-name=[^\s]+|-path_file_ls="[^"]+"|-path_file_ls=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-id":
			if value == "" {
				return "", errors.New("el id no puede estar vacío")
			}
			cmd.id = value
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-name":
			validNames := []string{"mbr", "disk", "inode", "block", "bm_inode", "bm_block", "sb", "file", "ls"}
			if !contains(validNames, value) {
				return "", errors.New("nombre inválido, debe ser uno de los siguientes: mbr, disk, inode, block, bm_inode, bm_block, sb, file, ls")
			}
			cmd.name = value
		case "-path_file_ls":
			cmd.path_file_ls = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.id == "" || cmd.path == "" || cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -id, -path, -name")
	}

	// Ejecutar el comando y capturar mensajes
	err := commandRep(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	return outputBuffer.String(), nil // Retorna los mensajes importantes al frontend
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func commandRep(rep *REP, outputBuffer *bytes.Buffer) error {
	// Obtener la partición montada
	mountedMbr, mountedSb, mountedDiskPath, err := global.GetMountedPartitionRep(rep.id)
	if err != nil {
		return err
	}

	file, err := os.Open(mountedDiskPath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}
	defer file.Close()

	// Mensaje de inicio de generación de reporte
	fmt.Fprintf(outputBuffer, "Generando reporte '%s'...\n", rep.name)
	fmt.Printf("Generando reporte '%s'...\n", rep.name) // Mensaje de depuración

	// Switch para manejar diferentes tipos de reportes
	switch rep.name {
	case "mbr":
		// Reporte del MBR
		err = reports.ReportMBR(mountedMbr, rep.path, file)
		if err != nil {
			fmt.Fprintf(outputBuffer, "Error generando reporte MBR: %v\n", err)
			fmt.Printf("Error generando reporte MBR: %v\n", err) // Depuración
			return err
		}
	case "disk":
		// Reporte del Disco
		err = reports.ReportDisk(mountedMbr, rep.path, mountedDiskPath)
		if err != nil {
			fmt.Fprintf(outputBuffer, "Error generando reporte del disco: %v\n", err)
			fmt.Printf("Error generando reporte del disco: %v\n", err) // Depuración
			return err
		}
	case "inode":
		// Reporte de Inodos
		err = reports.ReportInode(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Fprintf(outputBuffer, "Error generando reporte de inodos: %v\n", err)
			fmt.Printf("Error generando reporte de inodos: %v\n", err) // Depuración
			return err
		}
	case "block":
		// Reporte de Bloques
		err = reports.ReportBlock(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Fprintf(outputBuffer, "Error generando reporte de bloques: %v\n", err)
			fmt.Printf("Error generando reporte de bloques: %v\n", err) // Depuración
			return err
		}
	case "bm_inode":
		// Reporte del Bitmap de Inodos
		err = reports.ReportBMInode(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Fprintf(outputBuffer, "Error generando reporte de bitmap de inodos: %v\n", err)
			fmt.Printf("Error generando reporte de bitmap de inodos: %v\n", err) // Depuración
			return err
		}
	case "bm_block":
		// Reporte del Bitmap de Bloques
		err = reports.ReportBMBlock(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Fprintf(outputBuffer, "Error generando reporte de bitmap de bloques: %v\n", err)
			fmt.Printf("Error generando reporte de bitmap de bloques: %v\n", err) // Depuración
			return err
		}
	case "sb":
		// Reporte del Superbloque
		err = reports.ReportSuperblock(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Fprintf(outputBuffer, "Error generando reporte del superbloque: %v\n", err)
			fmt.Printf("Error generando reporte del superbloque: %v\n", err) // Depuración
			return err
		}
	case "file":
		// Reporte de Archivo
		err = reports.ReportFile(mountedSb, mountedDiskPath, rep.path, rep.path_file_ls)
		if err != nil {
			fmt.Fprintf(outputBuffer, "Error generando reporte de archivo: %v\n", err)
			fmt.Printf("Error generando reporte de archivo: %v\n", err) // Depuración
			return err
		}
	// Agrega más casos para otros tipos de reportes
	default:
		return fmt.Errorf("tipo de reporte no soportado: %s", rep.name)
	}

	// Mensaje de éxito en la generación de reporte
	fmt.Fprintf(outputBuffer, "Reporte '%s' generado exitosamente en la ruta: %s\n", rep.name, rep.path)
	fmt.Printf("Reporte '%s' generado exitosamente en la ruta: %s\n", rep.name, rep.path) // Depuración

	return nil
}
