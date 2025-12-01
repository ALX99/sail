package status

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/alx99/sail/internal/filesys"
	"github.com/alx99/sail/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var errDur = 1 * time.Second

type (
	dirSize struct {
		size int64
		path string
	}
)

type View struct {
	prevWD string
	sb     strings.Builder
	width  int

	wd    filesys.Dir
	navAt time.Time

	errorAt time.Time
	error   error

	dirSize int64
}

func New() *View {
	return &View{
		sb: strings.Builder{},
	}
}

func (v *View) Init() tea.Cmd {
	return v.calcDirSize
}

func (v *View) Update(msg tea.Msg) {
	switch msg := msg.(type) {
	case dirSize:
		if msg.path == v.wd.Path() {
			v.dirSize = msg.size
		}
	}
}

func (v *View) View() string {
	// Mode Segment
	mode := theme.DefaultTheme.StatusMode.Render(" NORMAL ")

	// Path Segment
	pathStr := v.wd.Path()
	if v.error != nil && time.Now().Before(v.errorAt.Add(errDur)) {
		pathStr = v.error.Error() // Show error in path area temporarily
	}
	path := theme.DefaultTheme.StatusPath.Render(" " + pathStr + " ")

	// Info Segment
	size := v.viewSize()
	counts := v.viewCounts()
	infoStr := fmt.Sprintf(" %s | %s ", counts, size)
	info := theme.DefaultTheme.StatusInfo.Render(infoStr)

	// Spacer to push info to the right
	// Calculate used width
	usedWidth := lipgloss.Width(mode) + lipgloss.Width(path) + lipgloss.Width(info)
	spacerWidth := max(0, v.width-usedWidth)
	spacer := theme.DefaultTheme.StatusBar.Width(spacerWidth).Render("")

	return lipgloss.JoinHorizontal(lipgloss.Top,
		mode,
		path,
		spacer,
		info,
	)
}

func (v *View) viewSize() string {
	// check GB
	if gb := float64(v.dirSize) / math.Pow(1024, 3); gb > 1 {
		return fmt.Sprintf("%06.2f GB", gb)
	}

	// check MB
	if mb := float64(v.dirSize) / math.Pow(1024, 2); mb > 1 {
		return fmt.Sprintf("%06.2f MB", mb)
	}

	// check KB
	if kb := float64(v.dirSize) / 1024; kb > 1 {
		return fmt.Sprintf("%06.2f KB", kb)
	}

	return fmt.Sprintf("%06.2f  B", float64(v.dirSize))
}

func (v *View) viewCounts() string {
	f, d := v.wd.Counts()
	return fmt.Sprintf("%d/%d", f, d)
}

func (v *View) SetWidth(width int) {
	v.width = max(0, width-1)
}

// SetWD sets a new working directory
func (v *View) SetWD(dir filesys.Dir) tea.Cmd {
	if v.wd.Path() == dir.Path() {
		return nil
	}

	v.prevWD = v.wd.Path()
	v.wd = dir
	v.navAt = time.Now()
	v.dirSize = 0

	return v.calcDirSize
}

func (v *View) SetError(err error) tea.Cmd {
	v.error = err
	v.errorAt = time.Now()
	return func() tea.Msg {
		time.Sleep(errDur)
		return struct{}{} // dummy message to trigger update
	}
}

func (v *View) calcDirSize() tea.Msg {
	size, err := v.wd.RealSize()
	if err != nil {
		return nil
	}
	return dirSize{size: size, path: v.wd.Path()}
}
