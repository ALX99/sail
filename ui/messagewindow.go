package ui

import (
	"github.com/alx99/fly/ui/pos"
	"github.com/gdamore/tcell/v2"
)

type msgWindow struct {
	a       pos.Area
	st      tcell.Style
	visible bool
	msg     string
}

func createMsgWindow(a pos.Area) *msgWindow {
	cw := &msgWindow{a: a}
	return cw
}

// SetArea sets the area where the messagewindow is
// allowed to draw
func (mw *msgWindow) SetArea(area pos.Area) {
	mw.a = area
}

func (mw *msgWindow) setMessage(message string, st tcell.Style) {
	mw.st = st
	mw.msg = message
}

func (mw msgWindow) Draw(screen tcell.Screen) {
	// Don't do anything if it isn't set to visible
	if !mw.visible {
		return
	}
	x, y := mw.a.GetXStart(), mw.a.GetYStart()

	for _, c := range mw.msg {
		screen.SetContent(x, y, c, nil, mw.st)
		x++
		// Todo does not work
		if x == mw.a.GetXMax() {
			y++
			x = mw.a.GetXStart()
		}
	}
}
