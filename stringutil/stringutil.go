package stringutil

import ()

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
