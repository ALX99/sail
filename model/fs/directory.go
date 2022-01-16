package fs

// todo different ways to sort files
import (
	"io/fs"
	"path/filepath"
	"time"

	"github.com/alx99/fly/config"
)

// Directory represents a directory in the filesystem
type Directory struct {
	path  string
	files map[int]File

	err       error
	selection int
	allInvis  bool

	queried   time.Time
	dirConfig config.DirConfig
}

// GetDirState returns the dirState
func (d Directory) GetDirState() config.DirConfig {
	return d.dirConfig
}

// GetEmptyDirectory returns an empty directory
func GetEmptyDirectory() Directory {
	files := make(map[int]File, 1)
	files[0] = File{f: fakefileinfo{name: "empty"}, selected: true}
	return Directory{files: files}
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
	if d.IsEmpty() {
		return File{f: fakefileinfo{name: "empty"}, selected: true}
	}
	return d.files[d.selection]
}

// GetPath returns the full path to the directory
func (d Directory) GetPath() string {
	return d.path
}

// ToggleMarked toggles the mark on the currently
// selected file if it exists
func (d *Directory) ToggleMarked() {
	if len(d.files) == 0 {
		return
	}
	file, ok := d.files[d.selection]
	if !ok {
		panic("programmer error")
	}
	file.marked = !file.marked
	d.files[d.selection] = file
}

// GetFiles returns the files inside the directory
func (d Directory) GetFiles() (map[int]File, error) {
	if d.err == nil && len(d.files) > 0 {
		return d.files, nil
	}

	files := make(map[int]File, 1)
	files[0] = File{f: fakefileinfo{name: "empty"}, selected: true}
	return files, d.err
}

// CheckForParent checks if there are any parent directories
func (d Directory) CheckForParent() bool {
	return d.path != filepath.Dir(d.path)
}

// SetSelectedFile sets selection based on filename
func (d *Directory) SetSelectedFile(filename string) {
	for i, f := range d.files {
		if f.f.Name() == filename {
			d.setFileSelected(i, true)
			return
		}
	}
	d.SelectTop()
}

// GetDirectory returns the directory from the full path
func GetDirectory(path string, conf config.DirConfig) Directory {
	files := make(map[int]File)
	fInfos, err := readDir(path)
	if err != nil {
		return Directory{path: path, err: err}
	}

	var i int
	var fInfo fs.FileInfo
	for i, fInfo = range fInfos {
		files[i] = createFile(fInfo, path)
	}
	d := Directory{path: path, files: files, queried: time.Now()}
	if i > 0 {
		d.setFileSelected(0, true)
	}
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

	files := make(map[int]File, len(fInfos))
	for i, fInfo := range fInfos {
		files[i] = createFile(fInfo, d.path)
	}

	if !d.IsEmpty() {
		prevSel = d.GetSelectedFile().f.Name()
	}

	prevMarkedFiles := d.getMarkedFiles()
	d.files = files
	d.setMarkedFiles(prevMarkedFiles)

	// This will force .SetDirConfig to go through all the files again
	d.dirConfig = config.DirConfig{}
	d.SetDirConfig(conf)

	if prevSel != "" {
		d.SetSelectedFile(prevSel)
	}
}

// setMarkedFiles sets the marked files
func (d *Directory) setMarkedFiles(filenames []string) {
	for _, filename := range filenames {
		for i, f := range d.files {
			if f.f.Name() == filename {
				f.marked = true
				d.files[i] = f
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
	// Can't select anything if the directory is
	// empty to the user
	if d.IsEmpty() {
		return
	}

	fCount := len(d.files)
	d.setFileSelected((d.selection+i+fCount)%fCount, true)

	for d.files[d.selection].invis {
		d.setFileSelected((d.selection+i+fCount)%fCount, true)
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
			file := d.files[i]
			file.invis = hideHidden
			d.files[i] = file
			changed++
		}
	}

	d.allInvis = hideHidden && changed == fCount

	// Select next file if the current file
	// became invis
	if fCount != 0 && d.files[d.selection].invis {
		d.SetNextSelection()
	}
}

func (d *Directory) setFileSelected(i int, selected bool) {
	f, ok := d.files[i]
	if !ok {
		panic("programmer error")
	}
	if selected {
		// Unselect previous file
		if d.files[d.selection].selected {
			d.setFileSelected(d.selection, false)
		}
		d.selection = i
	}
	f.selected = selected
	d.files[d.selection] = f
}
