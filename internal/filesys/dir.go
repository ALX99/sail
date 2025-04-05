package filesys

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
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
			log.Warn().Err(err).
				Str("path", d.path).
				Msg("Error walking directory, ignoring")
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	log.Debug().
		Dur("duration", time.Since(now)).
		Str("path", d.path).
		Msgf("Walk finished")
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
