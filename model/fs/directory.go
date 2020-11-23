package fs

// todo different ways to sort files
import (
	"io/ioutil"
	"path/filepath"
	"time"
)

// Directory represents a directory in the filesystem
type Directory struct {
	path      string
	files     []File
	selection int
	err       error
	queried   time.Time
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
	return Directory{path: path, files: files, queried: time.Now()}
}

// GetEmptyDirectory returns an empty Directory
func GetEmptyDirectory() Directory {
	return Directory{files: []File{}}
}

// GetQueryTime retrieves the time the fs was queried
func (d Directory) GetQueryTime() time.Time {
	return d.queried
}

// getMarkedFiles retries the marked filenames
func (d Directory) getMarkedFiles() []string {
	var marks []string
	for _, f := range d.files {
		if f.marked {
			marks = append(marks, f.f.Name())
		}
	}
	return marks
}

// Refresh refreshes the filelist
func (d *Directory) Refresh() {
	d.queried = time.Now()
	var files []File
	fInfos, err := ioutil.ReadDir(d.path)
	if err != nil {
		d.err = err
		return
	}

	for _, fInfo := range fInfos {
		files = append(files, createFile(fInfo))
	}
	marks := d.getMarkedFiles()
	sel := d.GetSelectedFile().f.Name()
	d.files = files
	d.setMarkedFiles(marks)
	d.SetSelectedFile(sel)
}

// setMarkedFiles sets the marked files
func (d *Directory) setMarkedFiles(filename []string) {
	for i, f := range d.files {
		for _, fn := range filename {
			if f.f.Name() == fn {
				d.files[i].marked = true
				break
			}
		}
	}
}

// SetNextSelection selects the next file
func (d *Directory) SetNextSelection() {
	fCount := len(d.files)
	d.selection = (d.selection + 1) % fCount
	for d.files[d.selection].invis {
		d.selection = (d.selection + 1) % fCount
	}
}

// SetPrevSelection selects the previous file
func (d *Directory) SetPrevSelection() {
	fCount := len(d.files)
	d.selection = (d.selection - 1 + fCount) % fCount
	for d.files[d.selection].invis {
		d.selection = (d.selection - 1 + fCount) % fCount
	}
}

// MarkBottom marks the last non invisible file
func (d *Directory) MarkBottom() {
	d.selection = 0
	d.SetPrevSelection()
}

// MarkTop marks the first non invisible file
func (d *Directory) MarkTop() {
	d.selection = len(d.files) - 1
	d.SetNextSelection()
}

// ToggleHiddenInvis toggles the insvisible
// field on hidden files
func (d *Directory) ToggleHiddenInvis() {
	for i := 0; i < len(d.files); i++ {
		if d.files[i].f.Name()[0:1] == "." {
			d.files[i].invis = !d.files[i].invis
		}
	}
	if !d.IsEmpty() && d.files[d.selection].invis {
		d.SetNextSelection()
	}
}
