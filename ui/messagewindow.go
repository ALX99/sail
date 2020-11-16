package ui

import (
	"github.com/alx99/fly/ui/pos"
	"github.com/gdamore/tcell/v2"
)

type msgWindow struct {
	a   pos.Area
	s   tcell.Screen
	st  tcell.Style
	msg string
}

func createMsgWindow(a pos.Area, s tcell.Screen) *msgWindow {
	cw := &msgWindow{a: a, s: s}
	return cw
}

// sets the allowed area where we are allowed to render
func (mw *msgWindow) SetPos(start, end pos.Coord) {
	mw.a.UpdateArea(start, end)
}
func (mw *msgWindow) setMessage(message string, st tcell.Style) {
	mw.st = st
	mw.msg = message
}

func (mw msgWindow) show() {
	x, y := mw.a.GetXStart(), mw.a.GetYStart()

	for _, c := range mw.msg {
		mw.s.SetContent(x, y, c, nil, mw.st)
		x++
		if x == mw.a.GetXMax() {
			y++
			x = mw.a.GetXStart()
		}
	}
}
