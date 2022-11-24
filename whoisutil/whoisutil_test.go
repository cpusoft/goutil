package whoisutil

import (
	"fmt"
	"testing"
)

func TestGetWhoisResult(t testing.T) {
	q := "baidu.com"
	r, e := GetWhoisResult(q)
	fmt.Println(r, e)

	q = "8.8.8.8/24"
	r, e = GetWhoisResult(q)
	fmt.Println(r, e)
}
