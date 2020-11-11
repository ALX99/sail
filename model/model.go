package model

import (
	"os"

	"github.com/alx99/fly/fs"
	"github.com/alx99/fly/logger"
)

// Direction direction of where to move
type Direction int

// Different type of directions
const (
	Up Direction = iota
	Down
	Left
	Right
	Top
	Bottom
)

// Logger identifier
const id = "MDL"

// todo notify loading so it can display loading indicator?

// Model model controls exposed to the controller
type Model interface {
	Navigate(d Direction)
	GetCD() fs.Directory
	GetWD() fs.Directory
	GetPD() fs.Directory
	MarkFile()
}

// CreateModel creates a new model
func CreateModel() (Model, error) {
	m := model{}
	if err := m.start(); err != nil {
		return nil, err
	}
	return &m, nil
}

// ? Do we want to cache more directories than just those displayed
type model struct {
	// Working Directory
	wd *fs.Directory
	// Parent Directory
	pd *fs.Directory
	// Child directory
	cd *fs.Directory
}

func (m *model) start() error {
	logger.LogMessage(id, "Starting", logger.DEBUG)
	wd, err := os.Getwd()
	if err != nil {
		logger.LogError(id, "Failed to get initial wd", err)
		return err
	}
	m.wd = fs.GetDirectory(wd)
	m.pd = m.wd.GetParentDirectory()
	m.setCD()
	return nil
}
func (m model) logCurrentDirState() {
	logger.LogMessage(id, "cd: "+m.cd.GetPath(), logger.DEBUG)
	logger.LogMessage(id, "wd: "+m.wd.GetPath(), logger.DEBUG)
	logger.LogMessage(id, "pd: "+m.pd.GetPath(), logger.DEBUG)
}

// if the selection of wd is a diretory cd is set to that
// otherwise it's set to an empty directory
// todo otherwise it will be ignored in the future when we have previews working
func (m *model) setCD() {
	if m.wd.GetSelectedFile().GetFileInfo().IsDir() {
		m.cd = fs.GetDirectory(m.wd.GetPath() + "/" + m.wd.GetSelectedFile().GetFileInfo().Name())
	} else {
		m.cd = fs.GetEmptyDirectory()
	}
}

func (m *model) Navigate(d Direction) {
	switch d {
	case Left:
		if m.wd.CheckForParent() {
			m.cd = m.wd
			m.wd = m.pd
			m.pd = m.wd.GetParentDirectory()
		}
		m.logCurrentDirState()
	case Right:
		if !m.wd.GetSelectedFile().GetFileInfo().IsDir() {
			logger.LogMessage(id, "Can't navigate into file currently", logger.DEBUG)
			return
		}
		m.pd = m.wd
		m.wd = m.cd
		m.setCD()
		// todo if cd is a file show some kinda preview

		m.logCurrentDirState()
	case Up:
		// todo enable setting to disable circular selection thing
		m.wd.SetSelection((m.wd.GetSelection() - 1 + m.wd.GetFileCount()) % m.wd.GetFileCount())
		m.setCD()
	case Down:
		// todo enable setting to disable circular selection thing
		m.wd.SetSelection((m.wd.GetSelection() + 1) % m.wd.GetFileCount())
		m.setCD()
	case Top:
		m.wd.SetSelection(0)
		m.setCD()
	case Bottom:
		m.wd.SetSelection(m.wd.GetFileCount() - 1)
		m.setCD()
	}

}

func (m model) MarkFile() {
	logger.LogMessage(id, "Marked "+m.wd.GetSelectedFile().GetFileInfo().Name(), logger.DEBUG)
	m.wd.ToggleMarked()
	m.Navigate(Down)
}

func (m model) GetCD() fs.Directory {
	return *m.cd
}
func (m model) GetWD() fs.Directory {
	return *m.wd
}
func (m model) GetPD() fs.Directory {
	return *m.pd
}
