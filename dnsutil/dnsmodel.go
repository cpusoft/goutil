package dnsutil

import "github.com/cpusoft/goutil/belogs"

type QrOpCodeZRCode uint16

func ComposeQrOpCodeZRCode(qr uint8, opCode uint8, rCode uint8) (qrOpCodeZRCode QrOpCodeZRCode) {
	q := uint16(0)
	if qr == DNS_QR_RESPONSE {
		q = (DNS_QR_RESPONSE << 15)
	}
	q |= uint16(opCode) << uint16(11)
	q |= uint16(rCode)
	qrOpCodeZRCode = QrOpCodeZRCode(q)
	belogs.Debug("ComposeQrOpCodeZRCode():qr:", qr, " opCode:", opCode, "  rCode:", rCode, "  qrOpCodeZRCode:", qrOpCodeZRCode)
	return qrOpCodeZRCode
}

func ParseQrOpCodeZRCode(qrOpCodeZRCode QrOpCodeZRCode) (qr uint8, opCode uint8, z uint8, rCode uint8) {
	qr = uint8((qrOpCodeZRCode >> 15) & 1)        // 1000 0000 0000 0000 --> 1
	opCode = uint8((qrOpCodeZRCode >> 11) & 0x0f) // 0111 1000 0000 0000 --> 1111
	z = uint8((qrOpCodeZRCode >> 4) & 0x7f)       // 0000 0111 1111 0000 --> 111 1111
	rCode = uint8(qrOpCodeZRCode & 0x0f)          // 0000 0000 0000 1111 --> 1111
	belogs.Debug("ParseQrOpCodeZRCode():qrOpCodeZRCode:", qrOpCodeZRCode, "  qr:", qr, " opCode:", opCode,
		"  z:", z, "  rCode:", rCode)
	return qr, opCode, z, rCode
}
