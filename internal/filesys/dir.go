package filesys

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type Dir struct {
	path      string
	fileCount int
	dirCount  int

	entries []DirEntry
}

func NewDir(path string) (Dir, error) {
	path = filepath.Clean(path)
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return Dir{}, err
	}

	fCount, dCount := 0, 0
	for _, entry := range dirEntries {
		if entry.IsDir() {
			dCount++
		} else {
			fCount++
		}
	}

	entries := make([]DirEntry, 0, len(dirEntries))
	for _, entry := range dirEntries {
		entries = append(entries, DirEntry{
			dirPath:  path,
			DirEntry: entry,
		})
	}

	return Dir{
		path:      path,
		fileCount: fCount,
		dirCount:  dCount,
		entries:   entries,
	}, nil
}

func (d Dir) RealSize() (int64, error) {
	now := time.Now()
	var size int64
	err := filepath.Walk(d.path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			slog.Warn("Error walking directory, ignoring",
				"error", err,
				"path", d.path)
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	slog.Debug("Walk finished",
		"duration", time.Since(now),
		"path", d.path)
	return size, err
}

func (d Dir) Path() string {
	return d.path
}

func (d Dir) Counts() (files, folders int) {
	return d.fileCount, d.dirCount
}

func (d Dir) Entries() []DirEntry {
	return d.entries
}
