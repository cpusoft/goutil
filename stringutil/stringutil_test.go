package stringutil

import (
	"fmt"
	"testing"
)

func TestTrimeSuffixAll(t *testing.T) {
	ips := []string{"16.70.0.0", "16.0.1.0"}

	for _, ip := range ips {
		str := TrimeSuffixAll(ip, ".0")
		fmt.Println(ip, " --> ", str)

	}
}

func TestGetValueFromJointStr(t *testing.T) {
	str := `a=111&b=222&c=333`
	v := GetValueFromJointStr(str, "a", "&")
	fmt.Println(v)
	v = GetValueFromJointStr(str, "b", "&")
	fmt.Println(v)
	v = GetValueFromJointStr(str, "c", "&")
	fmt.Println(v)
}
