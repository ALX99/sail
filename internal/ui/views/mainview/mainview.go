package mainview

import (
	"github.com/alx99/fly/internal/ui/views/fileview"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mainView struct {
	fws  []fileview.Window
	h, w int
}

func New() mainView {
	fws := make([]fileview.Window, 3)
	fws[0] = fileview.New("/", 0)
	fws[1] = fileview.New("/var", 1)
	fws[2] = fileview.New("/var/lib", 2)

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
			mw.fws[1].Move(fileview.Up)

		case "n":
			mw.fws[1].Move(fileview.Down)
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
