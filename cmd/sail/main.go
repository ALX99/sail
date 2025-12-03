package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/alx99/sail/internal/config"
	"github.com/alx99/sail/internal/style"
	"github.com/alx99/sail/internal/ui/app"
	"github.com/alx99/sail/internal/util"
	tea "github.com/charmbracelet/bubbletea"
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

	flush := util.SetupLogger(!isDev)
	defer func() {
		if logPath, err := flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed flushing logs to %q: %v\n", logPath, err)
		}
	}()
	slog.Info("Sail started")

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	cfg.PrintLastWD = *printLastWD

	styles := style.NewStyles(os.Getenv("LS_COLORS"))

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	var opts []tea.ProgramOption
	if cfg.Settings.AltScreen {
		opts = append(opts, tea.WithAltScreen())
	}

	_, err = tea.NewProgram(app.New(cwd, cfg, styles), opts...).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
