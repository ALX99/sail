package inputview

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type View struct {
	input     strings.Builder
	isFocused bool
}

func New() View {
	return View{}
}

func (v View) Init() tea.Cmd {
	return nil
}

func (v View) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			if v.isFocused {
				v.isFocused = false
				v.input.Reset()
			}
			return v, nil
		}

		switch kp := msg.String(); kp {
		case ":":
			if !v.isFocused {
				v.isFocused = true
			} else {
				v.input.WriteString(":")
			}
			return v, nil

		default:
			v.input.WriteString(kp)
			return v, nil
		}
	}

	return v, nil
}

func (v View) View() string {
	return ":" + v.input.String()
}

func (v View) Focused() bool {
	return v.isFocused
}
