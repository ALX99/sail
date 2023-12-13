package directory

import "github.com/alx99/fly/internal/fs"

type msgDirLoaded struct {
	path         string
	onLoadSelect string
	dir          fs.Directory
	role         Role
}

type msgDirError struct {
	err  error
	path string
	role Role
}

type msgFileSelected struct {
	path     string
	selected string
	role     Role
}
