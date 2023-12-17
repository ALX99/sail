package main

import (
	"flag"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/models/primary"
	"github.com/alx99/fly/internal/state"
	"github.com/alx99/fly/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
)

var pwdFile *string

// init parses the command line flags
func init() {
	pwdFile = flag.String("write-pwd", "", "file to write the last working directory to")
	flag.Parse()
}

func main() {
	util.SetupLogger()
	log.Info().Msg("Fly started")

	util.SetupStyles()
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	m, err := primary.New(state.NewState(), primary.Config{PWDFile: *pwdFile}, cfg)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
