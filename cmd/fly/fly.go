package main

import (
	"flag"
	"os"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/util"
	"github.com/alx99/fly/internal/views/model"
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

	// m, err := primary.New(state.NewState(), primary.Config{PWDFile: *pwdFile}, cfg)
	// if err != nil {
	// 	log.Fatal().Err(err).Send()
	// }

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	_, err = tea.NewProgram(model.New(cwd, cfg)).Run()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
