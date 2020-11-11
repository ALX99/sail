package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Default keybindings
var keybindings = map[Key]KeyBinding{
	"e":                      {MoveUp, nil},
	keyToKey[tcell.KeyUp]:    {MoveUp, nil},
	"n":                      {MoveDown, nil},
	keyToKey[tcell.KeyDown]:  {MoveDown, nil},
	"k":                      {MoveLeft, nil},
	keyToKey[tcell.KeyLeft]:  {MoveLeft, nil},
	"i":                      {MoveRight, nil},
	keyToKey[tcell.KeyRight]: {MoveRight, nil},
	"g":                      {Nil, map[Key]KeyBinding{"g": {MoveTop, nil}}},
	"G":                      {MoveBottom, nil},
	keyToKey[tcell.KeyEsc]:   {Quit, nil},
	"q":                      {Quit, nil},
	" ":                      {MarkSelection, nil},
	":":                      {OpenCommandMenu, nil},
}

// KeyBinding maps keys to commands
type KeyBinding struct {
	Command Command
	m       map[Key]KeyBinding
}

// MatchCommand returns a list of matches
func MatchCommand(k Key, m map[Key]KeyBinding) (Command, map[Key]KeyBinding) {
	mp := m
	// If the received map is nil we'll
	// query all defined keybindings
	if mp == nil {
		mp = keybindings
	}
	// If the KeyBinding map is nil it doesn't
	// map to any further keybindings, and so it
	// must exist a command there
	if mp[k].m == nil {
		return mp[k].Command, nil
	}
	// Otherwise if it isn't nill there must
	// exist some more keybindings to choose from
	return Nil, mp[k].m
}

// EventKeyToKey converts an EventKey to a Key
func EventKeyToKey(e *tcell.EventKey) Key {
	if e.Key() == tcell.KeyRune {
		return Key(fmt.Sprintf("%c", e.Rune()))
	}
	return keyToKey[e.Key()]
}
