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
	v := GetValueInWhoisResult(r, "country", "aut-num")
	fmt.Println("country:", v)

	v = GetValueInWhoisResult(r, "source", "aut-num")
	fmt.Println("source:", v)

	v = GetValueInWhoisResult(r, "as-name", "aut-num")
	fmt.Println("as-name:", v)
}
func TestWhiosCymru(t *testing.T) {
	host := `whois.cymru.com`
	q := `AS266087`
	whoisConfig := &WhoisConfig{
		Host: host,
		Port: "43",
	}
	r, e := GetWhoisResultWithConfig(q, whoisConfig)
	fmt.Println(jsonutil.MarshalJson(r), e)

}
