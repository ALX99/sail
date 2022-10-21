package mainview

import (
	"os"

	"github.com/alx99/fly/internal/ui/views/fileview"
	"github.com/alx99/fly/internal/util"
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

  fwWidth int
}

func New() mainView {
	fws := make([]fileview.Window, 3)
	home, ok := os.LookupEnv("HOME")
	if !ok {
		panic("$HOME not set")
	}

	fws[pd] = fileview.New(util.GetParentPath(home), 0, 0)
	fws[wd] = fileview.New(home, 0, 0)
	fws[cd] = fileview.New("todo", 0, 0)

	return mainView{fws: fws}
}

func (mw mainView) Init() tea.Cmd {
	return tea.Batch(mw.fws[0].Init, mw.fws[1].Init, mw.fws[2].Init, tea.EnterAltScreen)
}

func (mw mainView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mw.h, mw.w = msg.Height, msg.Width
		mw.fwWidth = msg.Width / 3
		mw.fws[pd].SetWidth(mw.fwWidth)
		mw.fws[wd].SetWidth(mw.fwWidth)
		mw.fws[cd].SetWidth(msg.Width - mw.fwWidth*2)

		util.Log.Debug().
			Int("height", msg.Height).
			Int("Width", msg.Width).
			Msg("Terminal size updated")

	case tea.KeyMsg:
		switch kp := msg.String(); kp {
		case "ctrl+c", "q":
			return mw, tea.Quit

		case "e":
			if mw.fws[wd].Move(fileview.Up).GetSelection().IsDir() {
				mw.fws[cd] = fileview.New(mw.fws[wd].GetSelectedPath(), mw.fwWidth, mw.h)
				return mw, mw.fws[cd].Init
			}
			return mw, nil

		case "n":
			if mw.fws[wd].Move(fileview.Down).GetSelection().IsDir() {
				mw.fws[cd] = fileview.New(mw.fws[wd].GetSelectedPath(), mw.fwWidth, mw.h)
				return mw, mw.fws[cd].Init
			}
			return mw, nil

		case "k":
			if mw.fws[pd].GetPath() == "/" {
				break
			}
			mw.fws[cd] = mw.fws[wd]
			mw.fws[wd] = mw.fws[pd]
			mw.fws[pd] = fileview.New(util.GetParentPath(mw.fws[pd].GetPath()), mw.w - mw.fwWidth*2, mw.h)
			return mw, mw.fws[pd].Init

		case "i":
			if !mw.fws[wd].GetSelection().IsDir() {
				break
			}
			mw.fws[pd] = mw.fws[wd]
			mw.fws[wd] = mw.fws[cd]
			mw.fws[cd] = fileview.New(mw.fws[wd].GetSelectedPath(), mw.w/3, mw.h)
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
