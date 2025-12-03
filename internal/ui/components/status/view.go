package status

import (
	"context"
	"errors"
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

type dirSize struct {
	size int64
	path string
}

type View struct {
	prevWD string
	sb     strings.Builder
	width  int

	wd filesys.Dir

	errorAt time.Time
	error   error

	dirSize int64

	cancel context.CancelFunc
}

func New() *View {
	return &View{
		sb: strings.Builder{},
	}
}

func (v *View) Init() tea.Cmd {
	// Only calculate dir size if a working directory has been set.
	// Otherwise, SetWD will trigger the calculation when called.
	if v.wd.Path() == "" {
		return nil
	}

	var ctx context.Context
	ctx, v.cancel = context.WithCancel(context.Background())
	wd := v.wd

	return func() tea.Msg {
		return v.performCalcDirSize(ctx, wd)
	}
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
	v.width = max(0, width)
}

// Height returns the rendered height of the status bar with current state.
func (v *View) Height() int {
	return lipgloss.Height(v.View())
}

// SetWD sets a new working directory
func (v *View) SetWD(dir filesys.Dir) tea.Cmd {
	if v.wd.Path() == dir.Path() {
		return nil
	}

	if v.cancel != nil {
		v.cancel()
	}

	var ctx context.Context
	ctx, v.cancel = context.WithCancel(context.Background())

	v.prevWD = v.wd.Path()
	v.wd = dir
	v.dirSize = 0

	return func() tea.Msg {
		return v.performCalcDirSize(ctx, dir)
	}
}

func (v *View) SetError(err error) tea.Cmd {
	v.error = err
	v.errorAt = time.Now()
	return func() tea.Msg {
		time.Sleep(errDur)
		return struct{}{} // dummy message to trigger update
	}
}

func (v *View) performCalcDirSize(ctx context.Context, dir filesys.Dir) tea.Msg {
	size, err := dir.RealSize(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil // Ignore canceled errors
		}
		return err
	}
	return dirSize{size: size, path: dir.Path()}
}
