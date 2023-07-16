package fs

import (
	"os"
	"sort"
)

type Directory struct {
	files     []File
	fileCount int
}

func NewDirectory(path string) (Directory, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return Directory{}, err
	}

	f := Directory{
		fileCount: len(files),
	}
	f.files = make([]File, 0, f.fileCount)

	for _, dEntry := range files {
		f.files = append(f.files, newFile(dEntry))
	}

	sort.Slice(f.files, func(i, j int) bool {
		return f.files[i].GetDirEntry().IsDir()
	})

	return f, nil
}

func (d *Directory) Files() []File {
	return d.files
}

func (d *Directory) FileCount() int {
	return d.fileCount
}

// GetFileAtIndex returns the file at a certain index i
func (d Directory) GetFileAtIndex(i int) File {
	return d.files[i]
}
