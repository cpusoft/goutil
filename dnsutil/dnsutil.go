package dnsutil

import (
	"encoding/binary"
	"errors"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
)

// host to domain:
// aaa.test.cn --> test.cn : rm first label
// test.cn --> test.cn: no change
func DomainTrimFirstLabel(domain string) string {
	splits := strings.Split(domain, ".")
	if len(splits) > 2 {
		return strings.TrimPrefix(domain, splits[0]+".")
	} else {
		return domain
	}
}

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
// Name:  03 64 77 6e 03 72 6f 6f 02 62 6f 00 --> dwn.roo.bo
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

// oneLenBytes:[2]byte
// if isCompression is true, then use offset
// if isCompression is false, then use labelLegth and label
func CheckDomainCompressionPointer(bytess []byte, offsetFromStart uint16) (isCompression bool, pointer uint16,
	labelLength uint8, label []byte, newOffsetFromStart uint16, err error) {
	if len(bytess) < 2 {
		belogs.Error("CheckDomainCompressionPointer():len(oneLenBytes) < 2 ,fail:",
			convert.PrintBytesOneLine(bytess))
		return false, 0, 0, nil, 0, errors.New("bytes is too small")
	}
	offset := binary.BigEndian.Uint16(bytess[:2])
	belogs.Debug("CheckDomainCompressionPointer(): bytess:", convert.PrintBytesOneLine(bytess),
		"  offset:", offset)

	// if is compression
	if (offset & DNS_DOMAIN_COMPRESSION_POINTER) == DNS_DOMAIN_COMPRESSION_POINTER {
		pointer = offset & (^DNS_DOMAIN_COMPRESSION_POINTER)
		return true, pointer, 0, nil, offsetFromStart + 2, nil
	}

	// if is not compression, len is just byte[0]
	labelLength = uint8(bytess[0])
	if labelLength <= DNS_DOMAIN_ONE_LABEL_MAXLENGTH &&
		len(bytess) > int(labelLength) {
		label = make([]byte, labelLength)
		copy(label, bytess[1:labelLength+1])
		return false, 0, labelLength, label, offsetFromStart + 1 + uint16(labelLength), nil
	} else {
		return false, 0, 0, nil, 0, errors.New("One label in domain should be 63 octets or less.")
	}
}
