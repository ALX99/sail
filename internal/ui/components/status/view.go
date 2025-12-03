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

const sizeAnimInterval = 120 * time.Millisecond

var sizeAnimFrames = []string{
	"Ooo.oo ??",
	"oOo.oo ??",
	"ooO.oo ??",
	"ooo.Oo ??",
	"ooo.oO ??",
	"ooo.Oo ??",
	"ooO.oo ??",
	"oOo.oo ??",
}

type dirSize struct {
	size int64
	path string
}

type sizeAnimTick struct {
	seq int
}

type View struct {
	prevWD string
	sb     strings.Builder
	width  int

	wd filesys.Dir

	errorAt time.Time
	error   error

	dirSize  int64
	selIdx   int
	selTotal int

	cancel context.CancelFunc

	animIdx     int
	animRunning bool
	animSeq     int
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

	v.animRunning = true
	v.animSeq++

	return tea.Batch(v.sizeTick(), func() tea.Msg {
		return v.performCalcDirSize(ctx, wd)
	})
}

func (v *View) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case dirSize:
		if msg.path == v.wd.Path() {
			v.dirSize = msg.size
			v.animRunning = false
		}
	case sizeAnimTick:
		if msg.seq != v.animSeq || !v.animRunning {
			return nil
		}
		v.animIdx = (v.animIdx + 1) % len(sizeAnimFrames)
		return v.sizeTick()
	case error:
		// If size calculation returned an error, stop animating.
		v.animRunning = false
	}

	return nil
}

func (v *View) View() string {
	pathStr := v.wd.Path()
	if v.error != nil && time.Now().Before(v.errorAt.Add(errDur)) {
		pathStr = v.error.Error() // Show error in path area temporarily
	}

	// Left cluster: mode + path styled as connected pills
	left := renderLeadingPills(
		"NORMAL",
		pathStr,
		theme.Blue,
		theme.Surface1,
		theme.Base,
		theme.Surface0,
	)

	// Info Segments (rendered as a joined pill: left + right)
	sizeText := v.viewSize()
	if v.animRunning {
		sizeText = sizeAnimFrames[v.animIdx]
	}
	info := renderConnectedPills(
		sizeText,
		v.viewSelection(),
		theme.Sapphire,
		theme.Mauve,
		theme.Base,
		theme.Surface0,
	)

	// Spacer to push info to the right
	// Calculate used width
	usedWidth := lipgloss.Width(left) + lipgloss.Width(info)
	spacerWidth := max(0, v.width-usedWidth)
	spacer := theme.DefaultTheme.StatusBar.Width(spacerWidth).Render("")

	return lipgloss.JoinHorizontal(lipgloss.Top,
		left,
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

func (v *View) viewSelection() string {
	if v.selTotal <= 0 {
		return "--/--"
	}

	clampedIdx := min(v.selIdx, v.selTotal)
	return fmt.Sprintf("%d/%d", clampedIdx, v.selTotal)
}

func (v *View) SetWidth(width int) {
	v.width = max(0, width)
}

func renderLeadingPills(modeText, pathText string, modeBG, pathBG, fg, barBG lipgloss.Color) string {
	leftCap := lipgloss.NewStyle().
		Foreground(modeBG).
		Background(barBG).
		Render("")

	modeBody := theme.DefaultTheme.StatusMode.
		Background(modeBG).
		Foreground(fg).
		Render(modeText)

	connector := lipgloss.NewStyle().
		Foreground(pathBG).
		Background(modeBG).
		Render("")

	pathBody := theme.DefaultTheme.StatusPath.
		Background(pathBG).
		Render(pathText)

	rightCap := lipgloss.NewStyle().
		Foreground(pathBG).
		Background(barBG).
		Render("")

	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftCap,
		modeBody,
		connector,
		pathBody,
		rightCap,
	)
}

func renderConnectedPills(leftText, rightText string, leftBG, rightBG, fg, barBG lipgloss.Color) string {
	// Left cap
	leftCap := lipgloss.NewStyle().
		Foreground(leftBG).
		Background(barBG).
		Render("")

	leftBody := theme.DefaultTheme.StatusInfo.
		Background(leftBG).
		Foreground(fg).
		Render(leftText)

	// Connector uses foreground of next BG and background of current BG for a seamless join.
	connector := lipgloss.NewStyle().
		Foreground(rightBG).
		Background(leftBG).
		Render("")

	rightWidth := max(6, lipgloss.Width(rightText))

	rightBody := theme.DefaultTheme.StatusInfo.
		Background(rightBG).
		Foreground(fg).
		Padding(0).
		Width(rightWidth).
		AlignHorizontal(lipgloss.Center).
		Render(rightText)

	rightCap := lipgloss.NewStyle().
		Foreground(rightBG).
		Background(barBG).
		Render("")

	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftCap,
		leftBody,
		connector,
		rightBody,
		rightCap,
	)
}

// SetSelection updates the cursor position display (1-based index / total).
func (v *View) SetSelection(idx, total int) {
	v.selIdx = max(0, idx)
	v.selTotal = max(0, total)
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
	v.animRunning = true
	v.animIdx = 0
	v.animSeq++

	return tea.Batch(v.sizeTick(), func() tea.Msg {
		return v.performCalcDirSize(ctx, dir)
	})
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

func (v *View) sizeTick() tea.Cmd {
	seq := v.animSeq
	return tea.Tick(sizeAnimInterval, func(time.Time) tea.Msg {
		return sizeAnimTick{seq: seq}
	})
}
