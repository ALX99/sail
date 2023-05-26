package directory

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

func (d Direction) String() string {
	if d == Up {
		return "up"
	}
	return "down"
}

var (
	id ID = 0
)

type windowMsg struct {
	msg interface{}
	to  ID
}

type Model struct {
	path string
	dir  fs.Directory
	err  error

	id          ID
	h, w        int
	offset      int
	cursorIndex int

	// Configurable settings
	scrollPadding int
}

func New(path string, width, height int, cfg config.Config) Model {
	id++

	return Model{
		id:            id,
		path:          path,
		scrollPadding: cfg.Settings.ScrollPadding,
		w:             width,
		h:             height,
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		dir, err := fs.NewDirectory(m.path)
		if err != nil {
			log.Err(err).
				Str("path", m.path).
				Msg("Failed to read directory")
			return windowMsg{to: m.id, msg: err}
		}
		return windowMsg{to: m.id, msg: dir}
	}
}

// Load loads the directory instantly
func (m *Model) Load() (err error) {
	m.dir, err = fs.NewDirectory(m.path)
	return
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case windowMsg:
		// Make sure it is addressed to me
		if msg.to != m.id {
			break
		}

		switch msg := msg.msg.(type) {
		case fs.Directory:
			m.dir = msg
			m.err = nil

		case error:
			m.err = msg

		}

	case tea.KeyMsg:
		switch kp := msg.String(); kp {
		case ".":
			m.dir.ToggleShowHiddenFiles()
			// TODO better logic
			m.offset = 0
			m.cursorIndex = 0
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil { // check error first
		return lipgloss.NewStyle().Width(m.w).Render(m.err.Error())
	}

	files := m.dir.VisibleFiles()
	var nameBuilder strings.Builder
	names := make([]string, 0, len(files))

	for i := m.offset; i < len(files); i++ {
		if i-m.offset == m.h {
			break
		}

		file := files[i]
		charsWritten := 0
		if i == m.cursorIndex {
			nameBuilder.WriteString("> ")
			charsWritten += 2
		}

		selectedFile := file.GetDirEntry()
		name := selectedFile.Name()
		if len(name)+charsWritten > m.w {
			name = name[:m.w-charsWritten-1] + "~"
		}
		charsWritten += len(name)

		nameBuilder.WriteString(util.GetStyle(selectedFile).Render(name))

		if charsWritten+1 <= m.w && selectedFile.IsDir() {
			nameBuilder.WriteString("/")
		}

		names = append(names, nameBuilder.String())
		nameBuilder.Reset()
	}

	style := lipgloss.NewStyle().Width(m.w)
	return style.Render(strings.Join(names, "\n"))
}

// SetSize sets the max allowed size of the window
func (m *Model) SetSize(w, h int) *Model {
	m.w = w
	m.h = h
	return m
}

// Move moves the cursor up or down
func (m *Model) Move(dir Direction) *Model {
	if dir == Up {
		if m.cursorIndex >= 1 {
			m.cursorIndex--
			if m.cursorIndex < m.offset {
				m.offset--
			}
		}
	} else {
		fileCount := len(m.dir.VisibleFiles())
		if m.cursorIndex < fileCount-1 {
			m.cursorIndex++
			if m.cursorIndex > m.h-1 {
				m.offset++
			}
		}
	}

	log.Debug().
		Int("cursorIndex", m.cursorIndex).
		Int("offset", m.offset).
		Str("direction", dir.String()).
		Str("fileName", m.dir.GetFileAtIndex(m.cursorIndex).GetDirEntry().Name()).
		Msg("moved")

	return m
}

// IsFocusable returns true if it is possible to focus the current view
func (m Model) IsFocusable() bool {
	return m.err == nil
}

// GetSelection returns the current file the cursor is over
func (m Model) GetSelection() fss.DirEntry {
	return m.dir.GetFileAtIndex(m.cursorIndex).GetDirEntry()
}

// GetSelectedPath returns the path to the viewed directory
func (m Model) GetPath() string {
	return m.path
}

// GetSelectedPath returns the path to the currently selected file
func (m Model) GetSelectedPath() string {
	if m.Empty() {
		return m.path
	}
	return path.Join(m.path, m.GetSelection().Name())
}

// Empty returns true if the directory is empty to the user
func (m Model) Empty() bool {
	return len(m.dir.VisibleFiles()) == 0
}

func (m Model) logState() {
	log.Debug().
		Str("path", m.path).
		Int("h", m.h).
		Int("w", m.w).
		Send()
}
