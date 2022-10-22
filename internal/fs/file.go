package fs

import "io/fs"

type File struct {
	dEntry fs.DirEntry
	hidden bool
}

func newFile(dEntry fs.DirEntry) File {
	f := File{dEntry: dEntry}
	f.hidden = dEntry.Name()[0] == '.'
	return f
}

func (f File) GetDirEntry() fs.DirEntry {
	return f.dEntry
}
