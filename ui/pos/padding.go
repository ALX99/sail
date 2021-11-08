package pos

// Pad can provide padding for windows
type Pad struct {
	start, end, top, bottom int
}

// CreatePadding creates a padding
func Padding(start, end, top, bottom int) Pad {
	return Pad{start: start, end: end, top: top, bottom: bottom}
}
