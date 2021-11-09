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

	fCount := len(files)
	ym := fw.a.GetYMax()
	xs := fw.a.GetXStart()
	ys := fw.a.GetYStart()

	s := 0
	for i, x := 0, xs; i <= ym && i < fCount; i, x = i+1, xs {
		f := files[i]
		// Don't render "invisible" files
		if f.IsInvis() {
			// Here we don't render anything so let's increase yMax as a hack
			ym++
			s++
			continue
		}
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
					fw.s.SetContent(x, ys+i-s, '+', nil, tcell.StyleDefault.Foreground(fg))
				} else {
					fw.s.SetContent(x, ys+i-s, '+', nil, tcell.StyleDefault)
				}
				// fw.s.SetContent(x, ys+i, ' ', nil, fStyle.Background(fg))
			}
		}
		// If we are pushing an extra space in the beginning
		// we have to increase the starting position and
		// decrease xMax
		if c.IndentAll || (c.IndentMarks && f.IsMarked()) {
			x++
			localXMax--
		}
		limit := util.Min(localXMax, fLen-1)
		j := 0
		for _ = 0; j <= limit; j, x = j+1, x+1 {
			fw.s.SetContent(x, ys+i-s, rune(fName[j]), nil, fStyle)
		}

		if j < fLen {
			// Did not manage to render full filename
			fw.s.SetContent(x-1, ys+i-s, '~', nil, fStyle)
		} else if c.DirCandy && f.GetFileInfo().IsDir() && j-1 < localXMax {
			// If extra space is available
			fw.s.SetContent(x, ys+i-s, '/', nil, tcell.StyleDefault)
		}
	}
}
