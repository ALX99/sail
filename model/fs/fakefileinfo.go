package fs

import (
	"os"
	"time"
)

type fakefileinfo struct {
	name string
}

func (f fakefileinfo) Name() string {
	return f.name
}
func (f fakefileinfo) Size() int64 {
	return 0
}
func (f fakefileinfo) Mode() os.FileMode {
	return 555
}
func (f fakefileinfo) ModTime() time.Time {
	return time.Now()
}
func (f fakefileinfo) IsDir() bool {
	return false
}
func (f fakefileinfo) Sys() interface{} {
	return 0
}
