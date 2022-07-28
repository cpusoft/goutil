package dnsutil

import (
	"fmt"
	"testing"
)

func TestDomainStrToBytes(t *testing.T) {
	d := `dwn.roo.bo.`
	b, err := DomainStrToBytes(d)
	fmt.Println(b, err)
}
