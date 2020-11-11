package config

// Command represents a command that fly can interpret
type Command int

// All defined commands
const (
	Nil Command = iota
	MoveUp
	MoveDown
	MoveLeft
	MoveRight
	MoveBottom
	MoveTop
	Quit
	MarkSelection
	OpenCommandMenu

	DirCandy
	DrawBox
	IndentMarks
	IndentAll
	Rainbow
)

// todo this
var commandDescription = map[Command]string{}

var strToCmd = map[string]Command{
	// Commands only are "Toggleable"
	"Box":         DrawBox,
	"DirCandy":    DirCandy,
	"IndentMarks": IndentMarks,
	"IndentAll":   IndentAll,
	"Rainbow":     Rainbow,
}

// ParseCommand parses a string to a command
func ParseCommand(command string) (Command, bool) {
	c, ok := strToCmd[command]
	return c, ok
}
