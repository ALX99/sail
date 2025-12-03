package app

import (
	"log/slog"
	"os"

	"github.com/alx99/sail/internal/config"
	"github.com/alx99/sail/internal/filesys"
	"github.com/alx99/sail/internal/style"
	"github.com/alx99/sail/internal/ui/browser"
	"github.com/alx99/sail/internal/ui/components/status"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	cfg       config.Config
	browser   *browser.Model
	status    *status.View
	altScreen bool
	printLast string
}

func New(cwd string, cfg config.Config, styles *style.Styles) *Model {
	return &Model{
		cfg:       cfg,
		browser:   browser.New(cwd, cfg, styles),
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
	var cmd tea.Cmd

	// Handle global keys
	if msg, ok := msg.(tea.KeyMsg); ok {
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
		}
	}

	// Update Status internal state
	if cmd := m.status.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Handle specific messages for status bar coordination
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.status.SetWidth(msg.Width)

		// adjusted height accounting for status bar
		adjHeight := max(msg.Height-m.status.Height(), 0)

		m.browser, cmd = m.browser.Update(tea.WindowSizeMsg{Width: msg.Width, Height: adjHeight})
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)
	case filesys.DirLoadedMsg:
		cmds = append(cmds, m.status.SetWD(msg.Dir))
	case error:
		slog.Error("Error occurred", "error", msg)
		cmds = append(cmds, m.status.SetError(msg))
	}

	// Forward to browser
	m.browser, cmd = m.browser.Update(msg)
	cmds = append(cmds, cmd)
	idx, total, selected, name := m.browser.SelectionStats()
	m.status.SetSelection(idx, total, selected, name)

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
