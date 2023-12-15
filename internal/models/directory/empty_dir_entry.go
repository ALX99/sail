package directory

import "io/fs"

type emptyDirEntry struct{}

func (d emptyDirEntry) Name() string               { return "empty" }
func (d emptyDirEntry) IsDir() bool                { return false }
func (d emptyDirEntry) Type() fs.FileMode          { return fs.FileMode(0) }
func (d emptyDirEntry) Info() (fs.FileInfo, error) { return nil, nil }
