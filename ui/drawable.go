package ui

import (
	"github.com/alx99/fly/ui/pos"
	"github.com/gdamore/tcell/v2"
)

type drawable interface {
	Draw(screen tcell.Screen)

	SetArea(area pos.Area)
}
