package model

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/alx99/fly/config"
	"github.com/alx99/fly/logger"
	"github.com/alx99/fly/model/fs"
)

// Logger identifier
const id = "MDL"

type dirObserver chan<- DirState
type dirRole int

const (
	parentDir dirRole = iota
	workingDir
	childDir
)

// todo notify loading so it can display loading indicator?

// Model model controls exposed to the controller
type Model interface {
	Navigate(d Direction)
	MarkFile()
	AddDirObserver(dirObserver)
	ToggleShowHidden()
}

// CreateModel creates a new model
func CreateModel() (Model, error) {
	m := model{d: &DirState{}, dirCache: make(map[string]*fs.Directory), dirConfig: config.DirConfig{}}
	if err := m.start(); err != nil {
		return nil, err
	}
	return &m, nil
}

// ? Do we want to cache more directories than just those displayed
type model struct {
	d         *DirState
	observers []dirObserver
	dirCache  map[string]*fs.Directory
	dirConfig config.DirConfig
	// If the child directory has changed
	cacheCD bool

	// The number of requests for different directories
	cdRequest *int32
	wdRequest *int32
	pdRequest *int32

	// Lock for the cache
	// ! It is currently only possible to have at most 1 reader
	// ! so there's no reason to use RWMutex
	cacheLock sync.Mutex
}

func (m *model) start() error {
	logger.LogMessage(id, "Starting", logger.DEBUG)
	wd, err := os.Getwd()
	if err != nil {
		logger.LogError(id, "Failed to get initial wd", err)
		return err
	}
	m.cacheCD = true
	m.cdRequest = new(int32)
	m.wdRequest = new(int32)
	m.pdRequest = new(int32)

	// ! Does currently not load wd async, should be fixed in future
	m.d.wd = fs.GetDirectory(wd, m.dirConfig)
	m.getandsetDir(filepath.Dir(m.d.wd.GetPath()), filepath.Base(m.d.wd.GetPath()), parentDir, *m.pdRequest)
	m.setCD()
	return nil
}

func (m model) logCurrentDirState() {
	if _, e := m.d.cd.GetFiles(); e == nil {
		logger.LogMessage(id, "cd: "+m.d.cd.GetPath(), logger.DEBUG)
	} else {
		logger.LogMessage(id, "cderr: "+e.Error(), logger.DEBUG)
	}
	if _, e := m.d.wd.GetFiles(); e == nil {
		logger.LogMessage(id, "wd: "+m.d.wd.GetPath(), logger.DEBUG)
	} else {
		logger.LogMessage(id, "wderr: "+e.Error(), logger.DEBUG)
	}
	if _, e := m.d.pd.GetFiles(); e == nil {
		logger.LogMessage(id, "pd: "+m.d.pd.GetPath(), logger.DEBUG)
	} else {
		logger.LogMessage(id, "pderr: "+e.Error(), logger.DEBUG)
	}
}

func (m *model) setCD() {
	atomic.AddInt32(m.cdRequest, 1)
	// Only cache when needed
	if m.cacheCD {
		m.cacheDir(m.d.cd)
	}
	if m.d.wd.GetSelectedFile().GetFileInfo().IsDir() {
		m.cacheCD = true
		path := ""
		if m.d.wd.GetPath() == "/" {
			path = "/" + m.d.wd.GetSelectedFile().GetFileInfo().Name()
		} else {
			path = m.d.wd.GetPath() + "/" + m.d.wd.GetSelectedFile().GetFileInfo().Name()
		}
		m.getandsetDir(path, "", childDir, *m.cdRequest)
	} else {
		m.cacheCD = false
	}
}

func (m *model) ToggleShowHidden() {
	m.cacheCD = true
	m.dirConfig.HideHidden = !m.dirConfig.HideHidden
	m.d.cd.SetDirConfig(m.dirConfig)
	m.d.wd.SetDirConfig(m.dirConfig)
	m.d.pd.SetDirConfig(m.dirConfig)
	m.notifyObservers()
}

func (m *model) Navigate(d Direction) {
	if d != Left && m.d.wd.IsEmpty() {
		logger.LogMessage(id, "Current dir is empty, ignoring navigation commands", logger.DEBUG)
		return
	}
	switch d {
	case Left:
		if m.d.wd.CheckForParent() {
			atomic.AddInt32(m.pdRequest, 1)
			m.cacheDir(m.d.cd)
			m.d.cd = m.d.wd
			m.d.wd = m.d.pd
			if m.d.wd.CheckForParent() {
				m.getandsetDir(filepath.Dir(m.d.wd.GetPath()), filepath.Base(m.d.wd.GetPath()), parentDir, *m.pdRequest)
			} else {
				m.logCurrentDirState()
				d := fs.GetEmptyDirectory()
				d.SetDirConfig(m.dirConfig)
				m.setAndNotify(d, parentDir, *m.pdRequest)
			}
		}

	case Right:
		// todo this if statement should not be needed
		if _, err := m.d.cd.GetFiles(); err != nil {
			logger.LogError(id, "Can't navigate to "+m.d.cd.GetPath(), err)
			return
		} else if !m.d.wd.GetSelectedFile().GetFileInfo().IsDir() {
			logger.LogMessage(id, "Can't navigate into file currently", logger.DEBUG)
			return
		}
		m.cacheDir(m.d.pd)
		m.d.pd = m.d.wd
		m.d.wd = m.d.cd
		m.setCD()
		// todo if cd is a file show some kinda preview
	case Up:
		// todo enable setting to disable circular selection thing
		m.d.wd.SetPrevSelection()
		m.setCD()
	case Down:
		// todo enable setting to disable circular selection thing
		m.d.wd.SetNextSelection()
		m.setCD()
	case Top:
		m.d.wd.SelectTop()
		m.setCD()
	case Bottom:
		m.d.wd.SelectBottom()
		m.setCD()
	}
	m.d.previewWDFile = !m.d.wd.GetSelectedFile().GetFileInfo().IsDir()
	m.notifyObservers()
}

func (m *model) MarkFile() {
	logger.LogMessage(id, "Marked "+m.d.wd.GetSelectedFile().GetFileInfo().Name(), logger.DEBUG)
	m.d.wd.ToggleMarked()
	m.Navigate(Down)
}

func (m *model) AddDirObserver(o dirObserver) {
	m.observers = append(m.observers, o)
	// Notify them of the current dirstate
	o <- *m.d
}

func (m *model) notifyObservers() {
	for _, o := range m.observers {
		o <- *m.d
	}
}

func (m *model) cacheDir(d fs.Directory) {
	logger.LogMessage(id, "Caching: "+d.GetPath(), logger.DEBUG)
	m.cacheLock.Lock()
	m.dirCache[d.GetPath()] = &d
	m.cacheLock.Unlock()
}

// getandsetDir checks the dirCache before
// getting the directory from the fs
func (m *model) getandsetDir(path, selectedFile string, role dirRole, stamp int32) {
	m.cacheLock.Lock()
	if d, ok := m.dirCache[path]; ok { // Cache hit
		m.cacheLock.Unlock()
		logger.LogMessage(id, "Cache hit: "+d.GetPath(), logger.DEBUG)

		// Directory has changed
		if i, err := os.Stat(path); err == nil && i.ModTime().After(d.GetQueryTime()) {
			logger.LogMessage(id, "Refreshing: "+d.GetPath(), logger.DEBUG)

			// Refresh dir async
			go func() {
				d.Refresh(m.dirConfig)
				d.SetDirConfig(m.dirConfig) // ! To be honest, a race condition can occur here
				m.setAndNotify(*d, role, stamp)
			}()
		} else {
			d.SetDirConfig(m.dirConfig)
			// Don't need to notify UI since we haven't
			// spun up a new goroutine
			switch role {
			case parentDir:
				m.d.pd = *d
			case workingDir:
				m.d.wd = *d
			case childDir:
				m.d.cd = *d
			}
		}
	} else { // No cache hit
		m.cacheLock.Unlock()
		go func() {
			d := fs.GetDirectory(path, m.dirConfig)
			// There's no telling how long the above call will take
			// so we have to set the dirconfig again in case it has changed.
			// This isn't too inefficient since it won't have any effect in case
			// the dirconfig is the same
			d.SetDirConfig(m.dirConfig) // ! To be honest, a race condition can occur here
			if role == parentDir {
				d.SetSelectedFile(selectedFile)
			}
			m.setAndNotify(d, role, stamp)
		}()
	}
}
func (m *model) setAndNotify(d fs.Directory, role dirRole, stamp int32) {
	switch role {
	case parentDir:
		if stamp == *m.pdRequest {
			m.d.pd = d
			m.notifyObservers()
			return
		}
	case workingDir:
		if stamp == *m.wdRequest {
			m.d.wd = d
			m.notifyObservers()
			return
		}
	case childDir:
		if stamp == *m.cdRequest {
			m.d.cd = d
			m.notifyObservers()
			return
		}
	}
	// Cache it if it isn't the most recent request
	m.cacheDir(d)
}
