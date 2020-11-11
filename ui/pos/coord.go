package pos

// Coord represents a coordinate on the screen
type Coord struct {
	X, Y int
}

// NewCoord creates a new Cord
func NewCoord(x, y int) Coord {
	return Coord{x, y}
}
