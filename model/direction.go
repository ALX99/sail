package model

// Direction direction of where to move
type Direction int

// Different type of directions
const (
	Up Direction = iota
	Down
	Left
	Right
	Top
	Bottom
)
