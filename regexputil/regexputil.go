package regexputil

import (
	"regexp"
)

func IsHex(s string) (bool, error) {
	return regexp.MatchString(`^[0-9a-fA-F]+$`, s)
}
