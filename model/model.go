package model

import (
	"os"
	"path/filepath"

	"github.com/alx99/fly/logger"
	"github.com/alx99/fly/model/fs"
)

// Logger identifier
const id = "MDL"

type dirObserver chan<- DirState

// todo notify loading so it can display loading indicator?

// Model model controls exposed to the controller
type Model interface {
	Navigate(d Direction)
	MarkFile()
	AddDirObserver(dirObserver)
}

// CreateModel creates a new model
func CreateModel() (Model, error) {
	m := model{d: DirState{}, dirCache: make(map[string]*fs.Directory)}
	if err := m.start(); err != nil {
		return nil, err
	}
	return &m, nil
}

// ? Do we want to cache more directories than just those displayed
type model struct {
	d         DirState
	observers []dirObserver
	dirCache  map[string]*fs.Directory
}

func (m *model) start() error {
	logger.LogMessage(id, "Starting", logger.DEBUG)
	wd, err := os.Getwd()
	if err != nil {
		logger.LogError(id, "Failed to get initial wd", err)
		return err
	}
	m.d.wd = fs.GetDirectory(wd)
	m.d.pd = fs.GetDirectory(filepath.Dir(m.d.wd.GetPath()))
	m.d.pd.SetSelectedFile(filepath.Base(m.d.wd.GetPath()))
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

// if the selection of wd is a diretory cd is set to that
// otherwise it's set to an empty directory else{

// todo otherwise it will be ignored in the future when we have previews working
func (m *model) setCD() {
	m.cacheDir(m.d.cd)
	if m.d.wd.GetSelectedFile().GetFileInfo().IsDir() {
		if m.d.wd.GetPath() == "/" {
			m.d.cd = m.getDir("/" + m.d.wd.GetSelectedFile().GetFileInfo().Name())
		} else {
			m.d.cd = m.getDir(m.d.wd.GetPath() + "/" + m.d.wd.GetSelectedFile().GetFileInfo().Name())
		}
	} else {
		m.d.cd = fs.GetEmptyDirectory()
	}
}

func (m *model) Navigate(d Direction) {
	switch d {
	case Left:
		if m.d.wd.CheckForParent() {
			m.cacheDir(m.d.cd)
			m.d.cd = m.d.wd
			m.d.wd = m.d.pd
			m.d.pd = m.getDir(filepath.Dir(m.d.wd.GetPath()))
			m.d.pd.SetSelectedFile(filepath.Base(m.d.wd.GetPath()))
		}
		m.logCurrentDirState()

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
		m.logCurrentDirState()
	case Up:
		// todo enable setting to disable circular selection thing
		m.d.wd.SetSelection((m.d.wd.GetSelection() - 1 + m.d.wd.GetFileCount()) % m.d.wd.GetFileCount())
		m.setCD()
	case Down:
		// todo enable setting to disable circular selection thing
		m.d.wd.SetSelection((m.d.wd.GetSelection() + 1) % m.d.wd.GetFileCount())
		m.setCD()
	case Top:
		m.d.wd.SetSelection(0)
		m.setCD()
	case Bottom:
		m.d.wd.SetSelection(m.d.wd.GetFileCount() - 1)
		m.setCD()
	}
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
	o <- m.d
}

func (m model) notifyObservers() {
	for _, o := range m.observers {
		o <- m.d
	}
}

func (m model) cacheDir(d fs.Directory) {
	// Cache non empty directories
	if !d.IsEmpty() {
		logger.LogMessage(id, "Caching: "+d.GetPath(), logger.DEBUG)
		m.dirCache[d.GetPath()] = &d
	}
}

// getDir checks the dirCache before
// getting the directory from the fs
func (m model) getDir(path string) fs.Directory {
	if d, ok := m.dirCache[path]; ok {
		logger.LogMessage(id, "Cache hit: "+d.GetPath(), logger.DEBUG)
		if i, err := os.Stat(path); err == nil && i.ModTime().After(d.GetQueryTime()) {
			logger.LogMessage(id, "Refreshing: "+d.GetPath(), logger.DEBUG)
			d.Refresh()
		}
		return *d
	}
	return fs.GetDirectory(path)
}
