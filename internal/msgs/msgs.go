package msgs

import "github.com/alx99/fly/internal/fs"

type MsgDirLoaded struct {
	To     uint32
	Path   string
	Select string
	Dir    fs.Directory
}
type MsgDirError struct {
	To   uint32
	Path string
	Err  error
}
type MsgDirReload struct {
}
