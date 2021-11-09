package ui

import (
	"github.com/alx99/fly/config"
	"github.com/alx99/fly/model/fs"
	"github.com/alx99/fly/ui/pos"
	"github.com/alx99/fly/util"
	"github.com/gdamore/tcell/v2"
)

// FileWindow is a window holding files
type FileWindow struct {
	a pos.Area
	s tcell.Screen
}

// CreateFileWindow Creates a filewindow
func CreateFileWindow(a pos.Area, s tcell.Screen) FileWindow {
	fw := FileWindow{a: a, s: s}
	return fw
}

// SetPos updates the position where the window is allowed to be drawn
func (fw *FileWindow) SetPos(start, end pos.Coord) {
	fw.a.UpdateArea(start, end)
}

// RenderFiles renders a list of files
func (fw *FileWindow) RenderFiles(files map[int]fs.File, c config.UI) {

	yMax := fw.a.GetYMax()
	xStart := fw.a.GetXStart()
	yStart := fw.a.GetYStart()

	visibleFiles := make(map[int]fs.File)

	fCount := 0
	for i := 0; i < len(files); i++ {
		if files[i].IsInvis() {
			continue
		}
		visibleFiles[fCount] = files[i]
		fCount++
	}

	for y, x := yStart, xStart; y <= yMax && y < fCount; y, x = y+1, xStart {
		f := visibleFiles[y]
		fName := f.GetFileInfo().Name()
		fLen := len(fName)
		fStyle := config.GetStyle(f.GetFileInfo())

		// Reverse current selection if the file
		// currently is selected by the UI
		if f.IsSelected() {
			fStyle = fStyle.Reverse(true)
		}

		localXMax := fw.a.GetXMax()

		// todo add setting to how files are marked
		if f.IsMarked() {
			fg, _, _ := fStyle.Decompose()
			fStyle = fStyle.Underline(true).Italic(true).Bold(true)
			if c.IndentMarks {
				if c.Rainbow {
					fw.s.SetContent(x, y, '+', nil, tcell.StyleDefault.Foreground(fg))
				} else {
					fw.s.SetContent(x, y, '+', nil, tcell.StyleDefault)
				}
				x++
				localXMax-- // todo this might be an error to have this?
			}
		}
		// If we are pushing an extra space in the beginning
		// we have to increase the starting position and
		// decrease xMax
		if c.IndentAll && !f.IsMarked() {
			x++
			localXMax-- // todo this might be an error to have this?
		}

		xLimit := util.Min(localXMax, fLen-1)
		xPos := 0
		for xPos = 0; xPos <= xLimit; xPos, x = xPos+1, x+1 {
			fw.s.SetContent(x, y, rune(fName[xPos]), nil, fStyle)
		}

		if xPos < fLen {
			// Did not manage to render full filename
			fw.s.SetContent(x-1, y, '~', nil, fStyle)
		} else if c.DirCandy && f.GetFileInfo().IsDir() && xPos-1 < localXMax {
			// If extra space is available
			fw.s.SetContent(x, y, '/', nil, tcell.StyleDefault)
		}
	}
}
