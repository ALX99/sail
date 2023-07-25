package util

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

// Contains checks if the string e is present in the string array a
func Contains(a []string, e string) bool {
	for _, a := range a {
		if a == e {
			return true
		}
	}
	return false
}
