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

	whoisConfig := &WhoisConfig{
		Host: "whois.apnic.net",
	}
	q = "AS45090"
	r, e = GetWhoisResultWithConfig(q, whoisConfig)
	fmt.Println(jsonutil.MarshalJson(r), e)
	v := GetValueInWhoisResult(r, "country")
	fmt.Println("country:", v)

}
