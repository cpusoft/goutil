package netutil

import (
	"fmt"
	"testing"
)

func TestResolveIP(t *testing.T) {
	host := "baidu.com"
	fmt.Println("resolveIp:", ResolveIp(host))
}

func TestLookupIp(t *testing.T) {
	host := "baidu.com"
	fmt.Println("resolveIp:", LookupIp(host))
}
func TestLookupIpByUrl(t *testing.T) {
	url := "https://rrdp.lacnic.net/rrdp/26afc8dd-3e3b-4b31-8ff1-710d8e944320/22733/delta.xml"
	fmt.Println("resolveIpByUrl:", LookupIpByUrl(url))
}
