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

// CheckSelected checks if the file is selected
func (f File) CheckSelected() bool {
	return f.selected
}

// CheckMarked checks if the file is marked
func (f File) CheckMarked() bool {
	return f.marked
}

// CheckInvis checks if the file
// is marked as invisible
func (f File) CheckInvis() bool {
	return f.invis
}
