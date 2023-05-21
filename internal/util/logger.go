package util

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"

	"github.com/rs/zerolog/log"
)

// SetupLogger sets up the global logger
func SetupLogger() {
	fPath := path.Join(os.TempDir(), "fly.log")
	f, err := os.Create(fPath)
	if err != nil {
		panic(err)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: f,
		FormatCaller: func(i interface{}) string {
			return filepath.Base(fmt.Sprintf("%s", i))
		},
	}).
		With().
		Caller().
		Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	switch strings.ToLower("debug") {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "info":
		fallthrough
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
