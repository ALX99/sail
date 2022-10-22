package fileview

import (
	"path"
	"strings"

	fss "io/fs"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/fs"
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
	dir      fs.Directory
	moveDown bool
	err      error

	id            ID
	h, w          int
	prevFileStart int
	fileStart     int

	// Configurable settings
	scrollPadding int
}

func New(path string, width, height int, cfg config.Config) Window {
	id++

	return Window{
		id:            id,
		path:          path,
		scrollPadding: cfg.Settings.ScrollPadding,
		w:             width,
		h:             height,
	}
}

// Init will return the necessary tea.Msg for the fileWindow
// to become initialized and ready
func (fw Window) Init() tea.Msg {
	dir, err := fs.NewDirectory(fw.path)
	if err != nil {
		util.Log.Err(err).Msg("Failed to read directory")
		return windowMsg{to: fw.id, msg: err}
	}
	return windowMsg{to: fw.id, msg: dir}
}

func (fw Window) Update(msg tea.Msg) (Window, tea.Cmd) {
	switch msg := msg.(type) {
	case windowMsg:
		if msg.to != fw.id {
			break
		}
		switch msg := msg.msg.(type) {
		case fs.Directory:
			fw.dir = msg
			fw.err = nil

		case error:
			fw.err = msg

		case tea.KeyMsg:
			switch kp := msg.String(); kp {
			case ".":
				return fw, nil
			}
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
	names := make([]string, 0, fw.dir.GetVisibleFileCount())
	drawn := 0

	for i := fw.fileStart; i < fw.dir.GetVisibleFileCount() && drawn < fw.h; i++ {
		charsWritten := 0
		drawn++
		if i == fw.dir.GetCursorIndex() {
			nameBuilder.WriteString("> ")
			charsWritten += 2
		}

		selectedFile := fw.dir.GetFileAtIndex(i).GetDirEntry()
		name := selectedFile.Name()
		if len(name)+charsWritten > fw.w {
			name = name[:fw.w-charsWritten-1] + "~"
		}
		charsWritten += len(name)

		nameBuilder.WriteString(util.GetStyle(selectedFile).Render(name))

		if charsWritten+1 <= fw.w && selectedFile.IsDir() {
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
		if fw.dir.GetCursorIndex() > 0 {
			fw.dir.MovCursorUp()

			if fw.fileStart > (fw.dir.GetCursorIndex() - fw.scrollPadding) {
				fw.fileStart = util.Max(0, fw.fileStart-1)
			}
		}
	} else {
		if fw.dir.GetCursorIndex() < fw.dir.GetVisibleFileCount()-1 {
			fw.dir.MovCursorDown()

			if fw.dir.GetCursorIndex()-fw.fileStart+1 > (fw.h - fw.scrollPadding) {
				fw.fileStart = util.Min(fw.dir.GetVisibleFileCount()-fw.h, fw.fileStart+1)
			}
		}
	}
	return fw
}

// GetSelection returns the current file the cursor is over
func (fw Window) GetSelection() fss.DirEntry {
	return fw.dir.GetFileAtCursor().GetDirEntry()
}

// GetSelectedPath returns the path to the viewed directory
func (fw Window) GetPath() string {
	return fw.path
}

// GetSelectedPath returns the path to the currently selected file
func (fw Window) GetSelectedPath() string {
	return path.Join(fw.path, fw.GetSelection().Name())
}

func (fw Window) logState() {
	util.Log.Debug().
		Str("path", fw.path).
		Int("h", fw.h).
		Int("w", fw.w).
		Send()
}
