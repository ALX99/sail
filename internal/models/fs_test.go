package models

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

type mockOS struct {
	files map[string][]fs.DirEntry
}

func (m mockOS) ReadDir(path string) ([]fs.DirEntry, error) {
	if files, ok := m.files[path]; ok {
		return files, nil
	}
	return nil, os.ErrNotExist
}

func (m mockOS) RemoveAll(fPath string) error {
	dir := filepath.Dir(fPath)
	files, ok := m.files[dir]
	if !ok {
		return os.ErrNotExist
	}

	if !slices.ContainsFunc(files, func(f fs.DirEntry) bool { return f.Name() == filepath.Base(fPath) }) {
		return os.ErrNotExist
	}

	m.files[dir] = slices.DeleteFunc(files, func(f fs.DirEntry) bool {
		return f.Name() == filepath.Base(fPath)
	})

	return nil
}

func (m mockOS) rename(oldPath, newPath string) error {
	dir := filepath.Dir(oldPath)
	files, ok := m.files[dir]
	if !ok {
		return os.ErrNotExist
	}

	i := slices.IndexFunc(files, func(f fs.DirEntry) bool { return f.Name() == filepath.Base(oldPath) })
	if i == -1 {
		return os.ErrNotExist
	}

	oldFile := files[i]
	m.files[dir] = slices.Delete(files, i, i+1)

	dir = filepath.Dir(newPath)
	if _, ok := m.files[dir]; !ok {
		m.files[dir] = []fs.DirEntry{}
	}

	m.files[dir] = append(m.files[dir], oldFile)
	return nil
}

func (m *mockOS) addFromModel(model Model) {
	if m.files == nil {
		m.files = make(map[string][]fs.DirEntry)
	}
	m.files[model.cwd] = make([]fs.DirEntry, len(model.files))
	copy(m.files[model.cwd], model.files)
}
