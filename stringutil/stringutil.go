package stringutil

import (
	"strings"
)

func ContainInSlice(slice []string, one string) bool {
	if len(slice) == 0 || len(one) == 0 {
		return false
	}
	for _, s := range slice {
		if s == one {
			return true
		}
	}
	return false
}

func TrimSpaceAneNewLine(str string) (s string) {
	s = strings.TrimSpace(str)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	return s
}
