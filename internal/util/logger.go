package util

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
		var mu sync.Mutex
		bufW := bufio.NewWriter(f)

		w = &syncWriter{w: bufW, mu: &mu}

		flush = func() (string, error) {
			mu.Lock()
			defer mu.Unlock()
			return fPath, bufW.Flush()
		}
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

type syncWriter struct {
	w  io.Writer
	mu *sync.Mutex
}

func (sw *syncWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.w.Write(p)
}
