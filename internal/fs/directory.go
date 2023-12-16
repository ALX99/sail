package fs

import (
	"os"
	"sort"
	"strings"
)

type Directory struct {
	path      string
	files     []File
	fileCount int
}

func NewDirectory(path string) (Directory, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return Directory{}, err
	}

	dir := Directory{
		fileCount: len(files),
		path:      path,
	}
	dir.files = make([]File, 0, dir.fileCount)

	for _, dEntry := range files {
		dir.files = append(dir.files, newFile(dEntry))
	}

	// Same sort order as ls
	sort.Slice(dir.files, func(i, j int) bool {
		// Directories come before files
		if dir.files[i].GetDirEntry().IsDir() && !dir.files[j].GetDirEntry().IsDir() {
			return true
		} else if !dir.files[i].GetDirEntry().IsDir() && dir.files[j].GetDirEntry().IsDir() {
			return false
		}

		nameI := strings.TrimPrefix(strings.ToLower(dir.files[i].GetDirEntry().Name()), ".")
		nameJ := strings.TrimPrefix(strings.ToLower(dir.files[j].GetDirEntry().Name()), ".")
		// Sort alphabetically within the same type (directories or files)
		return nameI < nameJ
	})

	return dir, nil
}

func (d Directory) Path() string {
	return d.path
}

func (d Directory) Files() []File {
	return d.files
}

func (d Directory) FileCount() int {
	return d.fileCount
}

// GetFileAtIndex returns the file at a certain index i
func (d Directory) GetFileAtIndex(i int) File {
	return d.files[i]
}
