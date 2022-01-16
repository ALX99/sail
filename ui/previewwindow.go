package ui

import (
	"os/exec"
	"strings"

	"github.com/alx99/fly/model/fs"
	"github.com/alx99/fly/ui/pos"
	"github.com/gdamore/tcell/v2"
)

// PreviewWindow is a window that can preview files
type PreviewWindow struct {
	a           pos.Area
	previewFile fs.File
	visible     bool
}

// CreatePreviewWindow Creates a previewwindow
func CreatePreviewWindow(a pos.Area) PreviewWindow {
	pw := PreviewWindow{a: a, visible: false}
	return pw
}

// SetArea sets the area where the PreviewWindow can render things
func (pw *PreviewWindow) SetArea(area pos.Area) {
	pw.a = area
}

// SetPreviewFile sets the file to preview
func (pw *PreviewWindow) SetPreviewFile(file fs.File) {
	pw.previewFile = file
}

func (pw *PreviewWindow) generatePreveiw() string {
	cmd := exec.Command("cat", pw.previewFile.GetFilePath())
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return string(stdout)
	}
	return string(stdout)
}

// Draw renders a the PreviewWindow component
func (pw *PreviewWindow) Draw(screen tcell.Screen) {
	if !pw.visible {
		return
	}

	xMax, yMax := pw.a.GetXEnd(), pw.a.GetYEnd()
	preview := strings.SplitAfter(pw.generatePreveiw(), "\n")
	previewLen := len(preview)
	pX, pY := 0, 0

	for y := pw.a.GetYStart(); y <= yMax && pY < previewLen; y, pY, pX = y+1, pY+1, 0 {
		for x := pw.a.GetXStart(); x <= xMax; x, pX = x+1, pX+1 {
			if pX < len(preview[pY]) {
				screen.SetContent(x, y, rune(preview[pY][pX]), nil, tcell.StyleDefault)
			} else {
				screen.SetContent(x, y, ' ', nil, tcell.StyleDefault)
			}
		}
	}
}
