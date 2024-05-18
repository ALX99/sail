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
	ScrollPadding   int      `yaml:"scrollPadding"`
	ShowHiddenFiles bool     `yaml:"showHiddenFiles"`
	Keybinds        Keybinds `yaml:"keybinds"`
}
type Keybinds struct {
	NavUp    string `yaml:"up"`
	NavDown  string `yaml:"down"`
	NavLeft  string `yaml:"left"`
	NavRight string `yaml:"right"`
	NavIn    string `yaml:"in"`
	NavOut   string `yaml:"out"`
	Delete   string `yaml:"delete"`
	Move     string `yaml:"move"`
}

// GetConfig reads, pareses and returns the configuration
func GetConfig() (Config, error) {
	// Sane defaults
	cfg := Config{
		Settings: Settings{
			ScrollPadding:   2,
			ShowHiddenFiles: false,
			Keybinds: Keybinds{
				NavLeft:  "h",
				NavDown:  "j",
				NavUp:    "k",
				NavRight: "l",
				Delete:   "d",
				Move:     "p",
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
		return path.Join(path.Clean(dir), "fly") + "/config.yaml"
	}

	return "~/.config/fly/config.yaml"
}
