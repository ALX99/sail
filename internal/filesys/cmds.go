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

type FilesDeletedMsg struct{ Paths []string }
type FilesMovedMsg struct{ Paths []string }
type FilesCopiedMsg struct{ Paths []string }

func DeleteCmd(paths []string) tea.Cmd {
	return func() tea.Msg {
		if err := DeletePaths(paths); err != nil {
			return err
		}
		return FilesDeletedMsg{Paths: paths}
	}
}

func MoveCmd(paths []string, dst string) tea.Cmd {
	return func() tea.Msg {
		if err := MovePaths(paths, dst); err != nil {
			return err
		}
		return FilesMovedMsg{Paths: paths}
	}
}

func CopyCmd(paths []string, dst string) tea.Cmd {
	return func() tea.Msg {
		if err := CopyPaths(paths, dst); err != nil {
			return err
		}
		return FilesCopiedMsg{Paths: paths}
	}
}
