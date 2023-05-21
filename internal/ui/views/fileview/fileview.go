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
	"github.com/rs/zerolog/log"
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

type View struct {
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

func New(path string, width, height int, cfg config.Config) View {
	id++

	return View{
		id:            id,
		path:          path,
		scrollPadding: cfg.Settings.ScrollPadding,
		w:             width,
		h:             height,
	}
}

// Init will return the necessary tea.Msg for the fileWindow
// to become initialized and ready
func (v View) Init() tea.Msg {
	dir, err := fs.NewDirectory(v.path)
	if err != nil {
		log.Err(err).
      Str("path",v.path).
      Msg("Failed to read directory")
		return windowMsg{to: v.id, msg: err}
	}
	return windowMsg{to: v.id, msg: dir}
}

func (v View) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case windowMsg:
		if msg.to != v.id {
			break
		}

		switch msg := msg.msg.(type) {
		case fs.Directory:
			v.dir = msg
			v.err = nil

		case error:
			v.err = msg

		}

	case tea.KeyMsg:
		switch kp := msg.String(); kp {
		case ".":
			v.dir.ToggleShowHiddenFiles()
			return v, nil
		}
	}

	return v, nil
}

func (v View) View() string {
	if v.err != nil { // check error first
		return v.err.Error()
	}

	var nameBuilder strings.Builder
	names := make([]string, 0, v.dir.GetVisibleFileCount())
	drawn := 0

	for i := v.fileStart; i < v.dir.GetVisibleFileCount() && drawn < v.h; i++ {
		charsWritten := 0
		drawn++
		if i == v.dir.GetCursorIndex() {
			nameBuilder.WriteString("> ")
			charsWritten += 2
		}

		selectedFile := v.dir.GetFileAtIndex(i).GetDirEntry()
		name := selectedFile.Name()
		if len(name)+charsWritten > v.w {
			name = name[:v.w-charsWritten-1] + "~"
		}
		charsWritten += len(name)

		nameBuilder.WriteString(util.GetStyle(selectedFile).Render(name))

		if charsWritten+1 <= v.w && selectedFile.IsDir() {
			nameBuilder.WriteString("/")
		}

		names = append(names, nameBuilder.String())
		nameBuilder.Reset()
	}

	style := lipgloss.NewStyle().Width(v.w)
	return style.Render(strings.Join(names, "\n"))
}

// SetSize sets the max allowed size of the window
func (v *View) SetSize(w, h int) *View {
	v.w = w
	v.h = h
	return v
}

// Move moves the cursor up or down
func (v *View) Move(dir Direction) *View {
	if dir == Up {
		if v.dir.GetCursorIndex() > 0 {
			v.dir.MoveCursorUp()

			if v.fileStart > (v.dir.GetCursorIndex() - v.scrollPadding) {
				v.fileStart = util.Max(0, v.fileStart-1)
			}
		}
	} else {
		if v.dir.GetCursorIndex() < v.dir.GetVisibleFileCount()-1 {
			v.dir.MoveCursorDown()

			if v.dir.GetCursorIndex()-v.fileStart+1 > (v.h - v.scrollPadding) {
				v.fileStart = util.Min(v.dir.GetVisibleFileCount()-v.h, v.fileStart+1)
			}
		}
	}
	return v
}

// GetSelection returns the current file the cursor is over
func (v View) GetSelection() fss.DirEntry {
	return v.dir.GetFileAtCursor().GetDirEntry()
}

// GetSelectedPath returns the path to the viewed directory
func (v View) GetPath() string {
	return v.path
}

// GetSelectedPath returns the path to the currently selected file
func (v View) GetSelectedPath() string {
	return path.Join(v.path, v.GetSelection().Name())
}

func (v View) logState() {
	log.Debug().
		Str("path", v.path).
		Int("h", v.h).
		Int("w", v.w).
		Send()
}
