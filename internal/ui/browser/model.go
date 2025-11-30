package browser

// todo there is an issue with navigation if the directory takes long to load
// due to that the cwd is only updated after a directory is loaded
// for example navigating up 2 directories super fast, might only result in 1 navigation
// to fix this, the cwd needs to be updated before the directory is even loaded.
// when a directory has loaded, we also need to check it is the directory we want, since
// things can arrive out of order.
// however, this is quite an edge case, so I will leave it as is for now.

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/alx99/sail/internal/config"
	"github.com/alx99/sail/internal/filesys"
	"github.com/alx99/sail/internal/ui/components/filelist"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	primaryColor = lipgloss.Color("#833f8f")
	pStyle       = lipgloss.NewStyle().Foreground(primaryColor)
	border       = lipgloss.RoundedBorder()
)

type Model struct {
	cfg config.Config

	cwd       string // current working directory
	selection *filesys.Selection

	termCols int // max width of the terminal window
	termRows int // max height of the terminal window

	pd, wd, cd   *pane
	childEnabled bool

	wdReqID    int
	childReqID int
}

func New(cwd string, cfg config.Config) *Model {
	parentDir := filepath.Dir(cwd)
	selection := filesys.NewSelection()
	v := &Model{
		wd:        newPane(cwd, filelist.State{}, selection, true, primaryColor),
		pd:        newPane(parentDir, filelist.State{}, selection, true, primaryColor),
		cd:        newPane(cwd, filelist.State{}, selection, false, primaryColor),
		cwd:       cwd,
		cfg:       cfg,
		selection: selection,
	}

	return v
}

func (v *Model) Init() tea.Cmd {
	return v.loadDir(v.cwd)
}

func (v *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case v.cfg.Settings.Keymap.NavUp:
			v.wd.MoveUp()
			return v, v.loadChildDir()

		case v.cfg.Settings.Keymap.NavDown:
			v.wd.MoveDown()
			return v, v.loadChildDir()

		case v.cfg.Settings.Keymap.NavLeft:
			if v.wd.Path() == "/" {
				return v, errorCmd(errors.New("can't navigate up from root"))
			}
			return v, v.loadDirWithSelection(filepath.Dir(v.wd.Path()), filepath.Base(v.wd.Path()))

		case v.cfg.Settings.Keymap.NavRight:
			e, ok := v.wd.CurrEntry()
			if !ok {
				return v, nil
			}

			e, err := e.ResolveSymlink()
			if err != nil {
				return v, errorCmd(err)
			}

			if e.IsDir() {
				return v, v.loadDir(e.Path())
			}

			return v, nil

		case v.cfg.Settings.Keymap.NavHome:
			home, err := os.UserHomeDir()
			if err != nil {
				return v, errorCmd(err)
			}
			return v, v.loadDir(home)

		case v.cfg.Settings.Keymap.Delete:
			paths := v.selection.Paths()
			if len(paths) == 0 {
				return v, nil
			}
			return v, filesys.DeleteCmd(paths)

		case v.cfg.Settings.Keymap.Cut:
			paths := v.selection.Paths()
			if len(paths) == 0 {
				return v, nil
			}
			return v, filesys.MoveCmd(paths, v.cwd)

		case v.cfg.Settings.Keymap.Copy:
			paths := v.selection.Paths()
			if len(paths) == 0 {
				return v, nil
			}
			return v, filesys.CopyCmd(paths, v.cwd)

		case v.cfg.Settings.Keymap.Select:
			e, ok := v.wd.CurrEntry()
			if !ok {
				return v, nil
			}

			path := e.Path()

			v.selection.Toggle(path)
			v.wd.MoveDown()

			return v, nil
		}

	case tea.WindowSizeMsg:
		v.termCols = msg.Width
		v.termRows = msg.Height

		parentVisible := v.wd.Path() != "/"
		v.pd.SetBounds(v.getFileHeight(), v.getParentFileWidth())
		v.wd.SetBounds(v.getFileHeight(), v.getFileWidth())
		v.cd.SetBounds(v.getFileHeight()-2, v.getChildFileWidth(parentVisible)) // - 2 for the top and bottom borders

		return v, nil

	case filesys.FilesDeletedMsg, filesys.FilesMovedMsg, filesys.FilesCopiedMsg:
		v.selection.Clear()
		return v, v.loadDir(v.cwd)

	case filesys.DirLoadedMsg:
		if msg.ReqID != v.wdReqID {
			return v, nil
		}

		v.wd.RememberCurrent()
		v.cwd = msg.Dir.Path()

		v.wd.SetDir(msg.Dir, filelist.State{
			SelectedName: msg.SelectName,
		})

		// Select the parent directory in the parent list
		v.pd.SetDir(msg.ParentDir, filelist.State{
			SelectedName: filepath.Base(v.wd.Path()),
		})

		v.childEnabled = false
		return v, v.loadChildDir()

	case filesys.ChildLoadedMsg:
		if msg.ReqID != v.childReqID {
			return v, nil
		}

		v.cd.RememberCurrent()
		v.cd.SetDir(msg.Dir, filelist.State{})
		v.childEnabled = true

		return v, nil
	case error:
		// Let the parent handle status updates for errors.
		return v, nil
	}

	// Default case if no message was handled
	return v, nil
}

func (v *Model) View() string {
	parentVisible := v.wd.Path() != "/"
	parentList, parentConnector := "", ""
	if parentVisible {
		parentList = v.pd.View()
		parentConnector = renderParentConnector(v.getFileHeight(), v.pd.SelectedRow(), v.wd.SelectedRow())
	}

	childPane := ""
	if v.childEnabled {
		childPane = lipgloss.NewStyle().
			BorderForeground(pStyle.GetForeground()).
			Border(border, true, true, true, false).
			Height(v.getFileHeight() - 2).
			Width(v.getChildFileWidth(parentVisible)).
			Render(v.cd.View())
	}

	secondBorder := ""
	if v.childEnabled {
		secondBorder = renderChildConnector(v.getFileHeight(), v.wd.SelectedRow())
	}

	view := []string{
		parentList,
		parentConnector,
		v.wd.View(),
		secondBorder,
		childPane,
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, view...)
}

func (v *Model) CWD() string {
	return v.cwd
}

func (v *Model) loadDir(path string) tea.Cmd {
	return v.loadDirWithSelection(path, "")
}

func (v *Model) loadDirWithSelection(path, selectName string) tea.Cmd {
	v.wdReqID++
	return filesys.LoadDirCmd(v.wdReqID, path, selectName)
}

func (v *Model) loadChildDir() tea.Cmd {
	e, ok := v.wd.CurrEntry()
	if !ok {
		return nil
	}

	resolved, err := e.ResolveSymlink()
	if err != nil {
		v.childEnabled = false
		return errorCmd(err)
	}

	if !resolved.IsDir() {
		v.childEnabled = false
		return nil
	}

	v.childEnabled = true
	v.childReqID++
	return filesys.LoadChildCmd(v.childReqID, resolved.Path())
}

func errorCmd(err error) tea.Cmd {
	if err == nil {
		return nil
	}
	return func() tea.Msg {
		return err
	}
}

func (v *Model) getFileHeight() int {
	return v.termRows - 1 - 2 // - 2 for the status bar
}

func (v *Model) getFileWidth() int {
	return (v.termCols / 6) * 2
}

func (v *Model) getParentFileWidth() int {
	return v.termCols / 6
}

func (v *Model) getChildFileWidth(parentVisible bool) int {
	base := v.termCols - (v.getFileWidth() + 3)
	if parentVisible {
		return base - v.getParentFileWidth()
	}
	return base
}
