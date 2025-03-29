package models

import "strings"

func longestCommonPath(s1, s2 string) string {
	split1 := strings.Split(s1, "/")
	split2 := strings.Split(s2, "/")
	minLen := min(len(split2), len(split1))
	common := make([]string, 0, minLen)

	for i := range minLen {
		if split1[i] != split2[i] {
			break
		}
		common = append(common, split1[i])
	}

	return strings.Join(common, "/")
}
