package main

import (
	"fmt"
	"os"

	"github.com/alx99/fly/internal/ui/views/mainview"
	"github.com/alx99/fly/internal/util"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	util.SetupLogger()
	util.SetupStyles()
	util.Log.Info().Msg("Fly started")
	mw := mainview.New()

	if err := tea.NewProgram(mw).Start(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
