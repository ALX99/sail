package filelist

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/alx99/sail/internal/filesys"
	"github.com/alx99/sail/internal/ui/theme"
	"github.com/alx99/sail/internal/util"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/collate"
)

type SelChecker interface {
	IsSelected(path string) bool
}

// View represents the state and view logic for a list of files.
type View struct {
	path       string
	entries    []filesys.DirEntry
	allEntries []filesys.DirEntry

	collator       *collate.Collator
	selChecker     SelChecker
	sb             strings.Builder
	highlightStyle lipgloss.Style
	applyHighlight bool
	showHidden     bool
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
	selChecker SelChecker,
	coll *collate.Collator,
	applyHighlight bool,
) *View {
	f := &View{
		path:           cwd,
		cursorIndex:    0,
		viewPortBuffer: 2,
		collator:       coll,
		selChecker:     selChecker,
		applyHighlight: applyHighlight,
		showHidden:     false,
	}

	f.SelectFileByName(state.SelectedName)
	if state.ViewportStart != 0 {
		f.viewportStart = state.ViewportStart
	}

	return f
}

func (v *View) SetShowHidden(show bool) {
	currEntry, ok := v.CurrEntry()
	targetName := ""
	if ok {
		targetName = currEntry.Name()
	}

	v.showHidden = show

	if !v.showHidden && isHidden(targetName) {
		targetName = v.findClosestVisible(targetName)
	}

	v.filterEntries()

	// Restore selection
	if targetName != "" {
		v.SelectFileByName(targetName)
	}
}

func (v *View) findClosestVisible(currentName string) string {
	idx := -1
	for i, e := range v.allEntries {
		if e.Name() == currentName {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ""
	}

	// Search forward
	for i := idx + 1; i < len(v.allEntries); i++ {
		if !v.allEntries[i].IsHidden() {
			return v.allEntries[i].Name()
		}
	}

	// Search backward
	for i := idx - 1; i >= 0; i-- {
		if !v.allEntries[i].IsHidden() {
			return v.allEntries[i].Name()
		}
	}

	return ""
}

func (v *View) filterEntries() {
	if v.showHidden {
		v.entries = v.allEntries
	} else {
		v.entries = make([]filesys.DirEntry, 0, len(v.allEntries))
		for _, e := range v.allEntries {
			if !e.IsHidden() {
				v.entries = append(v.entries, e)
			}
		}
	}

	if len(v.entries) > 0 {
		v.cursorIndex = min(v.cursorIndex, len(v.entries)-1)
	} else {
		v.cursorIndex = 0
	}
	v.setIdealViewPort()
}

// View renders the FileList.
func (v *View) View() string {
	if v.path == "" {
		return ""
	}

	v.sb.Reset()
	viewportEnd := min(v.viewportStart+v.maxHeight, len(v.entries))

	for i := v.viewportStart; i < viewportEnd; i++ {
		file := v.entries[i]
		currentFile := i == v.cursorIndex
		selected := v.selChecker.IsSelected(filepath.Join(v.path, file.Name()))

		// Base style
		style := lipgloss.NewStyle()

		// Icon
		icon := util.GetIcon(file.Name(), file.IsDir())
		iconWidth := lipgloss.Width(icon)

		// Prepare Name with Truncation
		name := file.Name()
		// Available width for name: maxWidth - icon - space
		availableWidth := max(0, v.maxWidth-iconWidth-1)

		if lipgloss.Width(name) > availableWidth {
			// Truncate
			// Use runic truncation if possible, but simple string slicing for now
			// Safety check for short available width
			if availableWidth > 1 {
				runes := []rune(name)
				if len(runes) > availableWidth {
					name = string(runes[:availableWidth-1]) + "…"
				}
			} else {
				name = "" // Too small to show name
			}
		}

		displayName := fmt.Sprintf("%s %s", icon, name)

		// Selection (yellow text)
		if selected {
			style = theme.DefaultTheme.SelectedFile
		} else {
			if file.IsDir() {
				style = style.Foreground(theme.Blue)
			} else {
				style = style.Foreground(theme.Text)
			}
		}

		// Cursor (highlighted row)
		if currentFile && v.applyHighlight {
			// Override background for the cursor line
			style = style.
				Background(theme.Surface2).
				Foreground(theme.Text).
				Bold(true)

			// If it was selected, maybe make it distinctive?
			if selected {
				style = style.Foreground(theme.Yellow)
			}
		}

		// Render
		renderedName := style.Render(displayName)

		// Fill remaining width if it's the cursor line to create a bar effect
		// Note: We pad to v.maxWidth. Since we ensured content <= v.maxWidth,
		// this padding will fill the rest of the line.
		if currentFile && v.applyHighlight {
			// Calculate width
			width := lipgloss.Width(renderedName)
			if width < v.maxWidth {
				padding := strings.Repeat(" ", v.maxWidth-width)
				renderedName += style.Render(padding)
			}
		}

		v.sb.WriteString(renderedName)

		if i != viewportEnd-1 {
			v.sb.WriteString("\n")
		}
	}

	if len(v.entries) == 0 {
		msg := theme.DefaultTheme.StatusInfo.Render("No files")
		v.sb.WriteString(lipgloss.NewStyle().Width(v.maxWidth).Align(lipgloss.Center).Render(msg))
	}

	return v.sb.String()
}

func (v *View) ChDir(dir filesys.Dir, state State) {
	v.path = dir.Path()
	v.allEntries = dir.Entries()
	sortEntries(v.collator, v.allEntries)
	v.filterEntries()

	v.SelectFileByName(state.SelectedName)
	if state.ViewportStart != 0 {
		v.viewportStart = state.ViewportStart
	}

	v.cursorIndex = min(v.cursorIndex, len(v.entries))

	slog.Debug("ChDir",
		"entriesCount", len(v.entries),
		"cursorIndex", v.cursorIndex,
		"viewportStart", v.viewportStart,
		"state.ViewportStart", state.ViewportStart,
		"state.selectedName", state.SelectedName,
		"path", v.path)
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
	slog.Debug("Setting max dims", "rows", rows, "cols", cols)

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

func isHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

// sortEntries orders entries to match `ls -hNlA --color=auto --group-directories-first`
// (same as `ls -1A --color=auto --group-directories-first`) using only Go:
// directories first, then locale-aware name comparison.
func sortEntries(coll *collate.Collator, entries []filesys.DirEntry) {
	slices.SortStableFunc(entries, func(e1, e2 filesys.DirEntry) int {
		aDir := e1.IsDir()
		bDir := e2.IsDir()
		if aDir != bDir {
			if aDir {
				return -1
			}
			return 1
		}
		return coll.CompareString(e1.Name(), e2.Name())
	})
}
