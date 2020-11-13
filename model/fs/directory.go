package fs

// todo different ways to sort files
// todo currently the UI has access to the whole of Directory structure including the setter method, it should only use an interface
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

// GetParentDirectory returns the parent directory
func (d Directory) GetParentDirectory() Directory {
	parentDir := filepath.Dir(d.path)
	// Return empty directory if we're trying to
	// get the parent from root directory
	if parentDir == d.path {
		return GetEmptyDirectory()
	}
	nD := GetDirectory(parentDir)

	// Set the right selection
	selectionName := filepath.Base(d.path)
	if files, err := nD.GetFiles(); err == nil {
		for i, f := range files {
			if !f.f.IsDir() {
				continue
			}
			if f.f.Name() == selectionName {
				nD.selection = i
				return nD
			}
		}
	}
	return nD
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
	return Directory{}
}
