package fs

// todo different ways to sort files
import (
	"path/filepath"
	"time"

	"github.com/alx99/fly/config"
)

// Directory represents a directory in the filesystem
type Directory struct {
	path      string
	files     []File
	selection int
	err       error
	queried   time.Time
	allInvis  bool
	dirConfig config.DirConfig
}

// GetDirState returns the dirState
func (d Directory) GetDirState() config.DirConfig {
	return d.dirConfig
}

// GetEmptyDirectory returns an empty directory
func GetEmptyDirectory() Directory {
	return Directory{}
}

// IsEmpty checks if the directory
// appears empty to the user
func (d Directory) IsEmpty() bool {
	return d.allInvis || len(d.files) == 0
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
	if !d.allInvis {
		return d.files, nil
	}
	return append(d.files, File{f: fakefileinfo{name: "empty"}}), nil
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
			return
		}
	}
	d.SelectTop()
}

// GetDirectory returns the directory from the full path
func GetDirectory(path string, conf config.DirConfig) Directory {
	var files []File
	fInfos, err := readDir(path)
	if err != nil {
		return Directory{path: path, err: err}
	}

	for _, fInfo := range fInfos {
		files = append(files, createFile(fInfo))
	}
	d := Directory{path: path, files: files, queried: time.Now()}
	d.SetDirConfig(conf)
	return d
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
func (d *Directory) Refresh(conf config.DirConfig) {
	d.dirConfig = conf
	d.queried = time.Now()
	var prevSel string
	fInfos, err := readDir(d.path)
	if err != nil {
		d.err = err
		return
	}

	files := make([]File, len(fInfos))
	for i, fInfo := range fInfos {
		files[i] = createFile(fInfo)
	}

	if !d.allInvis {
		prevSel = d.GetSelectedFile().f.Name()
	}

	marks := d.getMarkedFiles()
	d.files = files
	d.setMarkedFiles(marks)

	// This will force .SetDirConfig to go through all the files again
	d.dirConfig = config.DirConfig{}
	d.SetDirConfig(conf)

	if prevSel != "" {
		d.SetSelectedFile(prevSel)
	}
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
	d.moveSelection(1)
}

// SetPrevSelection selects the previous file
func (d *Directory) SetPrevSelection() {
	d.moveSelection(-1)
}

// precondition: directory is not empty
func (d *Directory) moveSelection(i int) {
	// Can't select anything if all files are invis
	if d.allInvis {
		return
	}
	fCount := len(d.files)
	d.selection = (d.selection + i + fCount) % fCount
	for d.files[d.selection].invis {
		d.selection = (d.selection + i + fCount) % fCount
	}
}

// SelectBottom selects the last non invisible file
func (d *Directory) SelectBottom() {
	d.selection = 0
	d.SetPrevSelection()
}

// SelectTop selects the first non invisible file
func (d *Directory) SelectTop() {
	d.selection = len(d.files) - 1
	d.SetNextSelection()
}

// SetDirConfig sets the dirconfig and updates the directory
// if needed
func (d *Directory) SetDirConfig(config config.DirConfig) {
	if d.dirConfig.HideHidden != config.HideHidden {
		d.setShowHidden(config.HideHidden)
	}
	d.dirConfig = config
}

// setShowHidden sets the insvisible
// field on hidden files
func (d *Directory) setShowHidden(hideHidden bool) {
	fCount := len(d.files)
	changed := 0
	for i := 0; i < fCount; i++ {
		if d.files[i].f.Name()[0:1] == "." {
			d.files[i].invis = hideHidden
			changed++
		}
	}

	if changed == fCount && hideHidden {
		d.allInvis = true
	} else if !hideHidden && fCount > 0 {
		d.allInvis = false
	}

	// Select next file if the current file
	// became invis
	if fCount != 0 && d.files[d.selection].invis {
		d.SetNextSelection()
	}
}
