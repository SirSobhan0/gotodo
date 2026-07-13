package main

import (
	"fmt"
	"os"

	"github.com/SirSobhan0/gotodo/internal/todo"

	tea "github.com/charmbracelet/bubbletea"
)

const errorRunningProgram = "Error running program: %v\n"

func main() {
	program := tea.NewProgram(todo.InitialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, errorRunningProgram, err)
		os.Exit(1)
	}
}
