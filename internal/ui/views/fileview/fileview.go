package fileview

import (
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/alx99/fly/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Direction uint8
type ID uint32

const (
	Up Direction = iota
	Down
)

var (
	id ID = 0
)

type windowMsg struct {
	msg interface{}
	to  ID
}

type Window struct {
	path     string
	files    []fs.DirEntry
	moveDown bool
	err      error

	id             ID
	h, w           int
	pos            int
	prevFileStart  int
	fileStart      int
	visibleFileLen int

	// Configurable settings
	scrollPadding int
}

func New(path string, width, height int) Window {
	id++
	return Window{
		id:            id,
		path:          path,
		scrollPadding: 2,
		w:             width,
		h:             height,
	}
}

// Init will return the necessary tea.Msg for the fileWindow
// to become initialized and ready
func (fw Window) Init() tea.Msg {
	files, err := os.ReadDir(fw.path)
	if err != nil {
		return windowMsg{to: fw.id, msg: err}
	}
	return windowMsg{to: fw.id, msg: files}
}

func (fw Window) Update(msg tea.Msg) (Window, tea.Cmd) {
	switch msg := msg.(type) {
	case windowMsg:
		if msg.to != fw.id {
			break
		}
		switch msg := msg.msg.(type) {
		case []fs.DirEntry:
			fw.files = msg
			fw.visibleFileLen = len(fw.files)
			fw.err = nil

		case error:
			fw.err = msg

		default:
			panic("developer error")
		}

	case tea.WindowSizeMsg:
		fw.h = msg.Height
	}

	return fw, nil
}

func (fw Window) View() string {
	if fw.err != nil { // check error first
		return fw.err.Error()
	}

	var nameBuilder strings.Builder
	names := make([]string, 0, fw.visibleFileLen)
	drawn := 0

	for i := fw.fileStart; i < fw.visibleFileLen && drawn < fw.h; i++ {
		drawn++
		if i == fw.pos {
			nameBuilder.WriteString("> ")
		}

		nameBuilder.WriteString(util.GetStyle(fw.files[i]).Render(fw.files[i].Name()))

		if fw.files[i].IsDir() {
			nameBuilder.WriteString("/")
		}

		names = append(names, nameBuilder.String())
		nameBuilder.Reset()
	}

	style := lipgloss.NewStyle().Width(fw.w)
	return style.Render(strings.Join(names, "\n"))
}

// SetWidth sets the max allowed width of the window
func (fw *Window) SetWidth(w int) *Window {
	fw.w = w
	return fw
}

// Move moves the cursor up or down
func (fw *Window) Move(dir Direction) *Window {
	if dir == Up {
		if fw.pos > 0 {
			fw.pos -= 1

			if fw.fileStart > (fw.pos - fw.scrollPadding) {
				fw.fileStart = util.Max(0, fw.fileStart-1)
			}
		}
	} else {
		if fw.pos < fw.visibleFileLen-1 {
			fw.pos += 1

			if fw.pos-fw.fileStart+1 > (fw.h - fw.scrollPadding) {
				fw.fileStart = util.Min(fw.visibleFileLen-fw.h, fw.fileStart+1)
			}
		}
	}
	return fw
}

// GetSelection returns the current file the cursor is over
func (fw Window) GetSelection() fs.DirEntry {
	return fw.files[fw.pos]
}

// GetSelectedPath returns the path to the viewed directory
func (fw Window) GetPath() string {
	return fw.path
}

// GetSelectedPath returns the path to the currently selected file
func (fw Window) GetSelectedPath() string {
	return path.Join(fw.path, fw.GetSelection().Name())
}
