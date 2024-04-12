package primary

import (
	"os"
	"os/exec"
	"path"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/msgs"
	"github.com/alx99/fly/internal/state"
	"github.com/alx99/fly/internal/views/dir"
	"github.com/alx99/fly/internal/views/input"
	"github.com/alx99/fly/internal/views/preview"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

type Config struct {
	PWDFile string
}

type model struct {
	pd      dir.Model // parent directory
	wd      dir.Model // working directory
	cd      dir.Model // child directory
	preview preview.Model
	im      input.Model
	state   *state.State

	// State
	h, w              int
	fwWidth, fwHeight int
	dirCache          map[string]dir.Model

	cfg      config.Config
	modelCfg Config
}

func New(state *state.State, modelCfg Config, cfg config.Config) (model, error) {
	d, err := os.Getwd()
	if err != nil {
		return model{}, nil
	}
	m := model{
		state:    state,
		cfg:      cfg,
		modelCfg: modelCfg,
		im:       input.New(),
		pd:       dir.New(path.Dir(d), dir.Parent, state, 0, 0, cfg),
		wd:       dir.New(d, dir.Working, state, 0, 0, cfg),
		cd:       dir.New("", dir.Child, state, 0, 0, cfg),
		dirCache: make(map[string]dir.Model),
	}

	return m, nil
}

// TODO, init on a parent directory that selects a directory out of view
// does not properly adjust the offset
func (m model) Init() tea.Cmd {
	// cd can't be initialized here since wd needs to be initialized first
	return tea.Batch(m.wd.Init(), m.pd.InitAndSelect(path.Base(m.wd.Path())))
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
			if m.modelCfg.PWDFile == "" {
				return m, tea.Quit
			}
			f, err := os.Create(m.modelCfg.PWDFile)
			if err != nil {
				return m, tea.Quit
			}
			f.WriteString(m.wd.Path())
			f.Close()
			return m, tea.Quit

		case m.cfg.Settings.Keybinds.NavUp:
			return m, m.move(dir.Up)

		case m.cfg.Settings.Keybinds.NavDown:
			return m, m.move(dir.Down)

		case m.cfg.Settings.Keybinds.NavLeft:
			if !m.pd.Focusable() {
				return m, nil
			}

			m.cacheAdd(m.cd)
			m.cd = m.wd.ChangeRole(dir.Child)
			m.wd = m.pd.ChangeRole(dir.Working)
			if m.pd.Path() == "/" {
				m.pd = dir.New("", dir.Parent, m.state, m.fwWidth, m.fwHeight, m.cfg)
			} else {
				m.pd, _ = m.cacheTryGet(path.Dir(m.pd.Path()), dir.Parent)
			}

			return m, m.pd.InitAndSelect(path.Base(m.wd.Path()))

		case m.cfg.Settings.Keybinds.NavRight:
			if !m.cd.Focusable() || m.cd.Empty() {
				return m, nil
			}
			if !m.wd.Selection().IsDir() {
				cmd := exec.Command("xdg-open", m.wd.SelectedPath())
				return m, func() tea.Msg {
					if err := cmd.Run(); err != nil {
						log.Err(err).Send()
					}
					return nil
				}
			}
			m.cacheAdd(m.pd)
			m.pd = m.wd.ChangeRole(dir.Parent)
			m.wd = m.cd.ChangeRole(dir.Working)
			if m.wd.Selection().IsDir() {
				var ok bool
				m.cd, ok = m.cacheTryGet(m.wd.SelectedPath(), dir.Child)
				if ok {
					return m, m.cd.Reinit()
				}
				return m, m.cd.Init()
			} else {
				// Note this is very much needed since otherwise
				// m.wd and m.cd will have the same ID and will
				// consume the same messages
				m.cd = dir.Model{}
				m.preview = preview.New(m.wd.SelectedPath(), m.fwWidth, m.fwHeight, m.cfg)
				return m, m.preview.Init()
			}

		case " ":
			m.state.ToggleSelect(m.wd.SelectedPath())
			return m, m.move(dir.Down)

		case m.cfg.Settings.Keybinds.Delete, m.cfg.Settings.Keybinds.Move:
			if !m.state.HasSelectedFiles() {
				m.state.ToggleSelect(m.wd.SelectedPath())
			}

			return m, func() tea.Msg {
				var err error
				if msg.String() == m.cfg.Settings.Keybinds.Delete {
					err = m.state.DeleteSelectedFiles()
				} else {
					err = m.state.MoveSelectedFiles(m.wd.Path())
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

	var cmds []tea.Cmd
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

func (m *model) move(direct dir.Direction) tea.Cmd {
	prev := m.wd.Selection()
	if m.wd.Move(direct).Selection() == prev {
		return nil // selection did not move, don't try to init a new dir
	}
	if m.wd.Selection().IsDir() {
		var ok bool
		m.cd, ok = m.cacheAdd(m.cd).cacheTryGet(m.wd.SelectedPath(), dir.Child)
		if ok {
			return m.cd.Reinit()
		}
		return m.cd.Init()
	} else if !m.wd.Selection().IsDir() {
		m.preview = preview.New(m.wd.SelectedPath(), m.fwWidth, m.fwHeight, m.cfg)
		return m.preview.Init()
	}
	return nil
}

func (m model) View() string {
	res := make([]string, 3)
	res = append(res, m.pd.View(), m.wd.View())
	if m.wd.Selection().IsDir() {
		res = append(res, m.cd.View())
	} else if !m.wd.Selection().IsDir() {
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

// cacheAdd adds a directory to the cache.
func (m *model) cacheAdd(dir dir.Model) *model {
	m.dirCache[dir.Path()] = dir
	return m
}

// cacheTryGet tries to get a directory from the cache.
// If the directory is not in the cache, it reads the directory and returns it.
func (m *model) cacheTryGet(path string, role dir.Role) (dir.Model, bool) {
	if dir, ok := m.dirCache[path]; ok {
		log.Trace().Str("path", dir.Path()).Msg("cache hit")
		return dir, true
	}
	return dir.New(path, role, m.state, m.fwWidth, m.fwHeight, m.cfg), false
}

func (m model) logState() {
	log.Debug().
		Str("pd", m.pd.Path()).
		Str("wd", m.wd.Path()).
		Str("cd", m.cd.Path()).
		Send()
}
