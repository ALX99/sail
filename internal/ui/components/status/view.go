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

type pillSegment struct {
	text     string
	bg       lipgloss.Color
	minWidth int
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
	selCount int
	selName  string

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

	sizeText := v.viewSize()
	if v.animRunning {
		sizeText = sizeAnimFrames[v.animIdx]
	}

	// Left cluster: path styled as a pill
	left := renderPathPill(
		pathStr,
		v.selName,
		theme.Sky,
		theme.Surface1,
		theme.Surface0,
	)

	// Info segments: selection count + cursor position + size
	info := renderPills(
		[]pillSegment{
			{text: v.viewSelectionCount(), bg: theme.Green},
			{text: v.viewSelection(), bg: theme.Mauve, minWidth: 7},
			{text: sizeText, bg: theme.Sapphire, minWidth: 7},
		},
		theme.Base,
		theme.Surface0,
	)

	// Spacer to push info (and size) to the right
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

func (v *View) viewSelectionCount() string {
	return fmt.Sprintf("[%d]", v.selCount)
}

func (v *View) SetWidth(width int) {
	v.width = max(0, width)
}

func renderPathPill(pathText, pathHighlight string, pathHLColor, pathBG, barBG lipgloss.Color) string {
	leftCap := lipgloss.NewStyle().
		Foreground(pathBG).
		Background(barBG).
		Render("")

	pathStyle := theme.DefaultTheme.StatusPath.
		Background(pathBG)

	if pathHighlight != "" && !strings.HasSuffix(pathText, "/") {
		pathText += "/"
	}

	pathBody := lipgloss.JoinHorizontal(lipgloss.Top,
		pathStyle.Render(pathText),
		pathStyle.Foreground(pathHLColor).Render(pathHighlight),
	)

	rightCap := lipgloss.NewStyle().
		Foreground(pathBG).
		Background(barBG).
		Render("")

	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftCap,
		pathBody,
		rightCap,
	)
}

func renderPills(pills []pillSegment, fg, barBG lipgloss.Color) string {
	segments := make([]string, 0, len(pills)*2+1)

	// Opening cap
	segments = append(segments, lipgloss.NewStyle().
		Foreground(pills[0].bg).
		Background(barBG).
		Render(""))

	for i, pill := range pills {
		if i > 0 {
			// Connector to next pill
			conn := lipgloss.NewStyle().
				Foreground(pills[i].bg).
				Background(pills[i-1].bg).
				Render("")
			segments = append(segments, conn)
		}

		body := theme.DefaultTheme.StatusInfo.
			Background(pills[i].bg).
			Foreground(fg)

		if pill.minWidth > 0 {
			body = body.
				Padding(0).
				Width(max(pill.minWidth, lipgloss.Width(pill.text))).
				AlignHorizontal(lipgloss.Center)
		}

		segments = append(segments, body.Render(pill.text))
	}

	// Closing cap
	segments = append(segments, lipgloss.NewStyle().
		Foreground(pills[len(pills)-1].bg).
		Background(barBG).
		Render(""))

	return lipgloss.JoinHorizontal(lipgloss.Top, segments...)
}

// SetSelection updates the cursor position display (1-based index / total),
// tracks the total number of selected files, and stores the current entry name.
func (v *View) SetSelection(idx, total, selected int, name string) {
	v.selIdx = max(0, idx)
	v.selTotal = max(0, total)
	v.selCount = max(0, selected)
	v.selName = name
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
