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

const (
	pd = iota // parent directory
	wd        // working directory
	cd        // child directory
)

type view struct {
	fws []fileview.View
	iv  inputview.View

	h, w     int
	fwWidth  int
	fwHeight int

	cfg config.Config
}

func New(cfg config.Config) (view, error) {
	fws := make([]fileview.View, 3)
	home, ok := os.LookupEnv("HOME")
	if !ok {
		return view{}, errors.New("$HOME not set")
	}

	fws[pd] = fileview.New(util.GetParentPath(home), 0, 0, cfg)
	fws[wd] = fileview.New(home, 0, 0, cfg)
	if err := fws[wd].Load(); err != nil {
		return view{}, err
	}

	if fws[wd].GetSelection().IsDir() {
		fws[cd] = fileview.New(fws[wd].GetSelectedPath(), 0, 0, cfg)
	} else {
		fws[cd] = fileview.New("", 0, 0, cfg)
	}

	return view{fws: fws, cfg: cfg, iv: inputview.New()}, nil
}

func (v view) Init() tea.Cmd {
	return tea.Batch(v.fws[cd].Init(), v.fws[pd].Init())
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
			if v.fws[wd].Move(fileview.Up).GetSelection().IsDir() {
				v.fws[cd] = fileview.New(v.fws[wd].GetSelectedPath(), v.fwWidth, v.fwHeight, v.cfg)
				return v, v.fws[cd].Init()
			}
			return v, nil

		case "n":
			if v.fws[wd].Move(fileview.Down).GetSelection().IsDir() {
				v.fws[cd] = fileview.New(v.fws[wd].GetSelectedPath(), v.fwWidth, v.fwHeight, v.cfg)
				return v, v.fws[cd].Init()
			}
			return v, nil

		case "m":
			if !v.fws[pd].IsFocusable() {
				return v, nil
			}

			v.fws[cd] = v.fws[wd]
			v.fws[wd] = v.fws[pd]
			if v.fws[pd].GetPath() == "/" {
				v.fws[pd] = fileview.New("", v.w/3, v.h, v.cfg)
			} else {
				v.fws[pd] = fileview.New(util.GetParentPath(v.fws[pd].GetPath()), v.w/3, v.h, v.cfg)
			}
			return v, v.fws[pd].Init()

		case "i":
			if !v.fws[cd].IsFocusable() || !v.fws[wd].GetSelection().IsDir() {
				return v, nil
			}
			v.fws[pd] = v.fws[wd]
			v.fws[wd] = v.fws[cd]
			v.fws[cd] = fileview.New(v.fws[wd].GetSelectedPath(), v.w/3, v.h, v.cfg)
			return v, v.fws[cd].Init()
		}
	}

	cmds := []tea.Cmd{}
	for i := range v.fws {
		var cmd tea.Cmd
		v.fws[i], cmd = v.fws[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v view) View() string {
	res := make([]string, len(v.fws))
	for _, fw := range v.fws {
		res = append(res, fw.View())
	}
	if v.iv.Focused() {
		return lipgloss.JoinVertical(0, lipgloss.JoinHorizontal(lipgloss.Left, res...), v.iv.View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, res...)
}

func (v *view) updateFWHeight(delta int) {
	v.fwHeight += delta

	v.fws[pd].SetSize(v.fwWidth, v.fwHeight)
	v.fws[wd].SetSize(v.fwWidth, v.fwHeight)
	v.fws[cd].SetSize(v.w-v.fwWidth*2, v.fwHeight)
}

func (v view) logState() {
	log.Debug().
		Str("pd", v.fws[pd].GetPath()).
		Str("wd", v.fws[wd].GetPath()).
		Str("cd", v.fws[cd].GetPath()).
		Send()
}
