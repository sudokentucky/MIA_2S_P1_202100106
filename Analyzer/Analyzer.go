package analyzer

import (
	commands "ArchivosP1/commands"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// mapCommands define un mapeo entre comandos y funciones correspondientes
var mapCommands = map[string]func([]string) (interface{}, error){
	"mkdisk": func(args []string) (interface{}, error) {
		return commands.ParserMkdisk(args)
	},
	"rmdisk": func(args []string) (interface{}, error) {
		return commands.ParserRmdisk(args)
	},
	"fdisk": func(args []string) (interface{}, error) {
		return commands.ParserFdisk(args)
	},
	"mount": func(args []string) (interface{}, error) {
		return commands.ParserMount(args)
	},
	"mkfs": func(args []string) (interface{}, error) {
		return commands.ParserMkfs(args)
	},
	"help": help,
}

func Analyzer(input string) (interface{}, error) {
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return nil, errors.New("no se proporcionó ningún comando")
	}

	cmdFunc, exists := mapCommands[tokens[0]]
	if !exists {
		if tokens[0] == "clear" {
			return clearTerminal()
		} else if tokens[0] == "exit" {
			os.Exit(0)
		}
		return nil, fmt.Errorf("comando desconocido: %s", tokens[0])
	}

	return cmdFunc(tokens[1:])
}

func help(args []string) (interface{}, error) {
	helpMessage := `
Comandos disponibles:
- mkdisk: Crea un nuevo disco. Ejemplo: mkdisk -size=100 -unit=M -fit=FF -path="/home/user/disco.mia"
- rmdisk: Elimina un disco existente. Ejemplo: rmdisk -path="/home/user/disco.mia"
- fdisk: Maneja las particiones del disco. Ejemplo: fdisk -size=50 -unit=M -path="/home/user/disco.mia" -type=P -name="Part1"
- mount: Monta una partición. Ejemplo: mount -path="/home/user/disco.mia" -name="Part1"
- mkfs: Formatea una partición. Ejemplo: mkfs -id=vd1 -type=full
- clear: Limpia la terminal.
- exit: Sale del programa.
- help: Muestra este mensaje de ayuda.
`
	fmt.Println(helpMessage)
	return nil, nil
}

func clearTerminal() (interface{}, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return nil, errors.New("no se pudo limpiar la terminal")
	}
	return nil, nil
}
