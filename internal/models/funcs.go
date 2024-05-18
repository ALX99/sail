package models

import "os"

var (
	removeFile = os.Remove
	readDir    = os.ReadDir
)
