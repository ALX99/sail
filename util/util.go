package util

// Contains checks if the string e is present in the string array a
func Contains(a []string, e string) bool {
	for _, a := range a {
		if a == e {
			return true
		}
	}
	return false
}

// Min gets the minimum of two numbers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max gets the maximum of two numbers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
