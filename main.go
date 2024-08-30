package main

import (
	analyzer "ArchivosP1/Analyzer"
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type 'help' for a list of available commands.")
	for {
		fmt.Print("$ ") // Prompt del usuario
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		_, err := analyzer.Analyzer(input)
		if err != nil {
			// Muestra el error y continua
			fmt.Println("Error:", err)
			continue
		}
		fmt.Println("Command executed successfully.")
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error al leer:", err)
	}
}
