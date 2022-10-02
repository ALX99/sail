package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mainWindow struct {
	fws  []fileWindow
	h, w int
}

func NewMainWindow() mainWindow {
	fws := make([]fileWindow, 3)
	fws[0] = NewFileWindow("/", 0)
	fws[1] = NewFileWindow("/var", 1)
	fws[2] = NewFileWindow("/var/lib", 2)

	return mainWindow{fws: fws}
}

func (mw mainWindow) Init() tea.Cmd {
	return tea.Batch(mw.fws[0].Init, mw.fws[1].Init, mw.fws[2].Init)
}

func (mw mainWindow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mw.h, mw.w = msg.Height, msg.Width

	case tea.KeyMsg:
		switch kp := msg.String(); kp {
		case "ctrl+c", "q":
			return mw, tea.Quit

		case "e":
			mw.fws[1].move(Up)

		case "n":
			mw.fws[1].move(Down)
		}
	}

	for i := range mw.fws {
		mw.fws[i], _ = mw.fws[i].Update(msg)
	}

	return mw, nil
}

func (mw mainWindow) View() string {
	res := make([]string, len(mw.fws))
	for _, fw := range mw.fws {
		res = append(res, fw.View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, res...)
}
