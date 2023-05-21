package inputview

import (
	"strings"

	"github.com/alx99/fly/internal/command"
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

func (v View) Update(msg tea.Msg) (View, command.Command) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC, tea.KeyEnter:
			if !v.isFocused {
				return v, command.NUL // Do nothing if not focused
			}
			v.isFocused = false
			v.input.Reset()
			return v, command.RecalculateViews
		}

		switch kp := msg.String(); kp {
		case ":":
			if !v.isFocused {
				v.isFocused = true
				return v, command.RecalculateViews
			} else {
				v.input.WriteString(":")
			}
			return v, command.NUL

		default:
			if v.isFocused {
				v.input.WriteString(kp)
			}
			return v, command.NUL
		}
	}

	return v, command.NUL
}

func (v View) View() string {
	return ":" + v.input.String()
}

func (v View) Focused() bool {
	return v.isFocused
}
