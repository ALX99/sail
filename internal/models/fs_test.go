package models

import (
	"io/fs"
	"os"
	"path"
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

func (m mockOS) Remove(fPath string) error {
	dir := path.Dir(fPath)
	files, ok := m.files[dir]
	if !ok {
		return os.ErrNotExist
	}

	if !slices.ContainsFunc(files, func(f fs.DirEntry) bool { return f.Name() == path.Base(fPath) }) {
		return os.ErrNotExist
	}

	m.files[dir] = slices.DeleteFunc(files, func(f fs.DirEntry) bool {
		return f.Name() == path.Base(fPath)
	})

	return nil
}

func (m *mockOS) addFromModel(model Model) {
	if m.files == nil {
		m.files = make(map[string][]fs.DirEntry)
	}
	m.files[model.cwd] = make([]fs.DirEntry, len(model.files))
	copy(m.files[model.cwd], model.files)
}
