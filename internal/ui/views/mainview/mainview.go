package mainview

import (
	"errors"
	"os"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/ui/views/fileview"
	"github.com/alx99/fly/internal/ui/views/inputview"
	"github.com/alx99/fly/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

type view struct {
	pd fileview.View // parent directory
	wd fileview.View // working directory
	cd fileview.View // child directory
	iv inputview.View

	h, w     int
	fwWidth  int
	fwHeight int

	cfg config.Config
}

func New(cfg config.Config) (view, error) {
	home, ok := os.LookupEnv("HOME")
	if !ok {
		return view{}, errors.New("$HOME not set")
	}
	v := view{
		pd:  fileview.New(util.GetParentPath(home), 0, 0, cfg),
		cfg: cfg,
		iv:  inputview.New(),
	}

	v.wd = fileview.New(home, 0, 0, cfg)
	if err := v.wd.Load(); err != nil {
		return v, err
	}

	if !v.wd.Empty() && v.wd.GetSelection().IsDir() {
		v.cd = fileview.New(v.wd.GetSelectedPath(), 0, 0, cfg)
	} else {
		v.cd = fileview.New("", 0, 0, cfg)
	}

	return v, nil
}

func (v view) Init() tea.Cmd {
	return tea.Batch(v.cd.Init(), v.pd.Init())
}

func (v view) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.h, v.w = msg.Height, msg.Width

		v.fwWidth = msg.Width / 3 // width
		v.fwHeight = msg.Height
		if v.iv.Focused() {
			v.updateFWHeight(-1)
		} else {
			v.updateFWHeight(0)
		}

		log.Debug().
			Int("height", msg.Height).
			Int("Width", msg.Width).
			Msg("Terminal size updated")

	case tea.KeyMsg:
		newIV, cmd := v.iv.Update(msg)
		// focus changed
		if v.iv.Focused() != newIV.Focused() {
			if newIV.Focused() {
				v.updateFWHeight(-1)
			} else {
				v.updateFWHeight(1)
			}
		}

		v.iv = newIV
		if v.iv.Focused() {
			return v, cmd // focus obtained, no need to propagate msg
		}

		switch kp := msg.String(); kp {
		case "ctrl+c", "q":
			return v, tea.Quit

		case "e":
			if !v.wd.Empty() && v.wd.Move(fileview.Up).GetSelection().IsDir() {
				v.cd = fileview.New(v.wd.GetSelectedPath(), v.fwWidth, v.fwHeight, v.cfg)
				return v, v.cd.Init()
			}
			return v, nil

		case "n":
			if !v.wd.Empty() && v.wd.Move(fileview.Down).GetSelection().IsDir() {
				v.cd = fileview.New(v.wd.GetSelectedPath(), v.fwWidth, v.fwHeight, v.cfg)
				return v, v.cd.Init()
			}
			return v, nil

		case "m":
			if !v.pd.IsFocusable() {
				return v, nil
			}

			v.cd = v.wd
			v.wd = v.pd
			if v.pd.GetPath() == "/" {
				v.pd = fileview.New("", v.w/3, v.h, v.cfg)
			} else {
				v.pd = fileview.New(util.GetParentPath(v.pd.GetPath()), v.w/3, v.h, v.cfg)
			}
			return v, v.pd.Init()

		case "i":
			if !v.cd.IsFocusable() || v.wd.Empty() || !v.wd.GetSelection().IsDir() {
				return v, nil
			}
			v.pd = v.wd
			v.wd = v.cd
			v.cd = fileview.New(v.wd.GetSelectedPath(), v.w/3, v.h, v.cfg)
			return v, v.cd.Init()
		}
	}

	cmds := []tea.Cmd{}
	var cmd tea.Cmd
	v.pd, cmd = v.pd.Update(msg)
	cmds = append(cmds, cmd)
	v.wd, cmd = v.wd.Update(msg)
	cmds = append(cmds, cmd)
	v.cd, cmd = v.cd.Update(msg)
	cmds = append(cmds, cmd)

	return v, tea.Batch(cmds...)
}

func (v view) View() string {
	res := make([]string, 3)
	res = append(res, v.pd.View(), v.wd.View(), v.cd.View())
	if v.iv.Focused() {
		return lipgloss.JoinVertical(0, lipgloss.JoinHorizontal(lipgloss.Left, res...), v.iv.View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, res...)
}

func (v *view) updateFWHeight(delta int) {
	v.fwHeight += delta

	v.pd.SetSize(v.fwWidth, v.fwHeight)
	v.wd.SetSize(v.fwWidth, v.fwHeight)
	v.cd.SetSize(v.w-v.fwWidth*2, v.fwHeight)
}

func (v view) logState() {
	log.Debug().
		Str("pd", v.pd.GetPath()).
		Str("wd", v.wd.GetPath()).
		Str("cd", v.cd.GetPath()).
		Send()
}
