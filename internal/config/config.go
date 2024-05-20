package config

import (
	"cmp"
	"os"
	"path"

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
			},
		},
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	f, err := os.ReadFile(configPath(home))
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, err
	}

	if err = yaml.Unmarshal(f, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// configPath returns the configuration file location
func configPath(homeDir string) string {
	return path.Join(cmp.Or(os.Getenv("XDG_CONFIG_HOME"), path.Join(homeDir, ".config")), "sail", "config.yaml")
}
