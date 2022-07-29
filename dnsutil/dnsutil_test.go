package dnsutil

import (
	"fmt"
	"testing"
)

func TestDomainStrToBytes(t *testing.T) {
	d := `dwn.roo.bo`
	b, err := DomainStrToBytes(d)
	fmt.Println(b, err)

	dd, err := DomainBytesToStr(b)
	fmt.Println(dd, err)
}
