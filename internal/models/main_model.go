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

var pathAnimDuration = 250 * time.Millisecond

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
			m.cursor.r = max(0, m.cursor.r-1)
		case m.cfg.Settings.Keymap.NavDown:
			if m.cursor.c == 0 {
				m.cursor.r = min(m.cursor.r+1, min(len(m.files)-1, m.maxRows-1))
			} else {
				if m.cursorOffset() < len(m.files)-1 {
					m.cursor.r = min(m.cursor.r+1, m.maxRows-1)
				}
			}
			return m, nil

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
			if len(m.files) > 0 && m.files[m.cursorOffset()].IsDir() {
				return m, m.loadDir(path.Join(m.cwd, m.files[m.cursorOffset()].Name()))
			}
		case m.cfg.Settings.Keymap.NavHome:
			home, err := os.UserHomeDir()
			if err != nil {
				m.lastError = err
				return m, nil
			}
			return m, m.loadDir(home)
		case m.cfg.Settings.Keymap.Delete:
			if len(m.files) > 0 {
				// we optimistically believe that the file will be deleted
				delete(m.cachedDirSelections, m.cwd)

				return m, sequentially(
					func() tea.Msg {
						return osi.RemoveAll(path.Join(m.cwd, m.files[m.cursorOffset()].Name()))
					},
					m.loadDir(m.cwd),
				)
			}
			return m, nil

		}
	case tea.WindowSizeMsg:
		var fName string
		if len(m.files) > 0 {
			fName = m.files[m.cursorOffset()].Name()
		}

		m.maxRows = min(defaultMaxRows, max(1, msg.Height-3))

		m.trySelectFile(fName)

		return m, nil

	case dirLoaded:
		oldDir := m.cwd
		newDir := msg.path

		/// special case
		if oldDir == newDir {
			m.files = msg.files
			if len(m.files) > 0 {
				name, ok := m.cachedDirSelections[newDir]
				if ok {
					m.trySelectFile(name)
					if m.cursor.r == 0 && m.cursor.c == 0 {
						delete(m.cachedDirSelections, newDir)
					}
				} else {
					m.trySelectFile(m.files[min(m.cursorOffset(), len(m.files)-1)].Name())
				}
				m.cachedDirSelections[newDir] = m.files[m.cursorOffset()].Name()
			} else {
				delete(m.cachedDirSelections, newDir)
				m.setCursor(0, 0)
			}
			return m, nil
		}

		if m.prevCWD == "" && oldDir != newDir {
			m.prevCWD = oldDir
		}
		if len(m.files) > 0 {
			// cache the selected file for the previous directory
			m.cachedDirSelections[oldDir] = m.files[m.cursorOffset()].Name()
		}

		m.cwd = newDir
		m.files = msg.files

		fName, ok := m.cachedDirSelections[newDir]
		if !ok && path.Join(newDir, path.Base(oldDir)) == oldDir {
			// in case of a navigation to the parent directory
			// select the parent directory in the parent directory
			fName = path.Base(oldDir)
		}

		m.trySelectFile(fName)
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

func (m Model) View() string {
	t := time.Now()
	defer func() { log.Trace().Msgf("Rendered in %v", time.Since(t)) }()

	m.sb.Reset()

	// some eye candy; directories end with a slash
	if m.cwd != "/" {
		m.cwd += "/"
	}

	if m.prevCWD != "" {
		if strings.HasPrefix(m.prevCWD, m.cwd) && len(m.cwd) < len(m.prevCWD) {
			m.sb.WriteString(m.cwd + lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render(strings.TrimPrefix(m.prevCWD+"/", m.cwd)))
		} else if strings.HasPrefix(m.cwd, m.prevCWD) && len(m.cwd) > len(m.prevCWD) {
			// some eye candy; directories end with a slash
			if m.prevCWD != "/" {
				m.prevCWD += "/"
			}
			m.sb.WriteString(m.prevCWD + lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render(strings.TrimPrefix(m.cwd, m.prevCWD)))
		}
	} else {
		m.sb.WriteString(m.cwd)
	}

	m.sb.WriteString("\n\n")

	grid := make([][]fs.DirEntry, 0, m.maxRows)
	maxColNameLen := make([]int, len(m.files)/m.maxRows+1)

	for i, f := range m.files {
		if i < m.maxRows {
			grid = append(grid, make([]fs.DirEntry, 0, defaultMaxRows))
		}
		r, c := i%m.maxRows, i/m.maxRows
		maxColNameLen[c] = max(maxColNameLen[c], len(f.Name()))

		grid[r] = append(grid[r], f)
		maxColNameLen[len(grid[r])-1] = max(maxColNameLen[len(grid[r])-1], len(f.Name()))
	}

	for row := range len(grid) {
		for col, f := range grid[row] {
			if m.cursor.r == row && m.cursor.c == col {
				m.sb.WriteString(">")
			}

			extraPadding := 0

			// only pad if the column is not the last column
			if col < len(grid[row]) {
				extraPadding = maxColNameLen[col] - len(f.Name()) + 2

				if m.cursor.r == row && m.cursor.c == col {
					extraPadding--
				}
			}

			m.sb.WriteString(util.GetStyle(f).
				PaddingRight(extraPadding).
				Render(f.Name()))
		}
		m.sb.WriteString("\n")

		if row == len(grid)-1 {
			if m.lastError != nil {
				m.sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render(m.lastError.Error()))
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
	m.cursor.c = r
	m.cursor.r = c
}

// trySelectFile tries to select a file by name or sets the cursor to the first file
// if the file is not found.
func (m *Model) trySelectFile(fName string) {
	index := slices.IndexFunc(m.files, func(dir fs.DirEntry) bool {
		return dir.Name() == fName
	})

	if index != -1 {
		m.setCursor(index/m.maxRows, index%m.maxRows)
	} else {
		m.setCursor(0, 0)
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
