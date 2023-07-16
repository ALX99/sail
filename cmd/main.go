package main

import (
	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/fs"
	"github.com/alx99/fly/internal/models/primary"
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

	m, err := primary.New(fs.NewBrowser(), cfg)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
