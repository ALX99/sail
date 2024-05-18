package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alx99/sail/internal/config"
	"github.com/alx99/sail/internal/models"
	"github.com/alx99/sail/internal/util"
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
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	util.SetupLogger()
	log.Info().Msg("Sail started")

	util.SetupStyles()

	// m, err := primary.New(state.NewState(), primary.Config{PWDFile: *pwdFile}, cfg)
	// if err != nil {
	// 	log.Fatal().Err(err).Send()
	// }

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	_, err = tea.NewProgram(models.NewMain(cwd, cfg)).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
