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

var (
	printVersion *bool
	printLastWD  *string

	version = "0.0.0-dev" // set by goreleaser
	isDev   = version == "0.0.0-dev"
)

// init parses the command line flags
func init() {
	printLastWD = flag.String("write-wd", "", "Write the last working directory to the given file")
	printVersion = flag.Bool("version", false, "Print the version")
	flag.Parse()
}

func main() {
	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	cfg.PrintLastWD = *printLastWD

	flush := util.SetupLogger(!isDev)
	defer func() {
		if logPath, err := flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed flushing logs to %q: %v\n", logPath, err)
		}
	}()
	log.Info().Msg("Sail started")

	util.SetupStyles()

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
