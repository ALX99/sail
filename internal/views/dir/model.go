package dir

//go:generate stringer -type=Role,Direction
import (
	"os"
	"path"
	"strings"
	"time"

	fss "io/fs"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/fs"
	"github.com/alx99/fly/internal/msgs"
	"github.com/alx99/fly/internal/state"
	"github.com/alx99/fly/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type (
	Direction uint8
	Role      uint8
)

const (
	Up Direction = iota
	Down
)

const (
	Parent Role = iota
	Working
	Child
)

type Model struct {
	state *state.State
	path  string
	dir   fs.Directory
	err   error

	h, w int

	// State
	role             Role
	offset           int
	cursor           int
	loaded           bool
	visibleFiles     []fs.File
	visibleFileCount int

	cfg    config.Settings
	logger zerolog.Logger
}

func New(path string, role Role, state *state.State, width, height int, cfg config.Config) Model {
	return Model{
		role:         role,
		state:        state,
		path:         path,
		cfg:          cfg.Settings,
		w:            width,
		h:            height,
		visibleFiles: []fs.File{},
		logger: log.With().
			Str("role", role.String()).
			Str("path", path).
			Logger(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.cmdRead("")
}

func (m Model) InitAndSelect(name string) tea.Cmd {
	return m.cmdRead(name)
}

// Reinit reinitializes the directory and tries (best-effort)
// to select the same file as before
func (m Model) Reinit() tea.Cmd {
	if m.visibleFileCount > 0 {
		return m.cmdRead(m.visibleFiles[m.cursor].GetDirEntry().Name())
	}
	return m.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case msgDirLoaded:
		if msg.role == m.role {
			m.path = msg.dir.Path()
			wasLoaded := m.loaded
			// TODO if a file/folder is deleted above the currently
			// selected one, the cursor needs to decrease by one
			// or however many were deleted
			m.setDirectory(msg.dir)
			if msg.onLoadSelect != "" {
				if m.cfg.ShowHiddenFiles {
					m.setSelectedFile(msg.onLoadSelect)
				} else if msg.onLoadSelect[0] != '.' { // only if not hidden
					m.setSelectedFile(msg.onLoadSelect)
				}
			}
			// First load on WD, means CD needs to be loaded
			if m.role == Working && !wasLoaded && m.GetSelection().IsDir() {
				m.logger.Debug().Str("select", m.GetSelection().Name()).Msg("running suboptimal cd init")
				return m, func() tea.Msg {
					dir, err := fs.NewDirectory(m.GetSelectedPath())
					if err != nil {
						m.logger.Err(err).
							Msg("Failed to read directory")
						return msgDirError{
							role: Child,
							err:  err,
						}
					}
					return msgDirLoaded{
						role: Child,
						dir:  dir,
					}
				}
			}
			return m, nil
			// TODO cmdTickRead no longer works, because
			// each new directory model does not have a unique ID
			// this might be better to do inside primary anyway
			// (issue the MsgDirReload command from primary every 5s)
			// return m, m.cmdTickRead()
		}

		return m, nil

	case msgDirError:
		// Make sure it is addressed to me
		if msg.role != m.role {
			break
		}
		m.err = msg.err
		return m, nil

	case msgs.MsgDirReload:
		if m.loaded {
			return m, m.Reinit()
		}
		return m, m.cmdRead("")

	case tea.KeyMsg:
		switch kp := msg.String(); kp {
		}
	}

	return m, nil
}

func (m Model) View() string {
	t := time.Now()
	defer func() {
		m.logger.Trace().
			Str("dur", time.Since(t).String()).
			Msg("View render")
	}()

	style := lipgloss.NewStyle().Height(m.h).Width(m.w)
	if m.err != nil { // check error first
		style = style.Foreground(lipgloss.Color("#ff0000")).Bold(true)
		if os.IsNotExist(m.err) {
			return style.Render("file/folder not found!")
		}
		return style.Render(m.err.Error())
	}

	if !m.loaded {
		return style.Render("loading...")
	}

	if m.visibleFileCount == 0 {
		return style.Render("empty")
	}

	var nameBuilder strings.Builder
	names := make([]string, 0, m.visibleFileCount)

	for i := m.offset; i < m.visibleFileCount; i++ {
		if i-m.offset == m.h {
			break
		}

		file := m.visibleFiles[i]

		charsWritten := 0
		if i == m.cursor {
			nameBuilder.WriteString("> ")
			charsWritten += 2
		}

		selectedFile := file.GetDirEntry()
		name := selectedFile.Name()
		if len(name)+charsWritten > m.w {
			name = name[:m.w-charsWritten-1] + "~"
		}
		charsWritten += len(name)

		nameBuilder.WriteString(
			util.GetStyle(selectedFile).
				Underline(m.state.IsSelected(path.Join(m.path, selectedFile.Name()))).
				Render(name))

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
		if m.cursor >= 1 {
			m.cursor--
		}
	} else {
		if m.cursor < m.visibleFileCount-1 {
			m.cursor++
		}
	}

	// todo fix offset to be able to move more than 1
	// update the offset
	if m.offset > 0 && m.cursor < m.offset+m.cfg.ScrollPadding {
		m.offset--
	} else {
		if m.cursor-m.offset >= m.h-m.cfg.ScrollPadding {
			m.offset = min(m.visibleFileCount-m.h, m.offset+1)
		}
	}

	m.logger.Trace().
		Int("i", m.cursor).
		Int("offset", m.offset).
		Str("file", path.Join(m.path, m.GetSelection().Name())).
		Int("fCount", m.visibleFileCount).
		Msg("cusor moved")

	return m
}

// setDirectory sets the current directory for the model
func (m *Model) setDirectory(dir fs.Directory) {
	m.dir = dir
	m.err = nil
	m.visibleFiles = make([]fs.File, 0, m.visibleFileCount)
	cnt := 0
	for _, f := range dir.Files() {
		if !f.Hidden() {
			cnt++
			m.visibleFiles = append(m.visibleFiles, f)
		} else if m.cfg.ShowHiddenFiles {
			cnt++
			m.visibleFiles = append(m.visibleFiles, f)
		}
	}
	m.visibleFileCount = cnt

	// Ensure that the cursor does not exceed the
	// amount of selectable files
	if m.cursor >= m.visibleFileCount {
		// cursor is zero-indexed
		m.cursor = max(0, m.visibleFileCount-1)
	}
	m.loaded = true
	m.logger.Debug().Int("cursor", m.cursor).
		Int("visibleFileCount", m.visibleFileCount).
		Msg("loaded directory")
}

func (m *Model) setSelectedFile(name string) {
	for i, file := range m.visibleFiles {
		if file.GetDirEntry().Name() == name {
			m.cursor = i
			break
		}
	}
}

// IsFocusable returns true if it is possible to focus the current view
func (m Model) IsFocusable() bool {
	return m.err == nil
}

// GetSelection returns the current file the cursor is over
func (m Model) GetSelection() fss.DirEntry {
	if !m.loaded {
		return emptyDirEntry{}
	}
	return m.visibleFiles[m.cursor].GetDirEntry()
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
	return m.visibleFileCount == 0
}

func (m Model) ChangeRole(role Role) Model {
	m.role = role
	return m
}

// cmdRead reads the current directory
func (m Model) cmdRead(selectName string) tea.Cmd {
	m.logger.Trace().Msgf("Loading directory")
	return func() tea.Msg {
		dir, err := fs.NewDirectory(m.path)
		if err != nil {
			m.logger.Err(err).Msg("Failed to read directory")
			return msgDirError{
				role: m.role,
				err:  err,
			}
		}
		return msgDirLoaded{
			role:         m.role,
			dir:          dir,
			onLoadSelect: selectName,
		}
	}
}

// cmdTickRead sleeps one second before calling cmdRead
func (m Model) cmdTickRead() tea.Cmd {
	return tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
		return m.cmdRead("")()
	})
}
