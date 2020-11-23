package fs

// todo different ways to sort files
import (
	"io/ioutil"
	"path/filepath"
)

// Directory represents a directory in the filesystem
type Directory struct {
	path      string
	files     []File
	selection int
	err       error
	empty     bool
}

// IsEmpty checks if the directory is empty
func (d Directory) IsEmpty() bool {
	return len(d.files) == 0
}

// GetSelection returns the index of the currently selected file
func (d Directory) GetSelection() int {
	return d.selection
}

// GetSelectedFile returns the currently selected file
func (d Directory) GetSelectedFile() File {
	return d.files[d.selection]
}

// SetSelection sets the index of the selected file
func (d *Directory) SetSelection(selection int) {
	// ignore dumb numbers
	if selection >= 0 && selection < len(d.files) {
		d.selection = selection
	}
}

// GetPath returns the full path to the directory
func (d Directory) GetPath() string {
	return d.path
}

// ToggleMarked toggles the mark on the current file
func (d Directory) ToggleMarked() {
	d.files[d.selection].marked = !d.files[d.selection].marked
}

// GetFileCount returns the amount of files in the directory
func (d Directory) GetFileCount() int {
	return len(d.files)
}

// GetFiles returns the files inside the directory
func (d Directory) GetFiles() ([]File, error) {
	if d.err != nil {
		return nil, d.err
	}
	return d.files, nil
}

// CheckForParent checks if there are any parent directories
func (d Directory) CheckForParent() bool {
	return d.path != filepath.Dir(d.path)
}

// SetSelectedFile sets selection based on filename
func (d *Directory) SetSelectedFile(filename string) {
	for i, f := range d.files {
		if f.f.Name() == filename {
			d.selection = i
			break
		}
	}
}

// GetDirectory returns the directory from the full path
func GetDirectory(path string) Directory {
	var files []File
	fInfos, err := ioutil.ReadDir(path)
	if err != nil {
		return Directory{path: path, err: err}
	}

	for _, fInfo := range fInfos {
		files = append(files, createFile(fInfo))
	}
	return Directory{path: path, files: files}
}

// GetEmptyDirectory returns an empty Directory
func GetEmptyDirectory() Directory {
	return Directory{files: []File{}}
}
