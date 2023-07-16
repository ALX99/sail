package state

type State struct {
	markedFiles map[string]any
}

func NewState() *State {
	return &State{
		markedFiles: map[string]any{},
	}
}

// Select a file or directory
func (n *State) Select(path string) {
}
