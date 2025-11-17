package filesys

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

type DirLoadedMsg struct {
	ReqID      int
	Dir        Dir
	ParentDir  Dir
	SelectName string
}

type ChildLoadedMsg struct {
	ReqID int
	Dir   Dir
}

func LoadDirCmd(reqID int, targetPath, selectName string) tea.Cmd {
	return func() tea.Msg {
		dir, err := NewDir(targetPath)
		if err != nil {
			return err
		}

		parentPath := filepath.Dir(targetPath)
		parentDir, err := NewDir(parentPath)
		if err != nil {
			return err
		}

		return DirLoadedMsg{
			ReqID:      reqID,
			Dir:        dir,
			ParentDir:  parentDir,
			SelectName: selectName,
		}
	}
}

func LoadChildCmd(reqID int, childPath string) tea.Cmd {
	return func() tea.Msg {
		dir, err := NewDir(childPath)
		if err != nil {
			return err
		}

		return ChildLoadedMsg{
			ReqID: reqID,
			Dir:   dir,
		}
	}
}
