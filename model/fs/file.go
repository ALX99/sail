package fs

import "os"

// File represents a file in the filesystem
type File struct {
	f        os.FileInfo
	selected bool
	marked   bool
	invis    bool
}

func createFile(fInfo os.FileInfo) File {
	return File{f: fInfo}
}

// SetMarked marks the file
func (f *File) SetMarked(marked bool) {
	f.marked = marked
}

// GetFileInfo returns the file's FileInfo
func (f File) GetFileInfo() os.FileInfo {
	return f.f
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
