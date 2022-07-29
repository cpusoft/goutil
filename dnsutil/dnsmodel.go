package dnsutil

import (
	"bytes"
	"encoding/binary"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

type ResourceRecordModel struct {
	DnsName     jsonutil.HexBytes `json:"dnsName"`  // 2*n bytes --> n uint16
	DnsType     uint16            `json:"dnsType"`  // 'type' is keyword in golang, so use dnsType
	DnsClass    uint16            `json:"dnsClass"` //
	DnsTtl      uint32            `json:"dnsTtl"`
	DnsRdLength uint16            `json:"dnsRdLength"`
	DnsRData    jsonutil.HexBytes `json:"dnsRData"`
}

func NewResourceRecordModel(dnsName []byte, dnsType uint16, dnsClass uint16,
	dnsTtl uint32, dnsRdLength uint16, dnsRData []byte) *ResourceRecordModel {
	c := &ResourceRecordModel{
		DnsName:     dnsName,
		DnsType:     dnsType,
		DnsClass:    dnsClass,
		DnsTtl:      dnsTtl,
		DnsRdLength: dnsRdLength,
		DnsRData:    dnsRData,
	}
	return c
}

func (c *ResourceRecordModel) Length() uint16 {
	// type(2)+class(2)+ttl(4)+rdlen(2)
	return uint16(len(c.DnsName) + 2 + 2 + 4 + 2 + len(c.DnsRData))
}
func (c *ResourceRecordModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.DnsName)
	binary.Write(wr, binary.BigEndian, c.DnsType)
	binary.Write(wr, binary.BigEndian, c.DnsClass)
	binary.Write(wr, binary.BigEndian, c.DnsTtl)
	binary.Write(wr, binary.BigEndian, c.DnsRdLength)
	if len(c.DnsRData) > 0 {
		binary.Write(wr, binary.BigEndian, c.DnsRData)
	}
	return wr.Bytes()
}

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
