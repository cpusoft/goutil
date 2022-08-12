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

func TestQrOpCodeZRCode(t *testing.T) {
	qr := uint8(DNS_QR_RESPONSE)
	opCode := uint8(DNS_OPCODE_DSO)
	rCode := uint8(DNS_RCODE_DSOTYPENI)
	qrOpCodeZRCode := ComposeQrOpCodeZRCode(qr, opCode, rCode)
	fmt.Println(qrOpCodeZRCode)

	qr1, opCode1, z, rCode1 := ParseQrOpCodeZRCode(qrOpCodeZRCode)
	fmt.Println(qr1, opCode1, z, rCode1)
}

func TestIsDomainCompressionPointer(t *testing.T) {
	oneLen := []byte{0xc0, 0x11}
	fmt.Printf("%0x\r\n", oneLen)

	is, offset, err := CheckDomainCompressionPointer(oneLen)
	fmt.Printf("%v,%0x,%v\r\n", is, offset, err)

}
