package model

import "github.com/alx99/fly/model/fs"

// DirState holds the state of all current directories
type DirState struct {
	// Working Directory
	wd fs.Directory
	// Parent Directory
	pd fs.Directory
	// Child directory
	cd fs.Directory
}

// GetPD returns the parent directory
func (d DirState) GetPD() fs.Directory {
	return d.pd
}

// GetWD returns the parent directory
func (d DirState) GetWD() fs.Directory {
	return d.wd
}

// GetCD returns the parent directory
func (d DirState) GetCD() fs.Directory {
	return d.cd
}
