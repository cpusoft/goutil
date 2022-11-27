package whoisutil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestGetWhoisResult(t *testing.T) {
	q := "baidu.com"
	r, e := GetWhoisResult(q)
	fmt.Println(jsonutil.MarshalJson(r), e)

	q = "8.8.8.8"
	r, e = GetWhoisResult(q)
	fmt.Println(jsonutil.MarshalJson(r), e)
}
