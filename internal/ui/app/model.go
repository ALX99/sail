package app

import (
	"log/slog"
	"os"

	"github.com/alx99/sail/internal/config"
	"github.com/alx99/sail/internal/filesys"
	"github.com/alx99/sail/internal/ui/browser"
	"github.com/alx99/sail/internal/ui/components/status"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	cfg       config.Config
	selection *filesys.Selection
	browser   *browser.Model
	status    *status.View
	altScreen bool
	printLast string
}

func New(cwd string, cfg config.Config) *Model {
	selection := filesys.NewSelection()
	return &Model{
		cfg:       cfg,
		selection: selection,
		browser:   browser.New(cwd, cfg, selection),
		status:    status.New(),
		altScreen: cfg.Settings.AltScreen,
		printLast: cfg.PrintLastWD,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.browser.Init(), m.status.Init())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.status.Update(msg)

	forwardToBrowser := func(message tea.Msg) {
		if message == nil {
			return
		}
		var cmd tea.Cmd
		m.browser, cmd = m.browser.Update(message)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.printLast != "" {
				if err := m.writeLastWD(); err != nil {
					slog.Error("Failed to write last working directory", "error", err)
				}
			}
			return m, tea.Quit
		case m.cfg.Settings.Keymap.ToggleAltScreen:
			m.altScreen = !m.altScreen
			if m.altScreen {
				return m, tea.EnterAltScreen
			}
			return m, tea.ExitAltScreen
		default:
			forwardToBrowser(msg)
		}
	default:
		forwardToBrowser(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.status.SetWidth(msg.Width)
	case filesys.DirLoadedMsg:
		cmds = append(cmds, m.status.SetWD(msg.Dir))
	case error:
		slog.Error("Error occurred", "error", msg)
		cmds = append(cmds, m.status.SetError(msg))
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.browser.View(),
		m.status.View(),
	)
}

func (m *Model) writeLastWD() error {
	f, err := os.Create(m.printLast)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(m.browser.CWD())
	return err
}
