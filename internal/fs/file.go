package fs

import "io/fs"

type File struct {
	dEntry fs.DirEntry
	hidden bool // is a hidden file
}

func newFile(dEntry fs.DirEntry) File {
	f := File{dEntry: dEntry}
	f.hidden = dEntry.Name()[0] == '.'
	return f
}

func (f File) GetDirEntry() fs.DirEntry {
	return f.dEntry
}

// Hidden reports whether the file
// is a hidden one
func (f File) Hidden() bool {
	return f.hidden
}
