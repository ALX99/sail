package model

import (
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
)

type dirLoaded struct {
	path  string
	files []fs.DirEntry
}

type position struct {
	c, r int
}

type Model struct {
	cfg config.Config

	cwd     string        // current working directory
	files   []fs.DirEntry // current files in that directory
	maxRows int

	cursor    position            // cursor
	positions map[string]position // previous positions

	// for performance purposes
	sb strings.Builder
}

func New(cwd string, cfg config.Config) Model {
	return Model{
		cwd:       cwd,
		cfg:       cfg,
		maxRows:   10,
		positions: make(map[string]position, 100),
		sb:        strings.Builder{},
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadDir(m.cwd)
}

func (m Model) logCursor() {
	log.Debug().Msgf("cursor.r: %v, cursor.c: %v, maxRows: %v", m.cursor.r, m.cursor.c, m.maxRows)
}

func (m Model) cursorOffset() int {
	m.logCursor()
	return m.cursor.c*m.maxRows + m.cursor.r
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case m.cfg.Settings.Keybinds.NavUp:
			m.cursor.r = max(0, m.cursor.r-1)
		case m.cfg.Settings.Keybinds.NavDown:
			if m.cursor.c < len(m.files)/m.maxRows {
				m.cursor.r = min(m.cursor.r+1, min(len(m.files), m.maxRows)-1)
			} else if m.cursor.c == len(m.files)/m.maxRows {
				m.cursor.r = min(m.cursor.r+1, len(m.files)%m.maxRows-1)
			}
			return m, nil

		case m.cfg.Settings.Keybinds.NavLeft:
			m.cursor.c = max(0, m.cursor.c-1)
			return m, nil
		case m.cfg.Settings.Keybinds.NavRight:
			m.cursor.c++
			if m.cursorOffset() >= len(m.files) {
				m.cursor.c-- // undo the cursor move
			}
			return m, nil
		case m.cfg.Settings.Keybinds.NavOut:
			return m, m.loadDir(path.Dir(m.cwd))
		case m.cfg.Settings.Keybinds.NavIn:
			if m.files != nil && m.files[m.cursorOffset()].IsDir() {
				return m, m.loadDir(path.Join(m.cwd, m.files[m.cursorOffset()].Name()))
			}

		}
	case dirLoaded:
		m.positions[m.cwd] = m.cursor // save old cursor pos

		if cursor, ok := m.positions[msg.path]; ok {
			log.Debug().Msgf("cache hit for %v: cursor.r: %v, cursor.c: %v", msg.path, cursor.r, cursor.c)
			m.cursor = cursor

			// sanity check in case the files has decreased
			if m.cursorOffset() >= len(msg.files) {
				m.cursor.r = 0
				m.cursor.c = 0
			}
		} else {
			log.Debug().Msgf("cache miss for %v", msg.path)

			prevDir := strings.TrimLeft(m.cwd, path.Dir(m.cwd))

			index := slices.IndexFunc(msg.files, func(dir fs.DirEntry) bool {
				return dir.Name() == prevDir
			})

			if index != -1 {
				m.cursor.c = index / m.maxRows
				m.cursor.r = index % m.maxRows
			} else {
				m.cursor.c = 0
				m.cursor.r = 0
			}

		}

		m.cwd = msg.path
		m.files = msg.files

		return m, nil
	case error:
		log.Error().Err(msg).Msg("Error occurred")
	}

	return m, nil
}

func (m Model) View() string {
	m.sb.Reset()
	m.sb.WriteString(m.cwd + "\n\n")

	grid := make([][]fs.DirEntry, 0, m.maxRows)
	maxColNameLen := make([]int, len(m.files)/m.maxRows+1)

	for i, f := range m.files {
		if i < m.maxRows {
			grid = append(grid, make([]fs.DirEntry, 0, 10))
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
