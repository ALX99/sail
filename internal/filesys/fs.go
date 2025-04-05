package filesys

import (
	"io"
	"iter"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FS struct {
	selectedFiles map[string]struct{}
	*sync.RWMutex
}

func NewFS() *FS {
	return &FS{
		selectedFiles: make(map[string]struct{}),
		RWMutex:       &sync.RWMutex{},
	}
}

// Select a file
func (fs *FS) Select(path string) {
	fs.Lock()
	slog.Debug("Selected file", "path", path)
	fs.selectedFiles[path] = struct{}{}
	fs.Unlock()
}

func (fs *FS) Deselect(path string) {
	fs.Lock()
	slog.Debug("Deselected file", "path", path)
	delete(fs.selectedFiles, path)
	fs.Unlock()
}

func (fs *FS) IsSelected(path string) bool {
	fs.RLock()
	_, ok := fs.selectedFiles[path]
	fs.RUnlock()
	return ok
}

func (fs *FS) Selections() iter.Seq[string] {
	fs.RLock()
	defer fs.RUnlock()
	return maps.Keys(fs.selectedFiles)
}

func (fs *FS) DeleteSelections() error {
	fs.Lock()
	defer fs.Unlock()

	for path := range fs.selectedFiles {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		slog.Info("Deleted", "path", path)
		delete(fs.selectedFiles, path)
	}

	return nil
}

func (fs *FS) MoveSelections(dst string) error {
	fs.Lock()
	defer fs.Unlock()

	for path := range fs.selectedFiles {
		if err := os.Rename(path, filepath.Join(dst, filepath.Base(path))); err != nil {
			return err
		}
		delete(fs.selectedFiles, path)
	}

	return nil
}

func (fs *FS) CopySelections(dst string) error {
	fs.Lock()
	defer fs.Unlock()

	for path := range fs.selectedFiles {
		if err := CopyAll(path, dst); err != nil {
			return err
		}
		delete(fs.selectedFiles, path)
	}

	return nil
}

// CopyAll copies all files in the given path to the new directory
func CopyAll(src, dst string) (err error) {
	slog.Info("Copy", "src", src, "dst", dst)

	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return copyFile(src, filepath.Join(dst, filepath.Base(src)))
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Create the destination directory
	dst = filepath.Join(dst, filepath.Base(src))
	if filepath.Clean(dst) == filepath.Clean(src) {
		dst += "_copy"
	}

	if err = os.Mkdir(dst, stat.Mode()); err != nil {
		return err
	}

	defer func() {
		if err != nil { // if there was an error, attempt to clean up
			if err2 := os.RemoveAll(dst); err2 != nil {
				slog.Error("Failed to remove directory", "error", err2, "path", dst)
			}
		}
	}()

	for _, entry := range entries {
		if err = CopyAll(filepath.Join(src, entry.Name()), dst); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(oldPath, newPath string) (err error) {
	if filepath.Clean(oldPath) == filepath.Clean(newPath) {
		newFileName := strings.TrimRight(filepath.Base(newPath), filepath.Ext(newPath)) + "_copy"
		newPath = filepath.Join(filepath.Dir(newPath), newFileName+filepath.Ext(newPath))
	}

	src, err := os.Open(oldPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(newPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	defer func() {
		if err != nil { // if there was an error, attempt to clean up
			if err2 := os.Remove(newPath); err2 != nil {
				slog.Error("Failed to remove file", "error", err2, "path", newPath)
			}
		}
	}()

	_, err = io.Copy(dst, src)
	return err
}
