package stringutil

import (
	"bytes"
	"strings"

	"github.com/cpusoft/goutil/convert"
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

func TrimNewLine(str string) (s string) {
	s = strings.Replace(str, "\r", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	return s
}

func TrimSpace(str string) (s string) {
	s = strings.Replace(str, "\t", "", -1)
	s = strings.Replace(s, " ", "", -1)
	return s
}

func TrimSpaceAndNewLine(str string) (s string) {
	s = TrimSpace(str)
	return TrimNewLine(s)
}

func TrimeSuffixAll(str, trim string) (s string) {

	s = strings.TrimSuffix(str, trim)
	if strings.HasSuffix(str, trim) {
		return TrimeSuffixAll(s, trim)
	}
	return s

}

// line: a=***&b=***&c=***
// key: a
// separator: &
func GetValueFromJointStr(line, key, separator string) string {
	split := strings.Split(line, separator)
	if len(split) == 0 {
		return ""
	}
	for i := range split {
		if strings.HasPrefix(split[i], key+"=") {
			return strings.Replace(split[i], key+"=", "", -1)
		}
	}
	return ""
}

// Ommitting too long string
func OmitString(str string, end uint64) string {
	len := uint64(len(str))
	if len == 0 {
		return ""
	}
	if end > len {
		end = len
	}
	return str[:int(end)]
}

func Int64sToInString(s []int64) string {
	if len(s) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteString("(")
	for i := 0; i < len(s); i++ {
		buffer.WriteString(convert.ToString(s[i]))
		if i < len(s)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}

func StringsToInString(s []string) string {
	if len(s) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteString("(")
	for i := 0; i < len(s); i++ {
		buffer.WriteString("\"" + s[i] + "\"")
		if i < len(s)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}
