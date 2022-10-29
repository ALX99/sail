package mainview

import (
	"os"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/ui/views/fileview"
	"github.com/alx99/fly/internal/ui/views/inputview"
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
	fws []fileview.Window
	iv  inputview.View

	h, w     int
	fwWidth  int
	fwHeight int

	cfg config.Config
}

func New(cfg config.Config) mainView {
	fws := make([]fileview.Window, 3)
	home, ok := os.LookupEnv("HOME")
	if !ok {
		panic("$HOME not set")
	}

	fws[pd] = fileview.New(util.GetParentPath(home), 0, 0, cfg)
	fws[wd] = fileview.New(home, 0, 0, cfg)
	fws[cd] = fileview.New("todo", 0, 0, cfg)

	return mainView{fws: fws, cfg: cfg}
}

func (mw mainView) Init() tea.Cmd {
	return tea.Batch(mw.fws[0].Init, mw.fws[1].Init, mw.fws[2].Init, tea.EnterAltScreen)
}

func (mw mainView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mw.h, mw.w = msg.Height, msg.Width

		mw.fwWidth = msg.Width / 3 // width
		mw.fwHeight = msg.Height
		if mw.iv.Focused() {
			mw.fwHeight -= 1
		}

		mw.updateFWSizes()

		util.Log.Debug().
			Int("height", msg.Height).
			Int("Width", msg.Width).
			Msg("Terminal size updated")

	case tea.KeyMsg:
		if mw.iv.Focused() {
			mw.iv, _ = mw.iv.Update(msg)

			if !mw.iv.Focused() {
				mw.fwHeight++ // inpuvtview is no longer visible
				mw.updateFWSizes()
			}

			return mw, nil
		}

		switch kp := msg.String(); kp {
		case "ctrl+c", "q":
			return mw, tea.Quit

		case "e":
			if mw.fws[wd].Move(fileview.Up).GetSelection().IsDir() {
				mw.fws[cd] = fileview.New(mw.fws[wd].GetSelectedPath(), mw.fwWidth, mw.fwHeight, mw.cfg)
				return mw, mw.fws[cd].Init
			}
			return mw, nil

		case "n":
			if mw.fws[wd].Move(fileview.Down).GetSelection().IsDir() {
				mw.fws[cd] = fileview.New(mw.fws[wd].GetSelectedPath(), mw.fwWidth, mw.fwHeight, mw.cfg)
				return mw, mw.fws[cd].Init
			}
			return mw, nil

		case "k":
			if mw.fws[pd].GetPath() == "/" {
				break
			}
			mw.fws[cd] = mw.fws[wd]
			mw.fws[wd] = mw.fws[pd]
			mw.fws[pd] = fileview.New(util.GetParentPath(mw.fws[pd].GetPath()), mw.w-mw.fwWidth*2, mw.h, mw.cfg)
			return mw, mw.fws[pd].Init

		case "i":
			if !mw.fws[wd].GetSelection().IsDir() {
				break
			}
			mw.fws[pd] = mw.fws[wd]
			mw.fws[wd] = mw.fws[cd]
			mw.fws[cd] = fileview.New(mw.fws[wd].GetSelectedPath(), mw.w/3, mw.h, mw.cfg)
			return mw, mw.fws[cd].Init
		}
	}

	mw.iv, _ = mw.iv.Update(msg)
	if !mw.iv.Focused() {
		for i := range mw.fws {
			mw.fws[i], _ = mw.fws[i].Update(msg)
		}
	} else {
		mw.fwHeight-- // inputview is now focused
		mw.updateFWSizes()
	}

	return mw, nil
}

func (mw mainView) View() string {
	res := make([]string, len(mw.fws))
	for _, fw := range mw.fws {
		res = append(res, fw.View())
	}
	if mw.iv.Focused() {
		return lipgloss.JoinVertical(0, lipgloss.JoinHorizontal(lipgloss.Left, res...), mw.iv.View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, res...)
}

func (mw *mainView) updateFWSizes() {
	mw.fws[pd].SetSize(mw.fwWidth, mw.fwHeight)
	mw.fws[wd].SetSize(mw.fwWidth, mw.fwHeight)
	mw.fws[cd].SetSize(mw.w-mw.fwWidth*2, mw.fwHeight)
}

func (mw mainView) logState() {
	util.Log.Debug().
		Str("pd", mw.fws[pd].GetPath()).
		Str("wd", mw.fws[wd].GetPath()).
		Str("cd", mw.fws[cd].GetPath()).
		Send()
}
