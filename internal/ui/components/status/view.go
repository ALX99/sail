package status

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/alx99/sail/internal/filesys"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	pathAnimDur = 250 * time.Millisecond
	errDur      = 1 * time.Second
	red         = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	green       = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
)

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

func splitHalf(width int) (int, int) {
	return width / 2, width - width/2
}

func (v *View) View() string {
	lw, rw := splitHalf(v.width)
	rlw, rrw := splitHalf(rw)

	fileInfo := "test x d r d"
	leftSide := fileInfo

	if v.error != nil && time.Now().Before(v.errorAt.Add(errDur)) {
		leftSide = red.Render(v.error.Error())
	}

	path := v.viewPath()
	size := v.viewSize()
	counts := v.viewCounts()

	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Width(lw).Render(path),
			lipgloss.NewStyle().Width(rlw).
				Render(leftSide),
			lipgloss.NewStyle().Width(rrw).
				Align(lipgloss.Right).
				Render(counts+" ("+size+")"),
		),
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
	return fmt.Sprintf("files: %2d dirs: %2d", f, d)
}

func (v *View) viewPath() string {
	v.sb.Reset()

	// If there is no previous CWD, just return the current CWD
	if v.prevWD == "" || time.Now().After(v.navAt.Add(pathAnimDur)) {
		v.sb.WriteString(v.wd.Path())
		if v.wd.Path() == "/" {
			return v.sb.String()
		}
		v.sb.WriteString("/")
		return v.sb.String()
	}

	if v.wd.Path() == v.prevWD {
		v.sb.WriteString(v.wd.Path())
		v.sb.WriteString("/")
		return v.sb.String()
	}

	common := longestCommonPath(v.wd.Path(), v.prevWD) + "/"

	if len(v.wd.Path()) < len(v.prevWD) {
		v.sb.WriteString(common + red.Render(strings.TrimPrefix(v.prevWD+"/", common)))
	} else {
		v.sb.WriteString(common + green.Render(strings.TrimPrefix(v.wd.Path()+"/", common)))
	}

	return v.sb.String()
}

func (v *View) SetWidth(width int) {
	v.width = width - 2 // - 2 for border
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

	return tea.Batch(
		func() tea.Msg {
			time.Sleep(pathAnimDur)
			return struct{}{} // dummy message to trigger update
		},
		v.calcDirSize,
	)
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

func longestCommonPath(s1, s2 string) string {
	split1 := strings.Split(s1, "/")
	split2 := strings.Split(s2, "/")
	minLen := min(len(split2), len(split1))
	common := make([]string, 0, minLen)

	for i := range minLen {
		if split1[i] != split2[i] {
			break
		}
		common = append(common, split1[i])
	}

	return strings.Join(common, "/")
}
