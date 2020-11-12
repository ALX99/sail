package cmd

// All defined commands that fly can handle
const (
	MoveUp Cmd = iota
	MoveDown
	MoveLeft
	MoveRight
	MoveBottom
	MoveTop
	Quit
	MarkSelection
	ToggleCommandMenu

	DirCandy
	DrawBox
	IndentMarks
	IndentAll
	Rainbow
)
