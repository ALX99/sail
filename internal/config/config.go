package config

import (
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

// Config represents all of the configuration options
type Config struct {
	Settings Settings `yaml:"settings"`
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
}

// GetConfig reads, pareses and returns the configuration
func GetConfig() (Config, error) {
	// Sane defaults
	cfg := Config{
		Settings: Settings{
			Keymap: Keymap{
				NavLeft:  "h",
				NavDown:  "j",
				NavUp:    "k",
				NavRight: "l",
				NavIn:    ".",
				NavOut:   ",",
			},
		},
	}

	cfgLoc := getCfgFileLoc()
	f, err := os.ReadFile(cfgLoc)
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

// getCfgFileLoc returns the configuration file location
func getCfgFileLoc() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir != "" {
		return path.Join(path.Clean(dir), "sail") + "/config.yaml"
	}

	return "~/.config/sail/config.yaml"
}
