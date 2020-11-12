package cmd

// Command is an interface for different types
// of commands that fly can handle
type Command interface {
	GetCommand() Cmd
}

// Cmd represents a command that fly can interpret
type Cmd int

// GetCommand returns the command
func (c Cmd) GetCommand() Cmd {
	// This is ambiguous ans is only here to allow
	// Cmd to implement the Command interface
	return c
}

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
