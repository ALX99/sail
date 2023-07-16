package preview

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alx99/fly/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

type Model struct {
	path string

	h, w           int
	stdout, stderr string
	err            error
}

type cmdCompletionMsg struct {
	stdout, stderr string
	err            error
}

func New(path string, width, height int, cfg config.Config) Model {
	return Model{
		path: path,
		h:    height,
		w:    width,
	}
}

func (m Model) Init() tea.Cmd {
	/* TODO support later
	return func() tea.Msg {
		cmd := exec.Command("bat", "--color", "always", "--style", "plain", m.path)

		stdout, err := cmd.Output()
		if err != nil {
			exitErr, ok := err.(*exec.ExitError)
			if ok {
				return cmdCompletionMsg{
					err:    err,
					stderr: string(exitErr.Stderr),
				}
			} else {
				return cmdCompletionMsg{
					err: err,
				}
			}
		}

		return cmdCompletionMsg{
			stdout: string(stdout),
		}*/

	return func() tea.Msg {
		file, err := os.Open(m.path)
		if err != nil {
			return cmdCompletionMsg{
				err: err,
			}
		}
		defer file.Close()

		buf := make([]byte, 1024)
		t := time.Now()
		n, err := file.Read(buf)
		if err != nil {
			return cmdCompletionMsg{
				err: err,
			}
		}
		log.Trace().
			Str("path", m.path).
			Str("dur", time.Since(t).String()).
			Msg("File read")

		contentType := http.DetectContentType(buf)
		if strings.HasPrefix(contentType, "text/plain;") {
			return cmdCompletionMsg{
				stdout: string(buf[:n]),
			}
		} else {
			return cmdCompletionMsg{
				err: fmt.Errorf("unsupported content type %q", contentType),
			}
		}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cmdCompletionMsg:
		m.stdout = msg.stdout
		m.stderr = msg.stderr
		m.err = msg.err
	}
	return m, nil
}

func (m Model) View() string {
	t := time.Now()
	defer func() {
		log.Trace().
			Str("dur", time.Since(t).String()).
			Str("file", m.path).
			Msg("View render")
	}()
	style := lipgloss.NewStyle().Width(m.w).MaxHeight(m.h)
	if m.err != nil {
		return style.Render(m.err.Error())
	}
	if m.stderr != "" {
		return style.Render(m.stderr)
	}
	if m.stdout != "" {
		return style.Render(m.stdout)
	}
	return style.Render("loading")
}

// SetSize sets the max allowed size of the window
func (m *Model) SetSize(w, h int) *Model {
	m.w = w
	m.h = h
	return m
}
