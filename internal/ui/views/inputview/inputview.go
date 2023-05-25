package inputview

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type View struct {
	input     *strings.Builder
	isFocused bool
}

func New() View {
	return View{input: &strings.Builder{}}
}

func (v View) Init() tea.Cmd {
	return nil
}

func (v View) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC, tea.KeyEnter:
			if !v.isFocused {
				return v, nil
			}
			v.isFocused = false
			v.input.Reset()
			return v, nil
		}

		switch kp := msg.String(); kp {
		case ":":
			if !v.isFocused {
				v.isFocused = true
				return v, nil
			} else {
				v.input.WriteString(":")
			}
			return v, nil

		default:
			if v.isFocused {
				v.input.WriteString(kp)
			}
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
