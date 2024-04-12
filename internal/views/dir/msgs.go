package dir

import "github.com/alx99/fly/internal/fs"

type msgDirLoaded struct {
	onLoadSelect string
	dir          fs.Directory
	role         Role
}

type msgDirError struct {
	err  error
	role Role
}

type msgFileSelected struct {
	path     string
	selected string
	role     Role
}
