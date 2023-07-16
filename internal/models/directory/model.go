package directory

import (
	"os"
	"path"
	"strings"
	"time"

	fss "io/fs"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/fs"
	"github.com/alx99/fly/internal/state"
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

type directoryMsg struct {
	msg any
	to  ID
}

type Model struct {
	state *state.State
	path  string
	dir   fs.Directory
	err   error

	id          ID
	h, w        int
	offset      int
	cursorIndex int
	loaded      bool

	// Configurable settings
	scrollPadding int
}

func New(path string, state *state.State, width, height int, cfg config.Config) Model {
	id++

	return Model{
		state:         state,
		id:            id,
		path:          path,
		scrollPadding: cfg.Settings.ScrollPadding,
		w:             width,
		h:             height,
	}
}

func (m Model) Init() tea.Cmd {
	return cmdRead(m)
}

// Load loads the directory instantly
func (m *Model) Load() (err error) {
	m.dir, err = fs.NewDirectory(m.path)
	return
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case directoryMsg:
		// Make sure it is addressed to me
		if msg.to != m.id {
			break
		}

		switch msg := msg.msg.(type) {
		case fs.Directory:
			// TODO if a file/folder is deleted above the currently
			// selected one, the cursorIndex needs to decrease by one
			// or however many were deleted

			// Ensure that the cursorIndex does not exceed the
			// amount of selectable files
			if m.cursorIndex >= msg.FileCount() {
				m.cursorIndex = msg.FileCount() - 1
			}
			m.dir = msg
			m.err = nil
			return m, cmdTickRead(m)

		case error:
			m.err = msg
			return m, nil // Stop ticking in this case
		}

	case tea.KeyMsg:
		switch kp := msg.String(); kp {
		}
	}

	return m, nil
}

func (m Model) View() string {
	style := lipgloss.NewStyle().Width(m.w)
	if m.err != nil { // check error first
		style = style.Foreground(lipgloss.Color("#ff0000"))
		if os.IsNotExist(m.err) {
			return style.Bold(true).Render("error file/folder not found!")
		}
		return style.Render(m.err.Error())
	}

	if m.dir.FileCount() == 0 {
		return style.Render("empty")
	}

	var nameBuilder strings.Builder
	skipped := 0
	files := m.dir.Files()
	names := make([]string, 0, m.dir.FileCount())

	for i := m.offset; i < m.dir.FileCount(); i++ {
		if i-m.offset-skipped == m.h {
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
		}
	} else {
		if m.cursorIndex < m.dir.FileCount()-1 {
			m.cursorIndex++
		}
	}

	if m.cursorIndex < m.offset {
		m.offset--
	} else {
		skipped := 0

		if m.cursorIndex-m.offset-skipped >= m.h {
			// roof hit
			m.offset++
		}
	}

	log.Debug().
		Int("cursorIndex", m.cursorIndex).
		Int("offset", m.offset).
		Str("path", m.path).
		Int("fCount", m.dir.FileCount()).
		Str("file", m.dir.GetFileAtIndex(m.cursorIndex).GetDirEntry().Name()).
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
	return m.dir.FileCount() == 0
}

// cmdRead reads the current directory
func cmdRead(m Model) tea.Cmd {
	return func() tea.Msg {
		dir, err := fs.NewDirectory(m.path)
		if err != nil {
			log.Err(err).
				Str("path", m.path).
				Msg("Failed to read directory")
			return directoryMsg{to: m.id, msg: err}
		}
		return directoryMsg{to: m.id, msg: dir}
	}
}

// cmdTickRead sleeps one second before calling cmdRead
func cmdTickRead(m Model) tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return cmdRead(m)()
	})
}
