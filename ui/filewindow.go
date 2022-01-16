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
	a       pos.Area
	visible bool
	files   map[int]fs.File
	config  filewWindowConfig
}
type filewWindowConfig struct {
	indentMarks bool
	indentAll   bool
	dirCandy    bool
	rainbow     bool
}

// CreateFileWindow Creates a filewindow
func CreateFileWindow(a pos.Area) FileWindow {
	fw := FileWindow{a: a, visible: true}
	return fw
}

// SetArea sets the area where the filewindow can render things
func (fw *FileWindow) SetArea(area pos.Area) {
	fw.a = area
}

func (fw *FileWindow) setConfig(config filewWindowConfig) {
	fw.config = config
}

// SetFiles sets the files inside of the filewindow
func (fw *FileWindow) SetFiles(files map[int]fs.File) {
	fw.files = files
}

// Draw renders a the filewindow component
func (fw *FileWindow) Draw(screen tcell.Screen) {
	if !fw.visible {
		return
	}

	yEnd := fw.a.GetYEnd()
	xStart := fw.a.GetXStart()
	yStart := fw.a.GetYStart()
	fCount := 0
	selectionIndex := 0
	offset := 0

	visibleFiles := make(map[int]fs.File)

	// Get all non invisible files and the currently
	// selected file's index
	for i := 0; i < len(fw.files); i++ {
		if fw.files[i].IsInvis() {
			continue
		}
		if fw.files[i].IsSelected() {
			selectionIndex = fCount
		}
		visibleFiles[fCount] = fw.files[i]
		fCount++
	}

	for selectionIndex > yEnd+offset-3 {
		offset++
	}

	for y, x := yStart, xStart; y <= yEnd && y+offset < fCount; y, x = y+1, xStart {
		f, ok := visibleFiles[y+offset]
		if !ok {
			panic("programmer error")
		}
		fName := f.GetFileInfo().Name()
		fLen := len(fName)
		fStyle := config.GetStyle(f.GetFileInfo())

		// Reverse current selection if the file
		// currently is selected by the UI
		if f.IsSelected() {
			fStyle = fStyle.Reverse(true)
		}

		localXMax := fw.a.GetXEnd()

		// todo add setting to how files are marked
		if f.IsMarked() {
			fStyle = fStyle.Underline(true).Italic(true).Bold(true)
			if fw.config.indentMarks {
				if fw.config.rainbow {
					fg, _, _ := fStyle.Decompose()
					screen.SetContent(x, y, '+', nil, tcell.StyleDefault.Foreground(fg))
				} else {
					screen.SetContent(x, y, '+', nil, tcell.StyleDefault)
				}
				x++
				localXMax-- // todo this might be an error to have this?
			}
		}
		// If we are pushing an extra space in the beginning
		// we have to increase the starting position and
		// decrease xMax
		if fw.config.indentAll && !f.IsMarked() {
			x++
			localXMax-- // todo this might be an error to have this?
		}

		xLimit := util.Min(localXMax, fLen-1)
		xPos := 0
		for xPos = 0; xPos <= xLimit; xPos, x = xPos+1, x+1 {
			screen.SetContent(x, y, rune(fName[xPos]), nil, fStyle)
		}

		if xPos < fLen {
			// Did not manage to render full filename
			screen.SetContent(x-1, y, '~', nil, fStyle)
		} else if fw.config.dirCandy && f.GetFileInfo().IsDir() && xPos-1 < localXMax {
			// If extra space is available
			screen.SetContent(x, y, '/', nil, tcell.StyleDefault)
		}
	}
}
