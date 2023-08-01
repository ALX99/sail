package primary

import (
	"os"
	"os/exec"
	"path"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/models/directory"
	"github.com/alx99/fly/internal/models/input"
	"github.com/alx99/fly/internal/models/preview"
	"github.com/alx99/fly/internal/msgs"
	"github.com/alx99/fly/internal/state"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

type model struct {
	pd      directory.Model // parent directory
	wd      directory.Model // working directory
	cd      directory.Model // child directory
	preview preview.Model
	im      input.Model
	state   *state.State

	// State
	h, w              int
	fwWidth, fwHeight int
	firstLoad         bool

	cfg config.Config
}

func New(state *state.State, cfg config.Config) (model, error) {
	dir, err := os.Getwd()
	if err != nil {
		return model{}, nil
	}
	m := model{
		state:     state,
		cfg:       cfg,
		im:        input.New(),
		wd:        directory.New(dir, state, 0, 0, cfg),
		firstLoad: true,
	}

	return m, nil
}

func (m model) Init() tea.Cmd {
	return m.wd.Init() // wd needs to initialized before pd and cd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case msgs.MsgDirLoaded, msgs.MsgDirError:
		// First-load addressed to wd
		if msg, ok := msg.(msgs.MsgDirLoaded); ok && m.firstLoad && msg.Path == m.wd.GetPath() {
			m.firstLoad = false
			var cmd tea.Cmd

			m.wd, cmd = m.wd.Update(msg)

			// Create pd
			m.pd = directory.New(path.Dir(msg.Path), m.state, m.fwWidth, m.fwHeight, m.cfg)

			// Create cd
			if !m.wd.Empty() && m.wd.GetSelection().IsDir() {
				m.cd = directory.New(m.wd.GetSelectedPath(), m.state, m.fwWidth, m.fwHeight, m.cfg)
			} else {
				m.cd = directory.New("", m.state, m.fwWidth, m.fwHeight, m.cfg)
			}

			// Initialize cd and pd
			return m, tea.Batch(m.pd.InitAndSelect(path.Base(m.wd.GetPath())), m.cd.Init(), cmd)
		}

		cmds := []tea.Cmd{}
		var cmd tea.Cmd
		m.pd, cmd = m.pd.Update(msg)
		cmds = append(cmds, cmd)
		m.wd, cmd = m.wd.Update(msg)
		cmds = append(cmds, cmd)
		m.cd, cmd = m.cd.Update(msg)
		cmds = append(cmds, cmd)

		// No need to propagate further
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.h, m.w = msg.Height, msg.Width

		m.fwWidth = msg.Width / 3 // width
		m.fwHeight = msg.Height
		if m.im.Focused() {
			m.updateFWHeight(-1)
		} else {
			m.updateFWHeight(0)
		}

		log.Debug().
			Int("height", msg.Height).
			Int("width", msg.Width).
			Msg("terminal size updated")

	case tea.KeyMsg:
		wasFocused := m.im.Focused()
		var cmd tea.Cmd
		m.im, cmd = m.im.Update(msg)
		// focus changed
		if wasFocused != m.im.Focused() {
			if m.im.Focused() {
				m.updateFWHeight(-1)
			} else {
				m.updateFWHeight(1)
			}
			return m, cmd // focus changed, no need to propagate msg
		} else if m.im.Focused() {
			return m, cmd // already focused, no need to propagate msg
		}

		switch kp := msg.String(); kp {
		case "ctrl+c", "q":
			return m, tea.Quit

		case m.cfg.Settings.Keybinds.NavUp:
			return m, m.moveUp()

		case m.cfg.Settings.Keybinds.NavDown:
			return m, m.moveDown()

		case m.cfg.Settings.Keybinds.NavLeft:
			if !m.pd.IsFocusable() {
				return m, nil
			}

			m.cd = m.wd
			m.wd = m.pd
			if m.pd.GetPath() == "/" {
				m.pd = directory.New("", m.state, m.fwWidth, m.fwHeight, m.cfg)
			} else {
				m.pd = directory.New(path.Dir(m.pd.GetPath()), m.state, m.fwWidth, m.fwHeight, m.cfg)
			}

			return m, m.pd.InitAndSelect(path.Base(m.wd.GetPath()))

		case m.cfg.Settings.Keybinds.NavRight:
			if !m.cd.IsFocusable() || m.cd.Empty() {
				return m, nil
			}
			if !m.wd.GetSelection().IsDir() {
				cmd := exec.Command("xdg-open", m.wd.GetSelectedPath())
				return m, func() tea.Msg {
					if err := cmd.Run(); err != nil {
						log.Err(err).Send()
					}
					return nil
				}
			}
			m.pd = m.wd
			m.wd = m.cd
			if m.wd.GetSelection().IsDir() {
				m.cd = directory.New(m.wd.GetSelectedPath(), m.state, m.fwWidth, m.fwHeight, m.cfg)
				return m, m.cd.Init()
			} else {
				// Note this is very much needed since otherwise
				// m.wd and m.cd will have the same ID and will
				// consume the same messages
				m.cd = directory.Model{}
				m.preview = preview.New(m.wd.GetSelectedPath(), m.fwWidth, m.fwHeight, m.cfg)
				return m, m.preview.Init()
			}

		case " ":
			m.state.ToggleSelect(m.wd.GetSelectedPath())
			return m, m.moveDown()

		case m.cfg.Settings.Keybinds.Delete, m.cfg.Settings.Keybinds.Move:
			if !m.state.HasSelectedFiles() {
				m.state.ToggleSelect(m.wd.GetSelectedPath())
			}

			return m, func() tea.Msg {
				var err error
				if msg.String() == m.cfg.Settings.Keybinds.Delete {
					err = m.state.DeleteSelectedFiles()
				} else {
					err = m.state.MoveSelectedFiles(m.wd.GetPath())
				}
				if err != nil {
					log.Err(err).Send()
				}
				return msgs.MsgDirReload{}
			}

		case ".":
			// TODO hidden files
			return m, nil
		}
	}

	cmds := []tea.Cmd{}
	var cmd tea.Cmd
	m.pd, cmd = m.pd.Update(msg)
	cmds = append(cmds, cmd)
	m.wd, cmd = m.wd.Update(msg)
	cmds = append(cmds, cmd)
	m.cd, cmd = m.cd.Update(msg)
	cmds = append(cmds, cmd)
	m.preview, cmd = m.preview.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// moveDown navigates downwards in the working directory
func (m *model) moveDown() tea.Cmd {
	if m.wd.Move(directory.Down).GetSelection().IsDir() {
		m.cd = directory.New(m.wd.GetSelectedPath(), m.state, m.fwWidth, m.fwHeight, m.cfg)
		return m.cd.Init()
	} else if !m.wd.GetSelection().IsDir() {
		m.preview = preview.New(m.wd.GetSelectedPath(), m.fwWidth, m.fwHeight, m.cfg)
		return m.preview.Init()
	}
	return nil
}

// moveUp navigates upwards in the working directory
func (m *model) moveUp() tea.Cmd {
	if m.wd.Move(directory.Up).GetSelection().IsDir() {
		m.cd = directory.New(m.wd.GetSelectedPath(), m.state, m.fwWidth, m.fwHeight, m.cfg)
		return m.cd.Init()
	} else if !m.wd.GetSelection().IsDir() {
		m.preview = preview.New(m.wd.GetSelectedPath(), m.fwWidth, m.fwHeight, m.cfg)
		return m.preview.Init()
	}
	return nil
}

func (m model) View() string {
	if m.firstLoad {
		return "loading..."
	}
	res := make([]string, 3)
	res = append(res, m.pd.View(), m.wd.View())
	if m.wd.GetSelection().IsDir() {
		res = append(res, m.cd.View())
	} else if !m.wd.GetSelection().IsDir() {
		res = append(res, m.preview.View())
	}
	if m.im.Focused() {
		return lipgloss.JoinVertical(lipgloss.Top,
			lipgloss.NewStyle().Height(m.h-1).Render(
				lipgloss.JoinHorizontal(lipgloss.Left, res...)),
			m.im.View())
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, res...)
}

func (m *model) updateFWHeight(delta int) {
	m.fwHeight += delta

	m.pd.SetSize(m.fwWidth, m.fwHeight)
	m.wd.SetSize(m.fwWidth, m.fwHeight)
	m.cd.SetSize(m.w-m.fwWidth*2, m.fwHeight)
	m.preview.SetSize(m.w-m.fwWidth*2, m.fwHeight)
}

func (m model) logState() {
	log.Debug().
		Str("pd", m.pd.GetPath()).
		Str("wd", m.wd.GetPath()).
		Str("cd", m.cd.GetPath()).
		Send()
}
