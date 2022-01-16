package pos

// Area represents an area bounded by the start and end coordinates
type Area struct {
	start, end Coord
	p          Pad

	// Max characters allowed to render in a direction
	xEnd, yEnd int
	// Starting positions
	xStart, yStart int
}

func CreateArea(start, end Coord, p Pad) Area {
	a := Area{start: start, end: end, p: p}
	a.calculateLimits()
	return a
}

// UpdateArea allows for updating the area
func (a *Area) UpdateArea(start, end Coord) {
	a.start = start
	a.end = end
	a.calculateLimits()
}

// calculate starting positions and limits
func (a *Area) calculateLimits() {
	a.yEnd = (a.end.Y - a.p.bottom)
	a.xEnd = (a.end.X - a.p.end)
	a.xStart = a.start.X + a.p.start
	a.yStart = a.start.Y + a.p.top
}

func (a Area) GetStart() Coord {
	return a.start
}
func (a Area) GetEnd() Coord {
	return a.end
}
func (a Area) GetXEnd() int {
	return a.xEnd
}
func (a Area) GetYEnd() int {
	return a.yEnd
}
func (a Area) GetXStart() int {
	return a.xStart
}
func (a Area) GetYStart() int {
	return a.yStart
}
