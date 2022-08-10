package dnsutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
)

//Name: dwn.roo.bo --> 03 64 77 6e 03 72 6f 6f 02 62 6f 00
func DomainStrToBytes(domainStr string) (domainBytes []byte, err error) {
	belogs.Debug("DomainStrToBytes(): domainStr:", domainStr)
	if len(domainStr) == 0 {
		return make([]byte, 0), nil
	}

	split := strings.Split(domainStr, ".")
	domainBytes = make([]byte, 0, len(domainStr)+10)
	for i := range split {
		if len(split[i]) == 0 {
			continue
		}
		lenByte, err := convert.IntToBytes(int8(len(split[i])))
		if err != nil {
			belogs.Error("DomainStrToBytes(): IntToBytes fail, domainStr:", domainStr, split[i])
			return nil, err
		}
		domainBytes = append(domainBytes, lenByte...)
		d := []byte(split[i])
		domainBytes = append(domainBytes, d...)
	}
	// ends with 0x00
	domainBytes = append(domainBytes, byte(0))
	belogs.Debug("DomainStrToBytes(): domainStr:", domainStr, "  domainBytes:", convert.PrintBytesOneLine(domainBytes))
	return domainBytes, nil
}

// not support compression pointer
func DomainBytesToStr(domainBytes []byte) (domainStr string, err error) {
	allLen := len(domainBytes)
	if allLen == 0 {
		return "", nil
	}
	domainStr = ""
	tmp := make([]byte, allLen)
	copy(tmp, domainBytes)
	for len(tmp) > 0 {
		oneLenBig := convert.ByteToBigInt(tmp[0])
		belogs.Debug("DomainBytesToStr():   tmp[0]:", tmp[0], " oneLenBig:", oneLenBig)
		if oneLenBig == nil {
			belogs.Error("DomainBytesToStr(): ByteToBigInt fail, domainBytes:", domainBytes, tmp[0])
			return "", errors.New("bytes cannot conver to domain")
		}

		oneLen := uint16(oneLenBig.Int64())
		if int(oneLen) > allLen {
			belogs.Error("DomainBytesToStr(): ByteToBigInt fail, oneLen is bigger than allLen, domainBytes:", domainBytes, tmp[0],
				"  oneLen:", oneLen, " allLen:", allLen)
			return "", errors.New("bytes cannot conver to domain")
		}
		is, pointer := IsDomainCompressionPointer(oneLen) //1100 0000
		if is {
			belogs.Error("DomainBytesToStr(): IsDomainCompressionPointer is, pointer:", is, fmt.Sprintf("%0x", pointer))
			return "", errors.New("not support compression pointer now")
		}
		one := tmp[1 : oneLen+1]
		domainStr += string(one)
		tmp = tmp[oneLen+1:]
		belogs.Debug("DomainBytesToStr(): one:", one, "  domainStr:", domainStr, "   tmp:", tmp)
		if len(tmp) > 1 {
			domainStr += "."
		} else if len(tmp) == 0 {
			// end
			break
		} else if len(tmp) == 1 && tmp[0] == 0x00 {
			// end with 0x00
			break
		}
		belogs.Debug("DomainBytesToStr():   tmp[0]:", tmp[0], " oneLen:", oneLen, " domainStr:", domainStr)
	}
	belogs.Debug("DomainBytesToStr():   domainBytes:", convert.PrintBytesOneLine(domainBytes), "   domainStr:", domainStr)
	return domainStr, nil
}

func IsDomainCompressionPointer(oneLen uint16) (is bool, pointerOffset uint16) {
	is = false
	if (oneLen & DNS_DOMAIN_COMPRESSION_POINTER) == DNS_DOMAIN_COMPRESSION_POINTER {
		is = true
		pointerOffset = oneLen & (^DNS_DOMAIN_COMPRESSION_POINTER)
		return is, pointerOffset
	}
	return false, 0
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
