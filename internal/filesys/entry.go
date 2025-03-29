package filesys

import (
	"io/fs"
	"os"
	"path/filepath"
)

type DirEntry struct {
	dirPath string
	fs.DirEntry
}

// Path returns the full path to the entry
func (e DirEntry) Path() string {
	return filepath.Join(e.dirPath, e.Name())
}

func (e DirEntry) ResolveSymlink() (DirEntry, error) {
	if e.Type() != fs.ModeSymlink {
		return e, nil
	}

	linkPath, err := os.Readlink(e.Path())
	if err != nil {
		return DirEntry{}, err
	}

	linkPath = filepath.Clean(linkPath)
	if !filepath.IsAbs(linkPath) {
		linkPath = filepath.Join(filepath.Dir(e.Path()), linkPath)
	}

	info, err := os.Stat(linkPath)
	if err != nil {
		return DirEntry{}, err
	}

	newEntry := DirEntry{
		dirPath:  filepath.Dir(linkPath),
		DirEntry: fs.FileInfoToDirEntry(info),
	}

	return newEntry.ResolveSymlink()
}
