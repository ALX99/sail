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

const maxRows = 10

type dirLoaded struct {
	path  string
	files []fs.DirEntry
}

type position struct {
	c, r int
}

type Model struct {
	cfg config.Config

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
		maxRows:             maxRows,
		cachedDirSelections: make(map[string]string, 100),
		sb:                  strings.Builder{},
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadDir(m.cwd)
}

func (m Model) cursorOffset() int {
	// m.logCursor()
	return m.cursor.c*m.maxRows + m.cursor.r
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// clear the last error
	if m.lastError != nil {
		m.lastError = nil
	}

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
			if m.cursor.c < len(m.files)/m.maxRows {
				m.cursor.r = min(m.cursor.r+1, min(len(m.files), m.maxRows)-1)
			} else if m.cursor.c == len(m.files)/m.maxRows {
				m.cursor.r = min(m.cursor.r+1, len(m.files)%m.maxRows-1)
			}
			return m, nil

		case m.cfg.Settings.Keymap.NavLeft:
			m.cursor.c = max(0, m.cursor.c-1)
			return m, nil
		case m.cfg.Settings.Keymap.NavRight:
			m.cursor.c++
			if m.cursorOffset() >= len(m.files) {
				m.cursor.c-- // undo the cursor move
			}
			return m, nil
		case m.cfg.Settings.Keymap.NavOut:
			return m, m.loadDir(path.Dir(m.cwd))
		case m.cfg.Settings.Keymap.NavIn:
			if m.files != nil && m.files[m.cursorOffset()].IsDir() {
				return m, m.loadDir(path.Join(m.cwd, m.files[m.cursorOffset()].Name()))
			}

		}
	case tea.WindowSizeMsg:
		var fName string
		if m.files != nil && len(m.files) > 0 {
			fName = m.files[m.cursorOffset()].Name()
		}

		m.maxRows = min(maxRows, max(1, msg.Height-3))

		m.trySelectFile(m.files, fName)

		return m, nil

	case dirLoaded:
		prevDir := m.cwd
		newDir := msg.path
		if len(m.files) > 0 {
			// cache the selected file for the previous directory
			m.cachedDirSelections[prevDir] = m.files[m.cursorOffset()].Name()
		}

		m.cwd = newDir
		m.files = msg.files

		fName, ok := m.cachedDirSelections[newDir]
		if !ok {
			// try to determine the previous file name
			fName = path.Base(prevDir)

			if strings.HasPrefix(newDir, prevDir) && len(newDir) > len(prevDir) {
				if slices.ContainsFunc(msg.files, func(dir fs.DirEntry) bool {
					return dir.Name() == fName
				}) {
					// This is a special case where the previous directory is a subdirectory of the new directory
					// and it contains a file with the same name as the previous directory

					// For example, if the previous directory was ~/go and the new directory has a file named go (~/go/src/go)
					// it would cause the "go" file to be selected in the new directory even though it should not
					fName = ""
				}
			}
		}

		m.trySelectFile(m.files, fName)

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
	defer func() { log.Debug().Msgf("Rendered in %v", time.Since(t)) }()
	m.sb.Reset()
	m.sb.WriteString(m.cwd + "\n\n")

	grid := make([][]fs.DirEntry, 0, m.maxRows)
	maxColNameLen := make([]int, len(m.files)/m.maxRows+1)

	for i, f := range m.files {
		if i < m.maxRows {
			grid = append(grid, make([]fs.DirEntry, 0, maxRows))
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
		files, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		return dirLoaded{files: files, path: path}
	}
}

func (m Model) logCursor() {
	log.Debug().Msgf("cursor.r: %v, cursor.c: %v, maxRows: %v", m.cursor.r, m.cursor.c, m.maxRows)
}

func (m *Model) setCursor(r, c int) {
	m.cursor.r = r
	m.cursor.c = c
}

// trySelectFile tries to select a file by name or sets the cursor to the first file
// if the file is not found.
func (m *Model) trySelectFile(files []fs.DirEntry, fName string) {
	index := slices.IndexFunc(files, func(dir fs.DirEntry) bool {
		return dir.Name() == fName
	})

	if index != -1 {
		m.setCursor(index%m.maxRows, index/m.maxRows)
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
