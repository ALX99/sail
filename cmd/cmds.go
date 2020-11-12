package cmd

// All defined commands that fly can handle
const (
	Nil Cmd = iota
	MoveUp
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
