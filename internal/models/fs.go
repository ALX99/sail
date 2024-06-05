package models

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type fsys interface {
	ReadDir(path string) ([]fs.DirEntry, error)
	RemoveAll(path string) error
	rename(oldPath, newPath string) error
	copyAll(oldPath, newPath string) error
}

var osi fsys = realOS{}

type realOS struct{}

func (realOS) ReadDir(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(path)
}

func (realOS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (realOS) rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (r realOS) copyAll(path, newDir string) error {
	log.Trace().Str("path", path).Str("newDir", newDir).Msg("copy")
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return copyFile(path, filepath.Join(newDir, filepath.Base(path)))
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	dst := filepath.Join(newDir, filepath.Base(path))
	if err := os.Mkdir(dst, stat.Mode()); err != nil {
		return err
	}

	for _, entry := range entries {
		src := filepath.Join(path, entry.Name())
		if err := r.copyAll(src, dst); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(oldPath, newPath string) error {
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

	_, err = io.Copy(dst, src)
	return err
}
