package input

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	input     *strings.Builder
	isFocused bool
}

func New() Model {
	return Model{input: &strings.Builder{}}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC, tea.KeyEnter:
			if !m.isFocused {
				return m, nil
			}
			m.isFocused = false
			m.input.Reset()
			return m, nil
		}

		switch kp := msg.String(); kp {
		case ":":
			if !m.isFocused {
				m.isFocused = true
				return m, nil
			} else {
				m.input.WriteString(":")
			}
			return m, nil

		default:
			if m.isFocused {
				m.input.WriteString(kp)
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	return ":" + m.input.String()
}

func (m Model) Focused() bool {
	return m.isFocused
}
