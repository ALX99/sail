package primary

import (
	"errors"
	"os"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/fs"
	"github.com/alx99/fly/internal/models/directory"
	"github.com/alx99/fly/internal/models/input"
	"github.com/alx99/fly/internal/models/preview"
	"github.com/alx99/fly/internal/util"
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
	browser *fs.Browser

	h, w     int
	fwWidth  int
	fwHeight int

	cfg config.Config
}

func New(browser *fs.Browser, cfg config.Config) (model, error) {
	home, ok := os.LookupEnv("HOME")
	if !ok {
		return model{}, errors.New("$HOME not set")
	}
	m := model{
		browser: browser,
		pd:      directory.New(util.GetParentPath(home), browser, 0, 0, cfg),
		cfg:     cfg,
		im:      input.New(),
	}

	m.wd = directory.New(home, browser, 0, 0, cfg)
	if err := m.wd.Load(); err != nil {
		return m, err
	}

	if !m.wd.Empty() && m.wd.GetSelection().IsDir() {
		m.cd = directory.New(m.wd.GetSelectedPath(), browser, 0, 0, cfg)
	} else {
		m.cd = directory.New("", browser, 0, 0, cfg)
	}

	return m, nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.cd.Init(), m.pd.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		newIV, cmd := m.im.Update(msg)
		// focus changed
		if m.im.Focused() != newIV.Focused() {
			if newIV.Focused() {
				m.updateFWHeight(-1)
			} else {
				m.updateFWHeight(1)
			}
		}

		m.im = newIV
		if m.im.Focused() {
			return m, cmd // focus obtained, no need to propagate msg
		}

		switch kp := msg.String(); kp {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "e":
			if !m.wd.Empty() && m.wd.Move(directory.Up).GetSelection().IsDir() {
				m.cd = directory.New(m.wd.GetSelectedPath(), m.browser, m.fwWidth, m.fwHeight, m.cfg)
				return m, m.cd.Init()
			} else if !m.wd.GetSelection().IsDir() {
				m.preview = preview.New(m.wd.GetSelectedPath(), m.fwWidth, m.h, m.cfg)
				return m, m.preview.Init()
			}
			return m, nil

		case "n":
			if !m.wd.Empty() && m.wd.Move(directory.Down).GetSelection().IsDir() {
				m.cd = directory.New(m.wd.GetSelectedPath(), m.browser, m.fwWidth, m.fwHeight, m.cfg)
				return m, m.cd.Init()
			} else if !m.wd.GetSelection().IsDir() {
				m.preview = preview.New(m.wd.GetSelectedPath(), m.fwWidth, m.h, m.cfg)
				return m, m.preview.Init()
			}
			return m, nil

		case "m":
			if !m.pd.IsFocusable() {
				return m, nil
			}

			m.cd = m.wd
			m.wd = m.pd
			if m.pd.GetPath() == "/" {
				m.pd = directory.New("", m.browser, m.w/3, m.h, m.cfg)
			} else {
				m.pd = directory.New(util.GetParentPath(m.pd.GetPath()), m.browser, m.w/3, m.h, m.cfg)
			}
			return m, m.pd.Init()

		case "i":
			if !m.cd.IsFocusable() || m.wd.Empty() || !m.wd.GetSelection().IsDir() {
				return m, nil
			}
			m.pd = m.wd
			m.wd = m.cd
			m.cd = directory.New(m.wd.GetSelectedPath(), m.browser, m.w/3, m.h, m.cfg)
			return m, m.cd.Init()

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

func (m model) View() string {
	res := make([]string, 3)
	res = append(res, m.pd.View(), m.wd.View())
	if m.wd.GetSelection().IsDir() {
		res = append(res, m.cd.View())
	} else if !m.wd.GetSelection().IsDir() {
		res = append(res, m.preview.View())
	}
	if m.im.Focused() {
		return lipgloss.JoinVertical(0, lipgloss.JoinHorizontal(lipgloss.Left, res...), m.im.View())
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
