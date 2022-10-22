package fs

import (
	"os"
	"sort"
)

type Directory struct {
	files           []File
	fileCount       int
	cursor          int
	showHiddenFiles bool
}

func NewDirectory(path string) (Directory, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return Directory{}, err
	}

	f := Directory{
		showHiddenFiles: true,
	}
	f.fileCount = len(files)

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
func (d *Directory) ToggleShowHiddenFiles() {
	d.showHiddenFiles = !d.showHiddenFiles

	for _, f := range d.files {
		if d.showHiddenFiles && f.hidden && !f.visible {
			f.visible = true
		}
		if !d.showHiddenFiles && f.hidden && f.visible {
			f.visible = false
		}
	}
}

// MoveCursorUp moves the cursor up in the directory
func (d *Directory) MoveCursorUp() {
	d.cursor--
}

// MoveCursorDown moves the cursor down in the directory
func (d *Directory) MoveCursorDown() {
	d.cursor++
}

// GetCursorIndex returs the current index of the cursor
func (d Directory) GetCursorIndex() int {
	return d.cursor
}

// GetFileCount returns the total numer of files in the directory
func (d Directory) GetFileCount() int {
	return d.fileCount
}

// GetFileAtCursor returns the file under the cursor
func (d Directory) GetFileAtCursor() File {
	return d.files[d.cursor]
}

// GetFileAtIndex returns the file at a certain index i
func (d Directory) GetFileAtIndex(i int) File {
	return d.files[i]
}

// GetVisibleFileCount returns the number of currently visible files
func (d Directory) GetVisibleFileCount() int {
	return d.fileCount
}
