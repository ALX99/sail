package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"

	"github.com/rs/zerolog/log"
)

// SetupLogger sets up the global logger
func SetupLogger(buffered bool) (flush func() (string, error)) {
	fPath := filepath.Join(os.TempDir(), "sail.log")
	f, err := os.Create(fPath)
	if err != nil {
		panic(err)
	}

	var w io.Writer = f
	flush = func() (string, error) { return fPath, nil }

	if buffered {
		w = bufio.NewWriter(f)
		flush = func() (string, error) { return fPath, w.(*bufio.Writer).Flush() }
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: w,
		FormatCaller: func(i any) string {
			return filepath.Base(fmt.Sprintf("%s", i))
		},
		TimeFormat: "15:04:05.999",
	}).
		With().
		Caller().
		Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	switch strings.ToLower("trace") {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
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
	return
}
