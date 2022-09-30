package main

import (
	"fmt"
	"os"

	"github.com/alx99/fly/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	fw := ui.NewFileWindow("/")

	if err := tea.NewProgram(fw).Start(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
