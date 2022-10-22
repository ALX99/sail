package fs

import "io/fs"

type File struct {
	dEntry fs.DirEntry
}

func (f File) GetDirEntry() fs.DirEntry {
	return f.dEntry
}
