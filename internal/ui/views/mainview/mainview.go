package mainview

import (
	"github.com/alx99/fly/internal/ui/views/fileview"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	pd = iota // parent directory
	wd        // working directoy
	cd        // child directory
)

type mainView struct {
	fws  []fileview.Window
	h, w int
}

func New() mainView {
	fws := make([]fileview.Window, 3)
	fws[0] = fileview.New("/", 0, 0)
	fws[1] = fileview.New("/var", 0, 0)
	fws[2] = fileview.New("/var/lib", 0, 0)

	return mainView{fws: fws}
}

func (mw mainView) Init() tea.Cmd {
	return tea.Batch(mw.fws[0].Init, mw.fws[1].Init, mw.fws[2].Init)
}

func (mw mainView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mw.h, mw.w = msg.Height, msg.Width

	case tea.KeyMsg:
		switch kp := msg.String(); kp {
		case "ctrl+c", "q":
			return mw, tea.Quit

		case "e":
			if mw.fws[wd].Move(fileview.Up).GetSelection().IsDir() {
				mw.fws[cd] = fileview.New(mw.fws[wd].GetSelectedPath(), mw.w, mw.h)
				return mw, mw.fws[2].Init
			}
			return mw, nil

		case "n":
			if mw.fws[wd].Move(fileview.Down).GetSelection().IsDir() {
				mw.fws[cd] = fileview.New(mw.fws[wd].GetSelectedPath(), mw.w, mw.h)
				return mw, mw.fws[2].Init
			}
			return mw, nil

		case "k":
			if mw.fws[pd].GetPath() == "/" {
				break
			}
			mw.fws[cd] = mw.fws[wd]
			mw.fws[wd] = mw.fws[pd]
			mw.fws[pd] = fileview.New(util.GetParentPath(mw.fws[pd].GetPath()), mw.w, mw.h)
			return mw, mw.fws[pd].Init

		case "i":
			if !mw.fws[wd].GetSelection().IsDir() {
				break
			}
			mw.fws[pd] = mw.fws[wd]
			mw.fws[wd] = mw.fws[cd]
			mw.fws[cd] = fileview.New(mw.fws[wd].GetSelectedPath(), mw.w, mw.h)
			return mw, mw.fws[cd].Init
		}
	}

	for i := range mw.fws {
		mw.fws[i], _ = mw.fws[i].Update(msg)
	}

	return mw, nil
}

func (mw mainView) View() string {
	res := make([]string, len(mw.fws))
	for _, fw := range mw.fws {
		res = append(res, fw.View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, res...)
}
func (mw mainView) logState() {
	util.Log.Debug().
		Str("pd", mw.fws[pd].GetPath()).
		Str("wd", mw.fws[wd].GetPath()).
		Str("cd", mw.fws[cd].GetPath()).
		Send()
}
