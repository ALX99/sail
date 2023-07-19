package state

import "os"

type State struct {
	selectedFiles map[string]bool
}

func NewState() *State {
	return &State{
		selectedFiles: map[string]bool{},
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

// HasSelectedFiles returns true if there are selected files
func (s *State) HasSelectedFiles() bool {
	return len(s.selectedFiles) > 0
}

// DeletSelectedFiles deletes the selected files
func (s *State) DeleteSelectedFiles() error {
	var err error
	for path := range s.selectedFiles {
		if err = os.RemoveAll(path); err != nil {
			break
		}
		s.selectedFiles[path] = true
	}

	for path, deleted := range s.selectedFiles {
		if deleted {
			delete(s.selectedFiles, path)
		}
	}

	return err
}
