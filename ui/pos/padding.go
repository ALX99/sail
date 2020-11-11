package pos

// Padding can provide padding for windows
type Padding struct {
	start, end, top, bottom int
}

// CreatePadding creates a padding
func CreatePadding(start, end, top, bottom int) Padding {
	return Padding{start: start, end: end, top: top, bottom: bottom}
}
