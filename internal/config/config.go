package config

import (
	"cmp"
	"log/slog"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// Config represents all of the configuration options
type Config struct {
	Settings    Settings `yaml:"settings"`
	PrintLastWD string
}

type Settings struct {
	Keymap    Keymap `yaml:"keymap"`
	AltScreen bool   `yaml:"alt_screen"`
	MinimalUI bool   `yaml:"minimal_ui"`
}
type Keymap struct {
	NavUp            string `yaml:"up"`
	NavDown          string `yaml:"down"`
	NavLeft          string `yaml:"left"`
	NavRight         string `yaml:"right"`
	NavHome          string `yaml:"go_home"`
	Delete           string `yaml:"delete"`
	Select           string `yaml:"select"`
	Cut              string `yaml:"cut"`
	Copy             string `yaml:"copy"`
	ToggleAltScreen  string `yaml:"toggle_alt_screen"`
	ToggleParentPane string `yaml:"toggle_parent_pane"`
	ToggleHidden     string `yaml:"toggle_hidden"`
	ToggleMinimalUI  string `yaml:"toggle_minimal_ui"`
}

// GetConfig reads, pareses and returns the configuration
func GetConfig() (Config, error) {
	// Sane defaults
	cfg := Config{
		Settings: Settings{
			Keymap: Keymap{
				NavUp:            "k",
				NavDown:          "j",
				NavLeft:          "h",
				NavRight:         "l",
				NavHome:          "~",
				Delete:           "d",
				Select:           " ",
				Cut:              "x",
				Copy:             "c",
				ToggleAltScreen:  "f",
				ToggleParentPane: "P",
				ToggleHidden:     ".",
				ToggleMinimalUI:  "M",
			},
			AltScreen: true,
			MinimalUI: false,
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
			slog.Warn("No configuration file found, using defaults", "path", cfgFile)
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
