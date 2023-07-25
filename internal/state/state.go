package state

import (
	"os"
	"path"

	"github.com/rs/zerolog/log"
)

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

// DeleteSelectedFiles deletes the selected files
func (s *State) DeleteSelectedFiles() error {
	var err error
	for path := range s.selectedFiles {
		if err = os.RemoveAll(path); err != nil {
			break
		}
		log.Debug().Str("path", path).Msg("Deleted")
		s.selectedFiles[path] = true
	}

	for path, deleted := range s.selectedFiles {
		if deleted {
			delete(s.selectedFiles, path)
		}
	}

	return err
}

// MoveSelectedFiles moves the selected files
// to the specified directory
func (s *State) MoveSelectedFiles(dirPath string) error {
	var err error
	for oldPath := range s.selectedFiles {
		newPath := path.Join(dirPath, path.Base(oldPath))
		if err = os.Rename(oldPath, newPath); err != nil {
			break
		}
		log.Debug().Str("oldPath", oldPath).Str("newPath", newPath).Msg("Moved")
		s.selectedFiles[oldPath] = true
	}

	for path, moved := range s.selectedFiles {
		if moved {
			delete(s.selectedFiles, path)
		}
	}

	return err
}
