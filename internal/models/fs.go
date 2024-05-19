package models

import (
	"io/fs"
	"os"
)

type fsys interface {
	ReadDir(path string) ([]fs.DirEntry, error)
	RemoveAll(path string) error
}

var osi fsys = realOS{}

type realOS struct{}

func (realOS) ReadDir(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(path)
}

func (realOS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}
