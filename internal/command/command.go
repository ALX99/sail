package command

type Command uint32

const (
	NUL Command = iota
	RecalculateViews
)
