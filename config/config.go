package config

import "github.com/gdamore/tcell/v2"

const id = "CFG"

// UIConfigObserver is a function that
// will be called after changes to the UI config
type UIConfigObserver func(UI)

// Global Config variable
var cfg Config
var listeners []UIConfigObserver

// Config holds all the settings for fly
type Config struct {
	UI   UI
	keys []KeyBinding
}

// UI holds all the settings available for the UI
type UI struct {
	Border      bool
	IndentMarks bool
	IndentAll   bool
	DirCandy    bool
	Rainbow     bool
	PDRatio     float64
	WDRatio     float64
	CDRatio     float64
	styles      map[string]tcell.Style
}

// todo read config file
func init() {
	listeners = make([]UIConfigObserver, 0)
	cfg.UI.PDRatio = 1.0
	cfg.UI.WDRatio = 2.0
	cfg.UI.CDRatio = 3.0
	cfg.UI.styles = getStyles()
}

// ReadConfig reads and parses the config from the fs.
// The config can then be retrieved with the GetConfig method
func ReadConfig() error {
	return nil
}

// GetConfig returns the config
func GetConfig() Config {
	// currently this is not thread safe,
	// but that should not matter so far
	return cfg
}

// SetUIConfig sets the UI config
func SetUIConfig(uiConfig UI) {
	cfg.UI = uiConfig
	for _, o := range listeners {
		o(cfg.UI)
	}
}

// AttachConfigObserver will attach an observer
// that'll listen in on changes to the UI config
func AttachConfigObserver(observer UIConfigObserver) {
	listeners = append(listeners, observer)
}
