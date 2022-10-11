package util

import (
	"os"
	"path"

	"github.com/rs/zerolog"
)

// Log is the global logging instance
var Log zerolog.Logger

// SetupLogger sets up the global logger
func SetupLogger() {
	fPath := path.Join(os.TempDir(), "fly.log")
	f, err := os.Create(fPath)
	if err != nil {
		panic(err)
	}

	Log = zerolog.New(f)
}
