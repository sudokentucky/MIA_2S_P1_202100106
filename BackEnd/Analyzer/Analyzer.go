package analyzer

import (
	commands "backend/commands"
	Disks "backend/commands/Disks"
	Users "backend/commands/Users"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// mapCommands define un mapeo entre comandos y funciones correspondientes
var mapCommands = map[string]func([]string) (string, error){ // Cambiamos a (string, error)
	"mkdisk": func(args []string) (string, error) {
		result, err := Disks.ParserMkdisk(args)
		return fmt.Sprintf("%v", result), err // Aseguramos que se devuelva un string
	},
	"rmdisk": func(args []string) (string, error) {
		result, err := Disks.ParserRmdisk(args)
		return fmt.Sprintf("%v", result), err
	},
	"fdisk": func(args []string) (string, error) {
		result, err := Disks.ParserFdisk(args)
		return fmt.Sprintf("%v", result), err
	},
	"mount": func(args []string) (string, error) {
		result, err := Disks.ParserMount(args)
		return fmt.Sprintf("%v", result), err
	},
	"mkfs": func(args []string) (string, error) {
		result, err := Disks.ParserMkfs(args)
		return fmt.Sprintf("%v", result), err
	},
	"rep": func(args []string) (string, error) {
		result, err := commands.ParserRep(args)
		return fmt.Sprintf("%v", result), err
	},
	"login": func(args []string) (string, error) {
		result, err := Users.ParserLogin(args)
		return fmt.Sprintf("%v", result), err
	},
	"logout": func(args []string) (string, error) {
		result, err := Users.ParserLogout(args)
		return fmt.Sprintf("%v", result), err
	},
	"mkgrp": func(args []string) (string, error) {
		result, err := Users.ParserMkgrp(args)
		return fmt.Sprintf("%v", result), err
	},
	"rmgrp": func(args []string) (string, error) {
		result, err := Users.ParserRmgrp(args)
		return fmt.Sprintf("%v", result), err
	},
	"mkusr": func(args []string) (string, error) {
		result, err := Users.ParserMkusr(args)
		return fmt.Sprintf("%v", result), err
	},
	"rmusr": func(args []string) (string, error) {
		result, err := Users.ParserRmusr(args)
		return fmt.Sprintf("%v", result), err
	},
	"chgrp": func(args []string) (string, error) {
		result, err := Users.ParserChgrp(args)
		return fmt.Sprintf("%v", result), err
	},
	"mkfile": func(args []string) (string, error) {
		result, err := commands.ParserMkfile(args)
		return fmt.Sprintf("%v", result), err
	},
	"mkdir": func(args []string) (string, error) {
		result, err := commands.ParserMkdir(args)
		return fmt.Sprintf("%v", result), err
	},
	"help": help,
}

func Analyzer(input string) (string, error) {
	// Verificar si es un comentario
	if strings.HasPrefix(strings.TrimSpace(input), "#") {
		// Retornamos el comentario sin procesarlo
		return fmt.Sprintf("Comentario detectado: %s", input), nil
	}

	// Separar el input en tokens
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return "", errors.New("no se proporcionó ningún comando")
	}

	// Verificar si el comando existe en el map
	cmdFunc, exists := mapCommands[tokens[0]]
	if !exists {
		if tokens[0] == "clear" {
			return clearTerminal()
		} else if tokens[0] == "exit" {
			os.Exit(0)
		}
		return "", fmt.Errorf("comando desconocido: %s", tokens[0])
	}

	// Ejecutar la función correspondiente
	return cmdFunc(tokens[1:])
}

func help(args []string) (string, error) {
	helpMessage := `
Comandos disponibles:
- mkdisk: Crea un nuevo disco. Ejemplo: mkdisk -size=100 -unit=M -fit=FF -path="/home/user/disco.mia"
- rmdisk: Elimina un disco existente. Ejemplo: rmdisk -path="/home/user/disco.mia"
- fdisk: Maneja las particiones del disco. Ejemplo: fdisk -size=50 -unit=M -path="/home/user/disco.mia" -type=P -name="Part1"
- mount: Monta una partición. Ejemplo: mount -path="/home/user/disco.mia" -name="Part1"
- mkfs: Formatea una partición. Ejemplo: mkfs -id=vd1 -type=full
- login: Inicia sesión en el sistema. Ejemplo: login -user=admin -pass=1234 -id=vd1
- logout: Cierra la sesión actual. Ejemplo: logout
- mkgrp: Crea un nuevo grupo. Ejemplo: mkgrp -name=users
- rmgrp: Elimina un grupo existente. Ejemplo: rmgrp -name=users
- mkusr: Crea un nuevo usuario. Ejemplo: mkusr -user=user1 -pass=user -grp=users
- rmusr: Elimina un usuario existente. Ejemplo: rmusr -user=user1
- chgrp: Cambia el grupo de un usuario. Ejemplo: chgrp -user=user1 -grp=users
- rep: Genera reportes. Ejemplo: rep -id=vd1 -path="/home/user/disco.mia" -name=mbr
- clear: Limpia la terminal.
- exit: Sale del programa.
- help: Muestra este mensaje de ayuda.

`
	return helpMessage, nil
}

func clearTerminal() (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return "", errors.New("no se pudo limpiar la terminal")
	}
	return "Terminal limpiada", nil
}
