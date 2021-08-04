package config

import (
	"fmt"

	"github.com/alx99/fly/cmd"
	"github.com/gdamore/tcell/v2"
)

// Default keybindings
var keybindings = KeyBindings{
	Command: nil,
	m: map[Key]KeyBindings{
		"e":                      {cmd.MoveUp, nil},
		keyToKey[tcell.KeyUp]:    {cmd.MoveUp, nil},
		"n":                      {cmd.MoveDown, nil},
		keyToKey[tcell.KeyDown]:  {cmd.MoveDown, nil},
		"k":                      {cmd.MoveLeft, nil},
		keyToKey[tcell.KeyLeft]:  {cmd.MoveLeft, nil},
		"i":                      {cmd.MoveRight, nil},
		keyToKey[tcell.KeyRight]: {cmd.MoveRight, nil},
		"g":                      {nil, map[Key]KeyBindings{"g": {cmd.MoveTop, nil}, "a": {cmd.ToggleCommandMenu, nil}}},
		"G":                      {cmd.MoveBottom, nil},
		keyToKey[tcell.KeyEsc]:   {cmd.Quit, nil},
		"q":                      {cmd.Quit, nil},
		" ":                      {cmd.MarkSelection, nil},
		":":                      {cmd.ToggleCommandMenu, nil},
		".":                      {cmd.ToggleShowHidden, nil},
	},
}

// KeyBindings maps keys to commands
type KeyBindings struct {
	Command cmd.Command
	m       map[Key]KeyBindings
}

// IsSingleKeyBinding returns true if it is a single keybinding
func (kbs KeyBindings) IsSingleKeyBinding() bool {
	return kbs.m == nil && kbs.Command != nil
}

// GetCommand returns the command associated with the keybinding
func (kbs KeyBindings) GetCommand() cmd.Command {
	return kbs.Command
}

// FindMatches returns a list of matching keybindings related to the key e
func (kbs KeyBindings) FindMatches(e *tcell.EventKey) (KeyBindings, bool) {
	res, ok := kbs.m[eventKeyToKey(e)]
	return res, ok
}

// eventKeyToKey converts an EventKey to a Key
func eventKeyToKey(e *tcell.EventKey) Key {
	if e.Key() == tcell.KeyRune {
		return Key(fmt.Sprintf("%c", e.Rune()))
	}
	return keyToKey[e.Key()]
}

// GetAllKeyBindings returns all defined keybindings
func GetAllKeyBindings() KeyBindings {
	return keybindings
}
