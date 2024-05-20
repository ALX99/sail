package models

import "strings"

func longestCommonPath(s1, s2 string) string {
	split1 := strings.Split(s1, "/")
	split2 := strings.Split(s2, "/")
	minLen := len(split1)

	if len(split2) < minLen {
		minLen = len(split2)
	}

	common := make([]string, 0, minLen)

	for i := 0; i < minLen; i++ {
		if split1[i] != split2[i] {
			break
		}
		common = append(common, split1[i])
	}

	return strings.Join(common, "/")
}
