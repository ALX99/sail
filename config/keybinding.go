package config

import (
	"fmt"

	"github.com/alx99/fly/cmd"
	"github.com/gdamore/tcell/v2"
)

// Default keybindings
var keybindings = map[Key]KeyBinding{
	"e":                      {cmd.MoveUp, nil},
	keyToKey[tcell.KeyUp]:    {cmd.MoveUp, nil},
	"n":                      {cmd.MoveDown, nil},
	keyToKey[tcell.KeyDown]:  {cmd.MoveDown, nil},
	"k":                      {cmd.MoveLeft, nil},
	keyToKey[tcell.KeyLeft]:  {cmd.MoveLeft, nil},
	"i":                      {cmd.MoveRight, nil},
	keyToKey[tcell.KeyRight]: {cmd.MoveRight, nil},
	"g":                      {nil, map[Key]KeyBinding{"g": {cmd.MoveTop, nil}}},
	"G":                      {cmd.MoveBottom, nil},
	keyToKey[tcell.KeyEsc]:   {cmd.Quit, nil},
	"q":                      {cmd.Quit, nil},
	" ":                      {cmd.MarkSelection, nil},
	":":                      {cmd.ToggleCommandMenu, nil},
}

// KeyBinding maps keys to commands
type KeyBinding struct {
	Command cmd.Command
	m       map[Key]KeyBinding
}

// MatchCommand returns a list of matches
func MatchCommand(k Key, m map[Key]KeyBinding) (cmd.Command, map[Key]KeyBinding) {
	// If the received map is nil we'll
	// query all defined keybindings
	if m == nil {
		m = keybindings
	}

	return m[k].Command, m[k].m
}

// EventKeyToKey converts an EventKey to a Key
func EventKeyToKey(e *tcell.EventKey) Key {
	if e.Key() == tcell.KeyRune {
		return Key(fmt.Sprintf("%c", e.Rune()))
	}
	return keyToKey[e.Key()]
}
