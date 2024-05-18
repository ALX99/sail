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
	numRows             int               // the number of rows to display

	// for performance purposes
	sb strings.Builder
}

func NewMain(cwd string, cfg config.Config) Model {
	return Model{
		cwd:                 cwd,
		cfg:                 cfg,
		numRows:             maxRows,
		cachedDirSelections: make(map[string]string, 100),
		sb:                  strings.Builder{},
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadDir(m.cwd)
}

func (m Model) cursorOffset() int {
	// m.logCursor()
	return m.cursor.c*m.numRows + m.cursor.r
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case m.cfg.Settings.Keymap.NavUp:
			m.cursor.r = max(0, m.cursor.r-1)
		case m.cfg.Settings.Keymap.NavDown:
			if m.cursor.c < len(m.files)/m.numRows {
				m.cursor.r = min(m.cursor.r+1, min(len(m.files), m.numRows)-1)
			} else if m.cursor.c == len(m.files)/m.numRows {
				m.cursor.r = min(m.cursor.r+1, len(m.files)%m.numRows-1)
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

		m.numRows = min(maxRows, max(1, msg.Height-3))

		m.trySelectFile(m.files, fName)

		return m, nil

	case dirLoaded:
		if len(m.files) > 0 {
			// cache the selected file for the previous directory
			m.cachedDirSelections[m.cwd] = m.files[m.cursorOffset()].Name()
		}

		fName, ok := m.cachedDirSelections[msg.path]
		if !ok {
			// try to determine the previous file name
			fName = path.Base(m.cwd)
		}

		m.cwd = msg.path
		m.files = msg.files
		m.trySelectFile(m.files, fName)

		return m, nil
	case error:
		log.Error().Err(msg).Msg("Error occurred")
	}

	return m, nil
}

func (m Model) View() string {
	t := time.Now()
	defer func() { log.Debug().Msgf("Rendered in %v", time.Since(t)) }()
	m.sb.Reset()
	m.sb.WriteString(m.cwd + "\n\n")

	grid := make([][]fs.DirEntry, 0, m.numRows)
	maxColNameLen := make([]int, len(m.files)/m.numRows+1)

	for i, f := range m.files {
		if i < m.numRows {
			grid = append(grid, make([]fs.DirEntry, 0, maxRows))
		}
		r, c := i%m.numRows, i/m.numRows
		maxColNameLen[c] = max(maxColNameLen[c], len(f.Name()))

		grid[r] = append(grid[r], f)
		maxColNameLen[len(grid[r])-1] = max(maxColNameLen[len(grid[r])-1], len(f.Name()))
	}

	for row := range len(grid) {
		for col, f := range grid[row] {
			if m.cursor.r == row && m.cursor.c == col {
				m.sb.WriteString(">")
			}
			extraPadding := maxColNameLen[col] - len(f.Name()) + 2

			if m.cursor.r == row && m.cursor.c == col {
				extraPadding--
			}
			m.sb.WriteString(util.GetStyle(f).
				PaddingRight(extraPadding).
				Render(f.Name()))
		}
		m.sb.WriteString("\n")
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
	log.Debug().Msgf("cursor.r: %v, cursor.c: %v, maxRows: %v", m.cursor.r, m.cursor.c, m.numRows)
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
		m.setCursor(index%m.numRows, index/m.numRows)
	} else {
		m.setCursor(0, 0)
	}
}
