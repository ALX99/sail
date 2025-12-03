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

	"github.com/alx99/sail/internal/collator"
	"github.com/alx99/sail/internal/config"
	"github.com/alx99/sail/internal/filesys"
	"github.com/alx99/sail/internal/style"
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

	pd, wd, cd    *pane
	childEnabled  bool
	parentEnabled bool
	showHidden    bool

	wdReqID    int
	childReqID int
}

func New(cwd string, cfg config.Config, styles *style.Styles) *Model {
	parentDir := filepath.Dir(cwd)
	selection := filesys.NewSelection()
	coll := collator.New()
	v := &Model{
		wd:            newPane(cwd, filelist.State{}, coll, selection, true, styles),
		pd:            newPane(parentDir, filelist.State{}, coll, selection, false, styles),
		cd:            newPane(cwd, filelist.State{}, coll, selection, false, styles),
		cwd:           cwd,
		cfg:           cfg,
		selection:     selection,
		parentEnabled: true,
		showHidden:    false,
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

		case v.cfg.Settings.Keymap.ToggleParentPane:
			v.parentEnabled = !v.parentEnabled
			v.updateLayout()
			return v, nil

		case v.cfg.Settings.Keymap.ToggleMinimalUI:
			v.cfg.Settings.MinimalUI = !v.cfg.Settings.MinimalUI
			v.updateLayout()
			return v, nil

		case v.cfg.Settings.Keymap.ToggleHidden:
			v.showHidden = !v.showHidden
			v.pd.SetShowHidden(v.showHidden)
			v.wd.SetShowHidden(v.showHidden)
			v.cd.SetShowHidden(v.showHidden)
			return v, nil
		}

	case tea.WindowSizeMsg:
		v.termCols = msg.Width
		v.termRows = msg.Height

		v.updateLayout()

		return v, nil

	case filesys.FilesDeletedMsg, filesys.FilesMovedMsg, filesys.FilesCopiedMsg:
		v.selection.Clear()
		return v, v.loadDir(v.cwd)

	case filesys.DirLoadedMsg:
		if msg.ReqID != v.wdReqID {
			return v, nil
		}

		v.wd.RememberCurrent()
		wasRoot := v.cwd == filepath.Dir(v.cwd)
		v.cwd = msg.Dir.Path()

		v.wd.SetDir(msg.Dir, filelist.State{
			SelectedName: msg.SelectName,
		})

		// Select the parent directory in the parent list
		if msg.ParentDir.Path() == msg.Dir.Path() {
			v.pd.SetDir(filesys.Dir{}, filelist.State{})
		} else {
			v.pd.SetDir(msg.ParentDir, filelist.State{
				SelectedName: filepath.Base(v.wd.Path()),
			})
		}

		isRoot := v.cwd == filepath.Dir(v.cwd)
		if wasRoot != isRoot {
			v.updateLayout()
		}

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

	pStyle, cStyle, dStyle := v.paneStyles(parentW, currentW, childW, paneHeight)

	currentView := cStyle.Render(v.wd.View())

	var childView string
	if v.childEnabled {
		childView = dStyle.Render(v.cd.View())
	} else {
		childView = dStyle.Render("")
	}

	isRoot := v.cwd == filepath.Dir(v.cwd)
	if v.parentEnabled && !isRoot {
		parentView := pStyle.Render(v.pd.View())
		return lipgloss.JoinHorizontal(lipgloss.Top, parentView, currentView, childView)
	} else {
		return lipgloss.JoinHorizontal(lipgloss.Top, currentView, childView)
	}
}

func (v *Model) calculatePaneWidths(totalWidth int) (int, int, int) {
	isRoot := v.cwd == filepath.Dir(v.cwd)

	if !v.parentEnabled || isRoot {
		// 2 panes: Current, Child
		width := max(0, totalWidth-v.borderDeduction(2))

		// Ratio 2:3 (Current:Child)
		// Current gets 40%
		currentW := (width * 2) / 5

		// Child gets 60%
		childW := width - currentW

		if currentW < 10 {
			currentW = 10
		}
		if childW < 10 {
			childW = 10
		}

		if currentW+childW > width {
			childW = max(0, width-currentW)
		}

		return 0, currentW, childW
	}

	// 3 panes
	width := max(0, totalWidth-v.borderDeduction(3))

	// Ratio 1:2:3
	// Parent gets ~16%
	parentW := width / 6

	// Current gets ~33%
	currentW := (width / 6) * 2

	// Child gets remainder (~50%) to ensure full width usage
	childW := width - parentW - currentW

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
	if parentW+currentW+childW > width {
		// If overflow, prioritize Current > Parent > Child
		// This is a basic responsive strategy
		childW = max(0, width-parentW-currentW)
	}

	return parentW, currentW, childW
}

func (v *Model) paneStyles(parentW, currentW, childW, paneHeight int) (lipgloss.Style, lipgloss.Style, lipgloss.Style) {
	if v.cfg.Settings.MinimalUI {
		divider := theme.DefaultTheme.MinimalDivider
		return divider.Width(parentW).Height(paneHeight),
			divider.Width(currentW).Height(paneHeight),
			lipgloss.NewStyle().Width(childW).Height(paneHeight)
	}

	return theme.DefaultTheme.InactiveBorder.Width(parentW).Height(paneHeight),
		theme.DefaultTheme.ActiveBorder.Width(currentW).Height(paneHeight),
		theme.DefaultTheme.InactiveBorder.Width(childW).Height(paneHeight)
}

func (v *Model) borderDeduction(paneCount int) int {
	if v.cfg.Settings.MinimalUI {
		// Minimal UI draws only the dividers between panes.
		// Each divider is 1 column wide.
		return max(0, (paneCount - 1))
	}

	// Standard UI uses left and right borders on each pane (2 columns per pane)
	return paneCount * 2
}

func (v *Model) heightDeduction() int {
	if v.cfg.Settings.MinimalUI {
		// Minimal UI uses dividers only; no vertical chrome to deduct.
		return 0
	}

	// Top and bottom borders of each pane.
	return 2
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
	h := v.termRows - v.heightDeduction()
	if h < 3 {
		return 3
	}
	return h
}

func (v *Model) updateLayout() {
	parentW, currentW, childW := v.calculatePaneWidths(v.termCols)
	paneHeight := v.getFileHeight()

	v.pd.SetBounds(max(0, paneHeight), max(0, parentW))
	v.wd.SetBounds(max(0, paneHeight), max(0, currentW))
	v.cd.SetBounds(max(0, paneHeight), max(0, childW))
}
