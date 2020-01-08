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
