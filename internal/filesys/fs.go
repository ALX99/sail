package filesys

import (
	"io"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Selection tracks selected file paths for operations and implements the
// filelist.SelChecker interface.
type Selection struct {
	files map[string]struct{}
}

func NewSelection() *Selection {
	return &Selection{
		files: make(map[string]struct{}),
	}
}

func (s *Selection) Select(path string) {
	if s.files == nil {
		s.files = make(map[string]struct{})
	}
	s.files[path] = struct{}{}
}

func (s *Selection) Deselect(path string) {
	delete(s.files, path)
}

func (s *Selection) Toggle(path string) bool {
	if s.IsSelected(path) {
		s.Deselect(path)
		return false
	}
	s.Select(path)
	return true
}

func (s *Selection) Clear() {
	clear(s.files)
}

func (s *Selection) IsSelected(path string) bool {
	_, ok := s.files[path]
	return ok
}

func (s *Selection) Paths() []string {
	return slices.Collect(maps.Keys(s.files))
}

func DeletePaths(paths []string) error {
	for _, path := range unique(paths) {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		slog.Info("Deleted", "path", path)
	}
	return nil
}

func MovePaths(paths []string, dst string) error {
	for _, path := range unique(paths) {
		if err := os.Rename(path, filepath.Join(dst, filepath.Base(path))); err != nil {
			return err
		}
		slog.Info("Moved", "path", path, "dst", dst)
	}
	return nil
}

func CopyPaths(paths []string, dst string) error {
	for _, path := range unique(paths) {
		if err := CopyAll(path, dst); err != nil {
			return err
		}
	}
	return nil
}

// CopyAll copies all files in the given path to the new directory.
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

	info, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.OpenFile(newPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
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

func unique(paths []string) []string {
	if len(paths) < 2 {
		return paths
	}
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	return out
}
