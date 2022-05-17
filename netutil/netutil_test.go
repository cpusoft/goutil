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
