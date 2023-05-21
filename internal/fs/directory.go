package fs

import (
	"os"
	"sort"
)

type Directory struct {
	files           []File
	fileCount       int
	showHiddenFiles bool
}

func NewDirectory(path string) (Directory, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return Directory{}, err
	}

	f := Directory{
		showHiddenFiles: true,
		fileCount:       len(files),
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

// ToggleShowHiddenFiles toggles showing hidden files
// and returns true if it is now showing hidden files
func (d *Directory) ToggleShowHiddenFiles() bool {
	d.showHiddenFiles = !d.showHiddenFiles

	for i := 0; i < d.fileCount; i++ {
		if d.showHiddenFiles && d.files[i].hidden && !d.files[i].visible {
			d.files[i].visible = true
		}
		if !d.showHiddenFiles && d.files[i].hidden && d.files[i].visible {
			d.files[i].visible = false
		}
	}
	return d.showHiddenFiles
}

// VisibleFiles returns the files visible to the user
func (d *Directory) VisibleFiles() []File {
	files := []File{}
	for i := 0; i < d.fileCount; i++ {
		if d.files[i].visible {
			files = append(files, d.files[i])
		}
	}
	return files
}

func (d *Directory) Files() []File {
	return d.files
}

// GetFileAtIndex returns the file at a certain index i
func (d Directory) GetFileAtIndex(i int) File {
	return d.files[i]
}
