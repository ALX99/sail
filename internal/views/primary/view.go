package primary

// todo there is an issue with navigation if the directory takes long to load
// due to that the cwd is only updated after a directory is loaded
// for example navigating up 2 directories super fast, might only result in 1 navigation
// to fix this, the cwd needs to be updated before the directory is even loaded.
// when a directory has loaded, we also need to check it is the directory we want, since
// things can arrive out of order.
// however, this is quite an edge case, so I will leave it as is for now.

import (
	"cmp"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alx99/sail/internal/config"
	"github.com/alx99/sail/internal/filesys"
	"github.com/alx99/sail/internal/views/filelist"
	"github.com/alx99/sail/internal/views/status"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

var (
	errShowDuration = time.Second
	primaryColor    = lipgloss.Color("#6462e3")
	pStyle          = lipgloss.NewStyle().Foreground(primaryColor)
)

type dirLoaded struct {
	dir        filesys.Dir
	fileSelect string
	parentDir  filesys.Dir
}

type childLoaded struct {
	dir filesys.Dir
}

type View struct {
	cfg config.Config

	cwd            string // current working directory
	cachedDirState map[string]filelist.State
	fs             *filesys.FS
	lastError      error     // last error that occurred
	clearErrAt     time.Time // time when the error should be cleared
	termCols       int       // max width of the terminal window
	termRows       int       // max height of the terminal window
	altScreen      bool      // whether to use the alternate screen

	pd, wd, cd filelist.View
	status     status.View
}

func New(cwd string, cfg config.Config) View {
	parentDir := filepath.Dir(cwd)
	fsys := filesys.NewFS()
	v := View{
		wd:             filelist.New(cwd, filelist.State{}, fsys, true, primaryColor),
		pd:             filelist.New(parentDir, filelist.State{}, fsys, true, primaryColor),
		cd:             filelist.New(cwd, filelist.State{}, fsys, false, primaryColor),
		cwd:            cwd,
		cfg:            cfg,
		status:         status.New(),
		cachedDirState: make(map[string]filelist.State, 100),
		fs:             fsys,
	}

	return v
}

func (v View) Init() tea.Cmd {
	return tea.Batch(v.loadDirCmd(v.cwd, ""), v.status.Init())
}

func (v View) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// clear the last error
	if v.lastError != nil && time.Now().After(v.clearErrAt) {
		v.lastError = nil
	}

	var cmds []tea.Cmd
	v.status.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if v.cfg.PrintLastWD != "" {
				err := v.writeLastWD()
				if err != nil {
					log.Error().Err(err).Send()
				}
			}
			return v, tea.Quit

		case v.cfg.Settings.Keymap.NavUp:
			v.wd.MoveUp()
			return v, v.loadChildDir()

		case v.cfg.Settings.Keymap.NavDown:
			v.wd.MoveDown()
			return v, v.loadChildDir()

		case v.cfg.Settings.Keymap.NavLeft:
			if v.wd.Path() == "/" {
				return v, v.status.SetError(errors.New("can't navigate up from root"))
			}
			return v, v.loadAndSelectDir(filepath.Dir(v.wd.Path()), filepath.Base(v.wd.Path()))

		case v.cfg.Settings.Keymap.NavRight:
			e, ok := v.wd.CurrEntry()
			if !ok {
				return v, nil
			}

			e, err := e.ResolveSymlink()
			if err != nil {
				cmds = append(cmds, v.status.SetError(err))
				return v, tea.Batch(cmds...)
			}

			if e.IsDir() {
				return v, v.loadDir(e.Path())
			}

			return v, nil

		case v.cfg.Settings.Keymap.NavHome:
			home, err := os.UserHomeDir()
			if err != nil {
				cmds = append(cmds, v.status.SetError(err))
				return v, tea.Batch(cmds...)
			}
			return v, v.loadDir(home)

		case v.cfg.Settings.Keymap.Delete:
			return v, tea.Sequence(
				func() tea.Msg {
					return v.fs.DeleteSelections()
				},
				v.loadDir(v.cwd),
			)

		case v.cfg.Settings.Keymap.Cut:
			return v, tea.Sequence(
				func() tea.Msg {
					return v.fs.MoveSelections(v.cwd)
				},
				v.loadDir(v.cwd),
			)

		case v.cfg.Settings.Keymap.Copy:
			return v, tea.Sequence(
				func() tea.Msg {
					return v.fs.CopySelections(v.cwd)
				},
				v.loadDir(v.cwd),
			)

		case v.cfg.Settings.Keymap.Select:
			e, ok := v.wd.CurrEntry()
			if !ok {
				return v, nil
			}

			path := e.Path()

			if v.fs.IsSelected(path) {
				v.fs.Deselect(path)
			} else {
				v.fs.Select(path)
			}
			v.wd.MoveDown()

			return v, nil

		case v.cfg.Settings.Keymap.ToggleAltScreen:
			if v.altScreen {
				v.altScreen = !v.altScreen
				return v, tea.ExitAltScreen
			}
			v.altScreen = !v.altScreen
			return v, tea.EnterAltScreen
		}

	case tea.WindowSizeMsg:
		v.termCols = msg.Width
		v.termRows = msg.Height

		v.pd.SetMaxDims(v.getFileHeight(), v.getParentFileWidth())
		v.wd.SetMaxDims(v.getFileHeight(), v.getFileWidth())
		v.cd.SetMaxDims(v.getFileHeight()-2, v.getChildFileWidth()) // - 2 for the top and bottom borders
		v.status.SetWidth(msg.Width)

		return v, nil

	case dirLoaded:
		oldDir := v.cwd
		v.cwd = msg.dir.Path()

		cmds = append(cmds, v.status.SetWD(msg.dir))

		// Cache state of the directory we are leaving
		v.cachedDirState[oldDir] = v.wd.State()

		// Get cached state for the new directory, applying file selection if provided
		state := v.cachedDirState[v.cwd]
		state.SelectedName = cmp.Or(msg.fileSelect, state.SelectedName)

		v.wd.ChDir(msg.dir, state)

		// Select the parent directory in the parent list
		state = filelist.State{
			SelectedName: filepath.Base(v.wd.Path()),
		}

		v.pd.ChDir(msg.parentDir, state)

		cmds = append(cmds, v.loadChildDir())

		return v, tea.Batch(cmds...)

	case childLoaded:
		v.cachedDirState[v.cd.Path()] = v.cd.State()

		v.cd.ChDir(msg.dir, v.cachedDirState[msg.dir.Path()])

		return v, nil
	case error:
		cmds = append(cmds, v.status.SetError(msg))
		log.Error().Err(msg).Msg("Error occurred")
		return v, tea.Batch(cmds...)
	}

	// Default case if no message was handled
	return v, nil
}

func (v View) View() string {
	parentList, childList, firstBorder := "", "", ""
	list := v.wd.View()

	// TODO, find better way to handle this.
	// This is a hack, but we are already at root
	// we should probably increase size of the child list
	if v.wd.Path() != "/" {
		parentList = v.pd.View()
		firstBorder = firstSnakeLine(
			v.pd.SelectedColum(),
			v.wd.SelectedColum(),
			lipgloss.RoundedBorder(),
		)
	}

	if e, ok := v.wd.CurrEntry(); ok {
		// Unfortunate IO during render when e is a symlink.
		// TODO: Find a better way to handle this.
		if e, err := e.ResolveSymlink(); err == nil && e.IsDir() {
			childList = v.cd.View()
		}
	}

	secondBorder := v.secondSnakeLine()

	fileView := lipgloss.JoinHorizontal(lipgloss.Left,
		parentList,
		firstBorder,
		list,
		secondBorder,
		lipgloss.NewStyle().BorderForeground(pStyle.GetForeground()).Border(lipgloss.RoundedBorder(), true, true, true, false).
			Height(v.getFileHeight()-2). // - 2 for the top and bottom borders
			Width(v.getChildFileWidth()).
			Render(
				childList,
			),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		fileView,
		v.status.View())
}

// loadDirCmd creates a command to load both the target directory and its parent.
func (v View) loadDirCmd(targetPath, fileSelect string) tea.Cmd {
	return func() tea.Msg {
		dir, err := filesys.NewDir(targetPath)
		if err != nil {
			return err
		}

		parentPath := filepath.Dir(targetPath)
		parentDir, err := filesys.NewDir(parentPath)
		if err != nil {
			return err
		}

		return dirLoaded{
			dir:        dir,
			fileSelect: fileSelect,
			parentDir:  parentDir,
		}
	}
}

// loadDir loads both the target directory and its parent.
func (v View) loadDir(path string) tea.Cmd {
	return v.loadDirCmd(path, "")
}

func (v View) loadChildDir() tea.Cmd {
	e, ok := v.wd.CurrEntry()
	if !ok {
		return nil
	}

	return func() tea.Msg {
		e, err := e.ResolveSymlink()
		if err != nil {
			return err
		}

		if !e.IsDir() {
			return nil
		}

		dir, err := filesys.NewDir(e.Path())
		if err != nil {
			return err
		}

		return childLoaded{
			dir: dir,
		}
	}
}

// loadAndSelectDir loads both the target directory (selecting a file) and its parent.
func (v View) loadAndSelectDir(path, fileSelect string) tea.Cmd {
	return v.loadDirCmd(path, fileSelect)
}

func (v View) writeLastWD() error {
	f, err := os.Create(v.cfg.PrintLastWD)
	if err != nil {
		return err
	}

	defer f.Close()
	_, err = f.WriteString(v.cwd)
	return err
}

func (v View) getFileHeight() int {
	return v.termRows - 1 - 2 // - 2 for the status bar
}

func (v View) getFileWidth() int {
	return (v.termCols / 6) * 2
}

func (v View) getParentFileWidth() int {
	return v.termCols / 6
}

func (v View) getChildFileWidth() int {
	return v.termCols - (v.getParentFileWidth() + v.getFileWidth() + 3)
}

func firstSnakeLine(leftSel, rightSel int, border lipgloss.Border) string {
	var sb strings.Builder

	// If selections are aligned, draw a simple top connector
	if leftSel == rightSel {
		return pStyle.Render(strings.Repeat("\n", leftSel) + border.Top)
	}

	// Determine border characters and dimensions
	var topCorner, bottomCorner string
	var startOffset, height int

	verticalLine := border.Right

	if leftSel < rightSel { // Selection is higher in left list
		topCorner = border.TopRight
		bottomCorner = border.BottomLeft
		startOffset = leftSel
		height = rightSel - leftSel
	} else { // Selection is lower in left list
		topCorner = border.TopLeft
		bottomCorner = border.BottomRight
		startOffset = rightSel
		height = leftSel - rightSel
	}

	sb.Grow(startOffset + height*2 + 2)

	sb.WriteString(strings.Repeat("\n", startOffset))
	sb.WriteString(topCorner)
	sb.WriteString("\n")

	for range height - 1 {
		sb.WriteString(verticalLine)
		sb.WriteString("\n")
	}

	sb.WriteString(bottomCorner)
	return pStyle.Render(sb.String())
}

func (v View) secondSnakeLine() string {
	var sb strings.Builder

	if v.wd.SelectedColum() == 0 {
		sb.WriteString(lipgloss.RoundedBorder().MiddleTop)
	} else {
		sb.WriteString(lipgloss.RoundedBorder().TopLeft)
	}
	sb.WriteString("\n")

	if v.wd.SelectedColum() == 0 {
		sb.WriteString(lipgloss.RoundedBorder().Left)
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat(lipgloss.RoundedBorder().Left+"\n", max(0, v.wd.SelectedColum()-1)))

	if v.wd.SelectedColum() == v.getFileHeight()-1 {
		sb.WriteString(lipgloss.RoundedBorder().MiddleBottom)
		return pStyle.Render(sb.String())
	}

	if v.wd.SelectedColum() != 0 {
		sb.WriteString(lipgloss.RoundedBorder().MiddleRight)
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat(lipgloss.RoundedBorder().Left+"\n",
		max(0, v.getFileHeight()-lipgloss.Height(sb.String()))))

	sb.WriteString(lipgloss.RoundedBorder().BottomLeft)

	return pStyle.Render(sb.String())
}
