package models

import (
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/alx99/sail/internal/config"
	"github.com/alx99/sail/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

const defaultMaxRows = 10

var (
	pathAnimDuration = 250 * time.Millisecond
	white            = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	red              = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
)

type dirLoaded struct {
	path  string
	files []fs.DirEntry
}
type clearPrevCWD struct{}

type position struct {
	r, c int
}

type Model struct {
	cfg config.Config

	clearAnimAt         time.Time         // time when the previous working directory should be cleared
	prevCWD             string            // previous working directory
	cwd                 string            // current working directory
	files               []fs.DirEntry     // current files in that directory
	cursor              position          // cursor
	cachedDirSelections map[string]string // cached file names for directories
	selectedFiles       map[string]any    // selected files
	maxRows             int               // the maximum number of rows to display
	lastError           error             // last error that occurred

	// for performance purposes
	sb strings.Builder
}

func NewMain(cwd string, cfg config.Config) Model {
	return Model{
		cwd:                 cwd,
		cfg:                 cfg,
		maxRows:             defaultMaxRows,
		cachedDirSelections: make(map[string]string, 100),
		selectedFiles:       make(map[string]any, 100),
		sb:                  strings.Builder{},
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadDir(m.cwd)
}

func (m Model) cursorOffset() int {
	// m.logCursor()
	return (m.cursor.c * m.maxRows) + m.cursor.r
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// clear the last error
	if m.lastError != nil {
		m.lastError = nil
	}
	defer func() {
		m.logCursor()
	}()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.cfg.PrintLastWD != "" {
				err := m.writeLastWD()
				if err != nil {
					log.Error().Err(err).Send()
				}
			}
			return m, tea.Quit

		case m.cfg.Settings.Keymap.NavUp:
			return m.goUp(false), nil

		case m.cfg.Settings.Keymap.NavDown:
			return m.goDown(false), nil

		case m.cfg.Settings.Keymap.NavLeft:
			m.cursor.c = max(0, m.cursor.c-1)

			return m, nil
		case m.cfg.Settings.Keymap.NavRight:
			m.cursor.c++
			if m.cursorOffset() >= len(m.files) {
				m.cursor.c--
			}

			return m, nil
		case m.cfg.Settings.Keymap.NavOut:
			return m, m.loadDir(path.Dir(m.cwd))

		case m.cfg.Settings.Keymap.NavIn:
			if len(m.files) <= 0 {
				return m, nil
			}

			currFile := m.currFile()
			if currFile.IsDir() {
				return m, m.loadDir(path.Join(m.cwd, currFile.Name()))
			}

			if currFile.Type() != fs.ModeSymlink {
				return m, nil
			}

			path, err := os.Readlink(path.Join(m.cwd, currFile.Name()))
			if err != nil {
				m.lastError = err
				return m, nil
			}

			info, err := os.Stat(path)
			if err != nil {
				m.lastError = err
				return m, nil
			}

			if info.IsDir() {
				return m, m.loadDir(path)
			}
			return m, nil

		case m.cfg.Settings.Keymap.NavHome:
			home, err := os.UserHomeDir()
			if err != nil {
				m.lastError = err
				return m, nil
			}
			return m, m.loadDir(home)

		case m.cfg.Settings.Keymap.Delete:
			if len(m.files) <= 0 {
				return m, nil
			}
			return m, sequentially(
				func() tea.Msg {
					return osi.RemoveAll(path.Join(m.cwd, m.currFile().Name()))
				},
				m.loadDir(m.cwd),
			)

		case m.cfg.Settings.Keymap.Select:
			if len(m.files) <= 0 {
				return m, nil
			}

			fName := path.Join(m.cwd, m.currFile().Name())
			if _, ok := m.selectedFiles[fName]; ok {
				delete(m.selectedFiles, fName)
				log.Debug().Msgf("Deselected %s", fName)
			} else {
				m.selectedFiles[fName] = nil
				log.Debug().Msgf("Selected %s", fName)
			}

			return m.goDown(true), nil
		}
	case tea.WindowSizeMsg:
		var fName string
		if len(m.files) > 0 {
			fName = m.currFile().Name()
		}

		m.maxRows = min(defaultMaxRows, max(1, msg.Height-3))

		m.trySelectFile(fName)

		return m, nil

	case dirLoaded:
		oldDir := m.cwd
		newDir := msg.path

		if m.prevCWD == "" && oldDir != newDir {
			m.prevCWD = oldDir
		}
		if len(m.files) > 0 {
			// cache the selected file for the previous directory
			m.cachedDirSelections[oldDir] = m.currFile().Name()
		}

		m.cwd = newDir
		m.files = msg.files

		fName, ok := m.cachedDirSelections[newDir]
		if !ok && path.Join(newDir, path.Base(oldDir)) == oldDir {
			// in case of a navigation to the parent directory
			// select the parent directory in the parent directory
			fName = path.Base(oldDir)
		}

		if fName != "" {
			if !m.trySelectFile(fName) {
				m.ensureNoCursorOOB()
			}
		} else {
			m.setCursor(0, 0)
		}

		m.clearAnimAt = time.Now().Add(pathAnimDuration)
		return m, func() tea.Msg {
			time.Sleep(pathAnimDuration)
			return clearPrevCWD{}
		}
	case clearPrevCWD:
		if time.Now().After(m.clearAnimAt) {
			m.prevCWD = ""
		}
		return m, nil
	case error:
		m.lastError = msg
		log.Error().Err(msg).Msg("Error occurred")
		return m, nil
	}

	return m, nil
}

// viewCWD renders the current working directory
func (m Model) viewCWD() string {
	// If there is no previous CWD, just return the current CWD
	if m.prevCWD == "" {
		m.sb.WriteString(m.cwd)
		if m.cwd == "/" {
			return m.sb.String()
		}
		m.sb.WriteString("/")
		return m.sb.String()
	}

	common := longestCommonPath(m.cwd, m.prevCWD) + "/"

	if len(m.cwd) < len(m.prevCWD) {
		m.sb.WriteString(common + red.Render(strings.TrimPrefix(m.prevCWD+"/", common)))
	} else {
		m.sb.WriteString(common + lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render(strings.TrimPrefix(m.cwd, common)+"/"))
	}

	return m.sb.String()
}

func (m Model) View() string {
	t := time.Now()
	defer func() { log.Trace().Msgf("Rendered in %v", time.Since(t)) }()

	m.sb.Reset()
	m.sb.WriteString(m.viewCWD())
	m.sb.WriteString("\n\n")

	grid := make([][]fs.DirEntry, 0, m.maxRows)
	maxColNameLen := make([]int, len(m.files)/m.maxRows+1)

	for i, f := range m.files {
		if i < m.maxRows {
			grid = append(grid, make([]fs.DirEntry, 0, defaultMaxRows))
		}
		r, c := i%m.maxRows, i/m.maxRows
		maxColNameLen[c] = max(maxColNameLen[c], lipgloss.Width(f.Name()))

		grid[r] = append(grid[r], f)
		maxColNameLen[len(grid[r])-1] = max(maxColNameLen[len(grid[r])-1], lipgloss.Width(f.Name()))
	}

	for row := range len(grid) {
		for col, f := range grid[row] {
			if m.cursor.r == row && m.cursor.c == col {
				m.sb.WriteString(white.Render(">"))
			}

			// +2 because of (cursor and 1 space for next row)
			rightPad := maxColNameLen[col] - lipgloss.Width(f.Name()) + 2

			if m.cursor.r == row && m.cursor.c == col {
				rightPad--
			}

			s := util.GetStyle(f)
			if m.isSelected(f.Name()) {
				s = s.Copy().Underline(true).Bold(true)
			}

			m.sb.WriteString(s.Render(f.Name()))
			m.sb.WriteString(strings.Repeat(" ", rightPad))
		}
		m.sb.WriteString("\n")

		if row == len(grid)-1 {
			if m.lastError != nil {
				m.sb.WriteString(red.Render(m.lastError.Error()))
			}
		}
	}

	return m.sb.String()
}

func (m Model) loadDir(path string) tea.Cmd {
	return func() tea.Msg {
		files, err := osi.ReadDir(path)
		if err != nil {
			return err
		}
		return dirLoaded{files: files, path: path}
	}
}

func (m Model) logCursor() {
	log.Trace().Msgf("cursor(%v, %v)", m.cursor.c, m.cursor.r)
}

func (m *Model) setCursor(r, c int) {
	m.cursor.c = c
	m.cursor.r = r
}

// trySelectFile tries to select a file by name.
// It returns true if the file was found and selected.
func (m *Model) trySelectFile(fName string) bool {
	index := slices.IndexFunc(m.files, func(dir fs.DirEntry) bool {
		return dir.Name() == fName
	})

	if index != -1 {
		m.setCursor(index%m.maxRows, index/m.maxRows)
	}
	return index != -1
}

// ensureNoCursorOOB ensures that the cursor is not out of bounds
// by moving the cursor upwards until it is within the bounds.
func (m *Model) ensureNoCursorOOB() {
	for m.cursorOffset() > 0 && m.cursorOffset() > len(m.files)-1 {
		*m = m.goUp(true)
	}
}

func (m Model) writeLastWD() error {
	f, err := os.Create(m.cfg.PrintLastWD)
	if err != nil {
		return err
	}

	defer f.Close()
	_, err = f.WriteString(m.cwd)
	return err
}

func (m Model) goDown(wrap bool) Model {
	prevCursor := m.cursor
	if m.cursorOffset() < len(m.files)-1 {
		m.cursor.r = min(m.cursor.r+1, m.maxRows-1)
	}

	if prevCursor == m.cursor && wrap {
		if m.cursorOffset()+1 < len(m.files) {
			m.setCursor(0, m.cursor.c+1)
		} else {
			// here we MUST be at the end of the list
			m.setCursor(0, 0)
		}
	}

	return m
}

func (m Model) goUp(wrap bool) Model {
	prevCursor := m.cursor
	if m.cursor.r > 0 {
		m.cursor.r--
	}

	if prevCursor == m.cursor && wrap {
		if m.cursor.c > 0 {
			m.setCursor(m.maxRows-1, m.cursor.c-1)
		} else {
			// here we MUST be at the beginning of the list
			m.setCursor((len(m.files)-1)%m.maxRows, (len(m.files)-1)/m.maxRows)
		}
	}

	return m
}

func (m Model) isSelected(name string) bool {
	_, ok := m.selectedFiles[path.Join(m.cwd, name)]
	return ok
}

func (m Model) currFile() fs.DirEntry {
	return m.files[m.cursorOffset()]
}

// sequentially produces a command that sequentially executes the given
// commands.
// The tea.Msg returned is the first non-nil message returned by a Cmd.
func sequentially(cmds ...tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}
			if msg := cmd(); msg != nil {
				return msg
			}
		}
		return nil
	}
}
