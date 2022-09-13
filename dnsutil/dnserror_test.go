package dnsutil

import (
	"fmt"
	"testing"
)

func TestDnsError(t *testing.T) {
	fmt.Println("TestDnsError")
	e := NewDnsError("just error", 0, 0, 0, 0)
	esss, ok := e.(dnsError)
	fmt.Println(ok, esss)
}
