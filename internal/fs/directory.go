package fs

import "os"

type Directory struct {
	files     []File
	fileCount int
	cursor    int
}

func NewDirectory(path string) (Directory, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return Directory{}, err
	}

	f := Directory{}
	f.fileCount = len(files)

	f.files = make([]File, 0, f.fileCount)
	for _, dEntry := range files {
		f.files = append(f.files, File{dEntry: dEntry})
	}

	return f, nil
}

// MovCursorUp moves the cursor up in the directory
func (d *Directory) MovCursorUp() {
	d.cursor--
}

// MovCursorDown moves the cursor down in the directory
func (d *Directory) MovCursorDown() {
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
