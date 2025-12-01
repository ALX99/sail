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
	"github.com/alx99/sail/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		wd:        newPane(cwd, filelist.State{}, selection, true),
		pd:        newPane(parentDir, filelist.State{}, selection, false),
		cd:        newPane(cwd, filelist.State{}, selection, false),
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

		parentW, currentW, childW := v.calculatePaneWidths(v.termCols)

		// paneHeight is the full height available for the panes (excluding status bar)
		paneHeight := v.getFileHeight()

		// We set the bounds for the *content* inside the panes.
		// Since we use a standard border (takes 2 chars width/height),
		// we subtract 2 from the available width/height.
		// We also use parentW, currentW, childW as the full width of the pane block.

		v.pd.SetBounds(max(0, paneHeight), max(0, parentW))
		v.wd.SetBounds(max(0, paneHeight), max(0, currentW))
		v.cd.SetBounds(max(0, paneHeight), max(0, childW))

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
	parentW, currentW, childW := v.calculatePaneWidths(v.termCols)
	paneHeight := v.getFileHeight()

	// Parent Pane
	// Width/Height on the style sets the dimensions of the *block* (including borders)

	// IMPORTANT: Lipgloss Width/Height on a border style sets the CONTENT width/height if standard border is used?
	// "The Width and Height methods set the width and height of the block. This includes padding and borders."
	// If this is true, then Width(parentW) is correct.
	// And SetBounds(parentW-2) is correct.
	// Let's try Width(parentW).

	// Override styles to enforce size
	pStyle := theme.DefaultTheme.InactiveBorder.Width(parentW).Height(paneHeight)
	cStyle := theme.DefaultTheme.ActiveBorder.Width(currentW).Height(paneHeight)
	dStyle := theme.DefaultTheme.InactiveBorder.Width(childW).Height(paneHeight)

	parentView := pStyle.Render(v.pd.View())
	currentView := cStyle.Render(v.wd.View())

	var childView string
	if v.childEnabled {
		childView = dStyle.Render(v.cd.View())
	} else {
		childView = dStyle.Render("")
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parentView, currentView, childView)
}

func (v *Model) calculatePaneWidths(totalWidth int) (int, int, int) {
	// Reduce total width by 8 to account for borders (3 panes * 2 borders = 6)
	// plus a safety margin of 2 to avoid edge wrapping.
	totalWidth = max(0, totalWidth-8)

	// Ratio 1:2:3
	// Parent gets ~16%
	parentW := totalWidth / 6

	// Current gets ~33%
	currentW := (totalWidth / 6) * 2

	// Child gets remainder (~50%) to ensure full width usage
	childW := totalWidth - parentW - currentW

	// Minimum width checks (optional, to prevent crash or visual glitches)
	if parentW < 5 {
		parentW = 5
	}
	if currentW < 10 {
		currentW = 10
	}
	if childW < 10 {
		childW = 10
	}

	// Re-adjust child if we forced minimums to avoid overflow (simplified)
	if parentW+currentW+childW > totalWidth {
		// If overflow, prioritize Current > Parent > Child
		// This is a basic responsive strategy
		childW = max(0, totalWidth-parentW-currentW)
	}

	return parentW, currentW, childW
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
	// Return termRows - 3.
	// Width wrapping is fixed by subtracting borders in calculatePaneWidths.
	// We subtract 3 here: 1 for status bar, 2 for borders.
	h := v.termRows - 3
	if h < 3 {
		return 3
	}
	return h
}
