package dnsutil

import (
	"fmt"
	"testing"
)

func TestIsDomainCompressionPointer(t *testing.T) {
	oneLen := []byte{0xc0, 0x11}
	fmt.Printf("%0x\r\n", oneLen)

	isCompression, pointer, labelLength, label, newOffsetFromStart, err := CheckDomainCompressionPointer(oneLen, 10)
	fmt.Printf("%v,%v,%v,%0x,%v,%v\r\n", isCompression, pointer, labelLength, label, newOffsetFromStart, err)

	oneLen = []byte{0x05, 0x5f, 0x68, 0x74, 0x74, 0x70, 0x04, 0x5f, 0x74, 0x63, 0x70, 0x06, 0x64, 0x6e, 0x73, 0x2d, 0x73, 0x64, 0x03, 0x6f, 0x72, 0x67, 0x00}
	isCompression, pointer, labelLength, label, newOffsetFromStart, err = CheckDomainCompressionPointer(oneLen, 10)
	fmt.Printf("%v,%v,%v,%v,%v,%v\r\n", isCompression, pointer, labelLength, string(label), newOffsetFromStart, err)
}
func TestDomainStrToBytes(t *testing.T) {
	d := `dwn.roo.bo`
	b, err := DomainStrToBytes(d)
	fmt.Println(b, err)

	dd, err := DomainBytesToStr(b)
	fmt.Println(dd, err)
}
