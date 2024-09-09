package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MKFILE : Estructura para el comando MKFILE
type MKFILE struct {
	Path string
	Size int
	Cont string
	R    bool
}

// ParserMkfile : Parseo de argumentos para el comando mkfile
func ParserMkfile(tokens []string) (*MKFILE, error) {
	cmd := &MKFILE{Size: 0, R: false} // Valores predeterminados

	// Expresión regular para encontrar los parámetros
	rePath := regexp.MustCompile(`-path="[^\"]+"|-path=[^\s]+`)
	reSize := regexp.MustCompile(`-size=[^\s]+`)
	reCont := regexp.MustCompile(`-cont="[^\"]+"|-cont=[^\s]+`)
	reR := regexp.MustCompile(`-r`)

	// Extraer los valores de los parámetros
	matchesPath := rePath.FindString(strings.Join(tokens, " "))
	matchesSize := reSize.FindString(strings.Join(tokens, " "))
	matchesCont := reCont.FindString(strings.Join(tokens, " "))
	matchesR := reR.FindString(strings.Join(tokens, " "))

	// Verificar que se proporcione el parámetro -path
	if matchesPath == "" {
		return nil, fmt.Errorf("falta el parámetro -path")
	}
	cmd.Path = strings.SplitN(matchesPath, "=", 2)[1]

	// Verificar el parámetro -size
	if matchesSize != "" {
		size, err := strconv.Atoi(strings.SplitN(matchesSize, "=", 2)[1])
		if err != nil || size < 0 {
			return nil, fmt.Errorf("el parámetro -size debe ser un entero positivo")
		}
		cmd.Size = size
	}

	// Verificar el parámetro -cont
	if matchesCont != "" {
		cmd.Cont = strings.SplitN(matchesCont, "=", 2)[1]
	}

	// Verificar el parámetro -r
	if matchesR != "" {
		cmd.R = true
	}

	// Ejecutar la lógica del comando mkfile
	err := commandMkfile(cmd)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

// commandMkfile : Ejecuta el comando MKFILE
func commandMkfile(mkfile *MKFILE) error {
	// Verificar si hay una sesión activa y si el usuario tiene permisos
	return nil
}
