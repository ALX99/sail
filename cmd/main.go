package main

import (
	"fmt"
	"os"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/ui/views/mainview"
	"github.com/alx99/fly/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
)

func main() {
	util.SetupLogger()
	log.Info().Msg("Fly started")

	util.SetupStyles()
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	mw := mainview.New(cfg)
	if err := tea.NewProgram(mw).Start(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
