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
	ScrollPadding   int  `yaml:"scrollPadding"`
	ShowHiddenFiles bool `yaml:"showHiddenFiles"`
}

// GetConfig reads, pareses and returns the configuration
func GetConfig() (Config, error) {
	cfgLoc := getCfgFileLoc()

	f, err := os.ReadFile(cfgLoc)
	if err != nil {
		if !os.IsNotExist(err) {
			return Config{}, err
		}

		// Safe default if no config found
		return Config{
			Settings: Settings{
				ScrollPadding: 2,
			},
		}, nil
	}

	cfg := Config{}
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
