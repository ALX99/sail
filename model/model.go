package model

import (
	"os"

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
	m := model{d: DirState{}}
	if err := m.start(); err != nil {
		return nil, err
	}
	return &m, nil
}

// ? Do we want to cache more directories than just those displayed
type model struct {
	d         DirState
	observers []dirObserver
}

func (m *model) start() error {
	logger.LogMessage(id, "Starting", logger.DEBUG)
	wd, err := os.Getwd()
	if err != nil {
		logger.LogError(id, "Failed to get initial wd", err)
		return err
	}
	m.d.wd = fs.GetDirectory(wd)
	m.d.pd = m.d.wd.GetParentDirectory()
	m.setCD()
	return nil
}
func (m model) logCurrentDirState() {
	logger.LogMessage(id, "cd: "+m.d.cd.GetPath(), logger.DEBUG)
	logger.LogMessage(id, "wd: "+m.d.wd.GetPath(), logger.DEBUG)
	logger.LogMessage(id, "pd: "+m.d.pd.GetPath(), logger.DEBUG)
}

// if the selection of wd is a diretory cd is set to that
// otherwise it's set to an empty directory
// todo otherwise it will be ignored in the future when we have previews working
func (m *model) setCD() {
	if m.d.wd.GetSelectedFile().GetFileInfo().IsDir() {
		m.d.cd = fs.GetDirectory(m.d.wd.GetPath() + "/" + m.d.wd.GetSelectedFile().GetFileInfo().Name())
	} else {
		m.d.cd = fs.GetEmptyDirectory()
	}
}

func (m *model) Navigate(d Direction) {
	switch d {
	case Left:
		if m.d.wd.CheckForParent() {
			m.d.cd = m.d.wd
			m.d.wd = m.d.pd
			m.d.pd = m.d.wd.GetParentDirectory()
		}
		m.logCurrentDirState()
	case Right:
		if !m.d.wd.GetSelectedFile().GetFileInfo().IsDir() {
			logger.LogMessage(id, "Can't navigate into file currently", logger.DEBUG)
			return
		}
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

func (m model) MarkFile() {
	logger.LogMessage(id, "Marked "+m.d.wd.GetSelectedFile().GetFileInfo().Name(), logger.DEBUG)
	m.d.wd.ToggleMarked()
	m.Navigate(Down)
	m.notifyObservers()
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
