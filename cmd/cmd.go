package cmd

// Command is an interface for different types
// of commands that fly can handle
type Command interface {
}

// Cmd represents a command that fly can interpret
type Cmd int

// todo this
var commandDescription = map[Cmd]string{}

var strToCmd = map[string]Cmd{
	"Box":         DrawBox,
	"DirCandy":    DirCandy,
	"IndentMarks": IndentMarks,
	"IndentAll":   IndentAll,
	"Rainbow":     Rainbow,
}

// ParseCommand parses a string to a command
func ParseCommand(command string) (Cmd, bool) {
	c, ok := strToCmd[command]
	return c, ok
}
