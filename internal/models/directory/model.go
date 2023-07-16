package directory

import (
	"os"
	"path"
	"strings"
	"sync/atomic"
	"time"

	fss "io/fs"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/fs"
	"github.com/alx99/fly/internal/msgs"
	"github.com/alx99/fly/internal/state"
	"github.com/alx99/fly/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

type Direction uint8

const (
	Up Direction = iota
	Down
)

var (
	uniqueID atomic.Uint32
)

type Model struct {
	state *state.State
	path  string
	dir   fs.Directory
	err   error

	h, w int

	// State
	offset      int
	cursorIndex int
	id          uint32

	// Configurable settings
	scrollPadding int
}

func New(path string, state *state.State, width, height int, cfg config.Config) Model {
	return Model{
		state:         state,
		path:          path,
		scrollPadding: cfg.Settings.ScrollPadding,
		w:             width,
		h:             height,
		id:            uniqueID.Add(1),
	}
}

func (m Model) Init() tea.Cmd {
	return m.cmdRead("")
}

func (m Model) InitAndSelect(name string) tea.Cmd {
	return m.cmdRead(name)
}

// Load loads the directory instantly
func (m *Model) Load() (err error) {
	m.dir, err = fs.NewDirectory(m.path)
	return
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case msgs.MsgDirLoaded:
		// Make sure it is addressed to me
		if msg.To != m.id {
			break
		}

		// TODO if a file/folder is deleted above the currently
		// selected one, the cursorIndex needs to decrease by one
		// or however many were deleted

		// Ensure that the cursorIndex does not exceed the
		// amount of selectable files
		if m.cursorIndex >= msg.Dir.FileCount() {
			m.cursorIndex = msg.Dir.FileCount() - 1
		}
		m.dir = msg.Dir
		m.err = nil
		if msg.Select != "" {
			m.SetSelectedFile(msg.Select)
		}
		return m, m.cmdTickRead()

	case msgs.MsgDirError:
		// Make sure it is addressed to me
		if msg.To != m.id {
			break
		}
		m.err = msg.Err
		return m, nil

	case tea.KeyMsg:
		switch kp := msg.String(); kp {
		}
	}

	return m, nil
}

func (m *Model) SetSelectedFile(name string) {
	for i, file := range m.dir.Files() {
		if file.GetDirEntry().Name() == name {
			m.cursorIndex = i
			break
		}
	}
}

func (m Model) View() string {
	t := time.Now()
	defer func() {
		log.Trace().
			Str("dur", time.Since(t).String()).
			Str("dir", m.path).
			Msg("View render")
	}()
	style := lipgloss.NewStyle().Width(m.w)
	if m.err != nil { // check error first
		style = style.Foreground(lipgloss.Color("#ff0000"))
		if os.IsNotExist(m.err) {
			return style.Bold(true).Render("file/folder not found!")
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

	log.Trace().
		Int("i", m.cursorIndex).
		Int("offset", m.offset).
		Str("file", path.Join(m.path, m.dir.GetFileAtIndex(m.cursorIndex).GetDirEntry().Name())).
		Int("fCount", m.dir.FileCount()).
		Msg("cusor moved")

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
func (m Model) cmdRead(selectName string) tea.Cmd {
	log.Trace().Msgf("Loading directory %q", m.path)
	return func() tea.Msg {
		dir, err := fs.NewDirectory(m.path)
		if err != nil {
			log.Err(err).
				Str("path", m.path).
				Msg("Failed to read directory")
			return msgs.MsgDirError{
				To:   m.id,
				Path: m.path,
				Err:  err,
			}
		}
		return msgs.MsgDirLoaded{
			To:     m.id,
			Path:   m.path,
			Dir:    dir,
			Select: selectName,
		}
	}
}

// cmdTickRead sleeps one second before calling cmdRead
func (m Model) cmdTickRead() tea.Cmd {
	return tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
		return m.cmdRead("")()
	})
}
