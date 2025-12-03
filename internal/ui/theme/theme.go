package theme

import "github.com/charmbracelet/lipgloss"

var (
	// Catppuccin Mocha inspired colors
	Rosewater = lipgloss.Color("#f5e0dc")
	Flamingo  = lipgloss.Color("#f2cdcd")
	Pink      = lipgloss.Color("#f5c2e7")
	Mauve     = lipgloss.Color("#cba6f7")
	Red       = lipgloss.Color("#f38ba8")
	Maroon    = lipgloss.Color("#eba0ac")
	Peach     = lipgloss.Color("#fab387")
	Yellow    = lipgloss.Color("#f9e2af")
	Green     = lipgloss.Color("#a6e3a1")
	Teal      = lipgloss.Color("#94e2d5")
	Sky       = lipgloss.Color("#89dceb")
	Sapphire  = lipgloss.Color("#74c7ec")
	Blue      = lipgloss.Color("#89b4fa")
	Lavender  = lipgloss.Color("#b4befe")
	Text      = lipgloss.Color("#cdd6f4")
	Subtext1  = lipgloss.Color("#bac2de")
	Subtext0  = lipgloss.Color("#a6adc8")
	Overlay2  = lipgloss.Color("#9399b2")
	Overlay1  = lipgloss.Color("#7f849c")
	Overlay0  = lipgloss.Color("#6c7086")
	Surface2  = lipgloss.Color("#585b70")
	Surface1  = lipgloss.Color("#45475a")
	Surface0  = lipgloss.Color("#313244")
	Base      = lipgloss.Color("#1e1e2e")
	Mantle    = lipgloss.Color("#181825")
	Crust     = lipgloss.Color("#11111b")
)

type Theme struct {
	PrimaryBorder  lipgloss.Style
	ActiveBorder   lipgloss.Style
	InactiveBorder lipgloss.Style
	MinimalDivider lipgloss.Style

	SelectedFile lipgloss.Style
	Cursor       lipgloss.Style

	StatusBar  lipgloss.Style
	StatusMode lipgloss.Style
	StatusPath lipgloss.Style
	StatusInfo lipgloss.Style
}

var DefaultTheme = Theme{
	ActiveBorder:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Blue),
	InactiveBorder: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Overlay0),
	MinimalDivider: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).BorderBottom(false).BorderLeft(false).BorderForeground(Overlay0),

	SelectedFile: lipgloss.NewStyle().Foreground(Sky).Bold(true),
	Cursor:       lipgloss.NewStyle().Background(Surface2).Foreground(Text),

	StatusBar:  lipgloss.NewStyle().Background(Surface0).Foreground(Text),
	StatusMode: lipgloss.NewStyle().Background(Blue).Foreground(Base).Bold(true).Padding(0, 1),
	StatusPath: lipgloss.NewStyle().Background(Surface1).Foreground(Text),
	StatusInfo: lipgloss.NewStyle().
		Background(Surface2).
		Foreground(Subtext0).
		Padding(0, 1),
}
