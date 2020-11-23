package fs

import "os"

// File represents a file in the filesystem
type File struct {
	f      os.FileInfo
	marked bool
	invis  bool
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

// CheckMarked checks if the file is marked
func (f File) CheckMarked() bool {
	return f.marked
}

// CheckInvis checks if the file
// is marked as invinsible
func (f File) CheckInvis() bool {
	return f.invis
}
