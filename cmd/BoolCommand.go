package cmd

// BoolCommand is a command that allows
// setting a boolean value
type BoolCommand struct {
	c         Cmd
	v, hasSet bool
}

// CreateBoolCommand creates a new BoolCommand
func CreateBoolCommand(c Cmd) BoolCommand {
	return BoolCommand{c: c}
}

// GetCommand returns the command
func (b BoolCommand) GetCommand() Cmd {
	return b.c
}

// GetValue returns the value of the boolean
func (b BoolCommand) GetValue() bool {
	return b.v
}

// HasValueSet checks if the command has a value set
func (b BoolCommand) HasValueSet() bool {
	return b.hasSet

}

// SetValue sets the boolean value
func (b *BoolCommand) SetValue(v bool) {
	b.hasSet = true
	b.v = v
}
