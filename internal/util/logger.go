package util

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/lmittmann/tint"
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
	handler := tint.NewHandler(w, &tint.Options{
		AddSource: true,
		Level:     getLogLevel("trace"),
	})

	slog.SetDefault(slog.New(handler))

	return
}

// getLogLevel converts a string log level to slog.Level
func getLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "trace", "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
