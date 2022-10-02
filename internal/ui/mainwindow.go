package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type mainWindow struct {
	fw   tea.Model
	h, w int
}

func NewMainWindow() mainWindow {
	return mainWindow{fw: NewFileWindow("/home/panda")}
}

func (mw mainWindow) Init() tea.Cmd {
	return mw.fw.Init()
}

func (mw mainWindow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mw.h, mw.w = msg.Height, msg.Width

	case tea.KeyMsg:
		switch kp := msg.String(); kp {
		case "ctrl+c", "q":
			return mw, tea.Quit
		}
	}

	mw.fw, _ = mw.fw.Update(msg)
	return mw, nil
}

func (mw mainWindow) View() string {
	return mw.fw.View()
}
