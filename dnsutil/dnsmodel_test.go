package dnsutil

import (
	"fmt"
	"testing"
)

func TestQrOpCodeZRCode(t *testing.T) {
	qr := uint8(DNS_QR_RESPONSE)
	opCode := uint8(DNS_OPCODE_DSO)
	rCode := uint8(DNS_RCODE_DSOTYPENI)
	qrOpCodeZRCode := ComposeQrOpCodeZRCode(qr, opCode, rCode)
	fmt.Println(qrOpCodeZRCode)

	qr1, opCode1, rCode1 := ParseQrOpCodeZRCode(qrOpCodeZRCode)
	fmt.Println(qr1, opCode1, rCode1)
}
