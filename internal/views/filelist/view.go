package filelist

import (
	"path/filepath"
	"strings"

	"github.com/alx99/sail/internal/filesys"
	"github.com/alx99/sail/internal/util"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

const cursor = "▶"

type selChecker interface {
	IsSelected(path string) bool
}

// View represents the state and view logic for a list of files.
type View struct {
	path    string
	entries []filesys.DirEntry

	selChecker     selChecker
	sb             strings.Builder
	highlightStyle lipgloss.Style
	applyHighlight bool
	maxHeight      int
	maxWidth       int
	cursorIndex    int
	viewportStart  int
	viewPortBuffer int
}

type State struct {
	ViewportStart int
	SelectedName  string
}

// New creates a new FileList.
func New(cwd string,
	state State,
	selChecker selChecker,
	applyHighlight bool,
	primaryColor lipgloss.Color,
) View {
	f := View{
		path:           cwd,
		cursorIndex:    0,
		viewPortBuffer: 2,
		selChecker:     selChecker,
		applyHighlight: applyHighlight,
		highlightStyle: lipgloss.NewStyle().Foreground(primaryColor),
	}

	f.SelectFileByName(state.SelectedName)
	if state.ViewportStart != 0 {
		f.viewportStart = state.ViewportStart
	}

	return f
}

// View renders the FileList.
func (v View) View() string {
	v.sb.Reset()
	viewportEnd := min(v.viewportStart+v.maxHeight, len(v.entries))

	for i := v.viewportStart; i < viewportEnd; i++ {
		file := v.entries[i]
		s := util.GetStyle(file)

		currentFile := i == v.cursorIndex
		selected := v.selChecker.IsSelected(filepath.Join(v.path, file.Name()))

		if selected {
			s = s.Underline(true).
				Bold(true).
				Foreground(v.highlightStyle.GetForeground())
		}

		usedCols := lipgloss.Width(file.Name())

		name := file.Name()
		if usedCols > v.maxWidth {
			name = file.Name()[:max(0, v.maxWidth-1)] + v.highlightStyle.Render("~")
		}

		if currentFile && v.applyHighlight {
			v.sb.WriteString(s.Render(name))
			v.sb.WriteString(strings.Repeat(v.highlightStyle.Render(lipgloss.RoundedBorder().Top), max(0, v.maxWidth-usedCols)))
		} else {
			v.sb.WriteString(s.Render(name))
		}

		if i != viewportEnd-1 {
			v.sb.WriteString("\n")
		}
	}

	if viewportEnd == 0 {
		msg := v.highlightStyle.Bold(true).Render("No files found")

		if v.applyHighlight {
			usedCols := lipgloss.Width(msg)
			v.sb.WriteString(msg)
			v.sb.WriteString(strings.Repeat(v.highlightStyle.Render(lipgloss.RoundedBorder().Top), max(0, v.maxWidth-usedCols)))
		} else {
			v.sb.WriteString(v.highlightStyle.Width(v.maxWidth).
				Height(v.maxHeight).
				Align(lipgloss.Center, lipgloss.Center).
				Render(msg))
		}
	}

	return v.sb.String()
}

func (v *View) ChDir(dir filesys.Dir, state State) {
	v.path = dir.Path()
	v.entries = dir.Entries()

	v.SelectFileByName(state.SelectedName)
	if state.ViewportStart != 0 {
		v.viewportStart = state.ViewportStart
	}

	log.Debug().
		Int("entriesCount", len(v.entries)).
		Int("cursorIndex", v.cursorIndex).
		Int("viewportStart", v.viewportStart).
		Str("path", v.path).
		Msg("ChDir")
}

// MoveUp moves the cursor up in the list.
func (v *View) MoveUp() {
	if len(v.entries) == 0 {
		return
	}
	if v.cursorIndex == 0 {
		v.cursorIndex = len(v.entries) - 1
		v.viewportStart = max(0, len(v.entries)-v.maxHeight)
	} else {
		v.cursorIndex--
		if v.cursorIndex < v.viewportStart+v.viewPortBuffer {
			v.viewportStart = max(0, v.cursorIndex-v.viewPortBuffer)
		}
	}
}

// MoveDown moves the cursor down in the list.
func (v *View) MoveDown() {
	if len(v.entries) == 0 {
		return
	}
	if v.cursorIndex == len(v.entries)-1 {
		v.cursorIndex = 0
		v.viewportStart = 0
	} else {
		v.cursorIndex++
		viewportEnd := v.viewportStart + v.maxHeight
		if v.cursorIndex > viewportEnd-v.viewPortBuffer-1 {
			// Ensure viewportStart calculation is correct and non-negative
			newStart := v.cursorIndex - v.maxHeight + v.viewPortBuffer + 1
			v.viewportStart = min(max(0, len(v.entries)-v.maxHeight), newStart)
		}
	}
}

// SetMaxDims sets the maximum dimensions for the view.
func (v *View) SetMaxDims(rows, cols int) {
	if rows <= 0 || cols <= 0 {
		return
	}
	log.Debug().Msgf("Setting max dims to %d x %d", rows, cols)

	v.maxHeight = rows
	v.maxWidth = cols
	// Ensure cursorIndex stays within bounds after resize
	if len(v.entries) > 0 {
		v.cursorIndex = min(v.cursorIndex, len(v.entries)-1)
	} else {
		v.cursorIndex = 0
	}

	v.setIdealViewPort()
}

// SelectFileByName selects a file by its name.
func (v *View) SelectFileByName(name string) {
	for i, file := range v.entries {
		if file.Name() == name {
			v.cursorIndex = i
			v.setIdealViewPort()
			return
		}
	}

	// No name found, or possible empty
	v.cursorIndex = 0
	v.setIdealViewPort()
}

// setIdealViewPort adjusts the viewport based on the cursor position.
func (v *View) setIdealViewPort() {
	if v.maxHeight <= 0 || len(v.entries) == 0 { // Prevent division by zero or issues with empty list
		v.viewportStart = 0
		return
	}
	// Calculate ideal centered position
	idealStart := v.cursorIndex - v.maxHeight/2

	// Ensure we don't go before start of list
	v.viewportStart = max(0, idealStart)

	// Adjust if we'd show past end of list
	if v.viewportStart+v.maxHeight > len(v.entries) {
		v.viewportStart = max(0, len(v.entries)-v.maxHeight)
	}

	// Still maintain minimum buffer if needed (ensure viewportStart doesn't become negative)
	if v.cursorIndex-v.viewportStart < v.viewPortBuffer {
		v.viewportStart = max(0, v.cursorIndex-v.viewPortBuffer)
	}

	// Final check to ensure viewportStart is not negative
	v.viewportStart = max(0, v.viewportStart)
}

func (v *View) CurrEntry() (filesys.DirEntry, bool) {
	if len(v.entries) == 0 || v.cursorIndex < 0 || v.cursorIndex >= len(v.entries) {
		return filesys.DirEntry{}, false
	}
	return v.entries[v.cursorIndex], true
}

// Path returns the current working directory.
func (v *View) Path() string {
	return v.path
}

// State returns the current state of the FileList for caching.
func (v *View) State() State {
	name := ""
	if e, ok := v.CurrEntry(); ok {
		name = e.Name()
	}
	return State{
		ViewportStart: v.viewportStart,
		SelectedName:  name,
	}
}

func (v *View) SelectedRow() int {
	return max(0, v.cursorIndex-v.viewportStart)
}
