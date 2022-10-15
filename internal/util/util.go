package util

import "strings"

func Max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func GetParentPath(path string) string {
	p := strings.Split(path, "/")
	if len(p) == 1 {
		return "/"
	}
	return strings.Join(p[:len(p)-1], "/")
}
