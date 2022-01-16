package fs

import (
	"os"
	"path/filepath"
)

// File represents a file in the filesystem
type File struct {
	f        os.FileInfo
	selected bool
	marked   bool
	invis    bool
	path     string
}

func createFile(fInfo os.FileInfo, dirPath string) File {
	return File{f: fInfo, path: filepath.Join(dirPath, fInfo.Name())}
}

// SetMarked marks the file
func (f *File) SetMarked(marked bool) {
	f.marked = marked
}

// GetFileInfo returns the file's FileInfo
func (f File) GetFileInfo() os.FileInfo {
	return f.f
}

// GetFilePath returns the full path to the file
func (f File) GetFilePath() string {
	return f.path
}

// IsSelected checks if the file is selected
func (f File) IsSelected() bool {
	return f.selected
}

// IsMarked checks if the file is marked
func (f File) IsMarked() bool {
	return f.marked
}

// IsInvis checks if the file
// is marked as invisible
func (f File) IsInvis() bool {
	return f.invis
}
