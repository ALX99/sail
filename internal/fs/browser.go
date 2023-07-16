package fs

import (
	"sync"
)

type Browser struct {
	markedFiles map[string]any
	cache       map[string]Directory
	sync.RWMutex
}

func NewBrowser() *Browser {
	return &Browser{
		markedFiles: map[string]any{},
		cache:       map[string]Directory{},
	}
}

// Load a file system directory
func (n *Browser) Load(path string) (Directory, error) {
	n.RLock()
	d, ok := n.cache[path]
	if ok {
		return d, nil
	}
	n.RUnlock()

	dir, err := NewDirectory(path)
	if err != nil {
		return d, err
	}

	n.Lock()
	n.cache[path] = dir
	n.Unlock()

	return dir, nil
}
