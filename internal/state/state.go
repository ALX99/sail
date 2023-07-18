package state

type State struct {
	selectedFiles map[string]any
}

func NewState() *State {
	return &State{
		selectedFiles: map[string]any{},
	}
}

// ToggleSelect toggles the selection of a file or directory
func (s *State) ToggleSelect(path string) {
	if _, ok := s.selectedFiles[path]; ok {
		delete(s.selectedFiles, path)
	} else {
		s.selectedFiles[path] = true
	}
}

// IsSelected returns true if a file or folder is selected
func (s *State) IsSelected(path string) bool {
	_, ok := s.selectedFiles[path]
	return ok
}
