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
func CreateFileWindow(a pos.Area, s tcell.Screen) *FileWindow {
	fw := &FileWindow{a: a, s: s}
	return fw
}

// SetPos updates the position where the window is allowed to be drawn
func (fw *FileWindow) SetPos(start, end pos.Coord) {
	fw.a.UpdateArea(start, end)
}

// RenderDir renders a directory
func (fw *FileWindow) RenderDir(d fs.Directory, mChan chan<- Message, c config.UI) {
	files, err := d.GetFiles()
	if err != nil {
		mChan <- CreateMessage(err.Error(), true)
		return
	}

	fCount := len(files)
	fileOffset := 0
	ym := fw.a.GetYMax()
	xs := fw.a.GetXStart()
	ys := fw.a.GetYStart()
	sel := d.GetSelection()

	// Offset the files displayed if the selection can't
	// be shown in the amount of lines we have
	for sel > ym+fileOffset {
		fileOffset += ym + 1
	}

	s := 0
	for i, x := 0, xs; i <= ym && i+fileOffset < fCount; i, x = i+1, xs {
		f := files[i+fileOffset]
		// Don't render "invisible" files
		if f.CheckInvis() {
			s++
			continue
		}
		fName := f.GetFileInfo().Name()
		fLen := len(fName)
		fStyle := config.GetStyle(f)

		// Reverse current selection if the file
		// currently is selected by the UI
		if i == sel-fileOffset {
			fStyle = fStyle.Reverse(true)
		}

		localXMax := fw.a.GetXMax()

		// todo add setting to how files are marked
		if f.CheckMarked() {
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
		if c.IndentAll || (c.IndentMarks && f.CheckMarked()) {
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
