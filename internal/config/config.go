package config

import (
	"cmp"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Config represents all of the configuration options
type Config struct {
	Settings    Settings `yaml:"settings"`
	PrintLastWD string
}

type Settings struct {
	Keymap Keymap `yaml:"keymap"`
}
type Keymap struct {
	NavUp    string `yaml:"up"`
	NavDown  string `yaml:"down"`
	NavLeft  string `yaml:"left"`
	NavRight string `yaml:"right"`
	NavIn    string `yaml:"in"`
	NavOut   string `yaml:"out"`
	NavHome  string `yaml:"go_home"`
	Delete   string `yaml:"delete"`
	Select   string `yaml:"select"`
	Paste    string `yaml:"paste"`
	Copy     string `yaml:"copy"`
}

// GetConfig reads, pareses and returns the configuration
func GetConfig() (Config, error) {
	// Sane defaults
	cfg := Config{
		Settings: Settings{
			Keymap: Keymap{
				NavUp:    "k",
				NavDown:  "j",
				NavLeft:  "h",
				NavRight: "l",
				NavIn:    ".",
				NavOut:   ",",
				NavHome:  "~",
				Delete:   "d",
				Select:   " ",
				Paste:    "p",
				Copy:     "c",
			},
		},
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	cfgFile := configPath(home)
	f, err := os.ReadFile(cfgFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn().Str("path", cfgFile).Msg("No configuration file found, using defaults")
			return cfg, nil
		}
		return Config{}, err
	}

	return cfg, yaml.Unmarshal(f, &cfg)
}

// configPath returns the configuration file location
func configPath(homeDir string) string {
	return filepath.Join(cmp.Or(os.Getenv("XDG_CONFIG_HOME"), filepath.Join(homeDir, ".config")), "sail", "config.yaml")
}
