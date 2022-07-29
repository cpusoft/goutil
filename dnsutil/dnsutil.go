package dnsutil

import (
	"errors"
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

func DomainBytesToStr(domainBytes []byte) (domainStr string, err error) {
	allLen := len(domainBytes)
	if allLen == 0 {
		return "", nil
	}
	domainStr = ""
	tmp := make([]byte, allLen)
	copy(tmp, domainBytes)
	for len(tmp) > 0 {
		oneLen := convert.ByteToBigInt(tmp[0])
		belogs.Debug("DomainBytesToStr():   tmp[0]:", tmp[0], " oneLen:", oneLen)
		if oneLen == nil {
			belogs.Error("DomainBytesToStr(): ByteToBigInt fail, domainBytes:", domainBytes, tmp[0])
			return "", errors.New("bytes cannot conver to domain")
		} else if int(oneLen.Int64()) > allLen {
			belogs.Error("DomainBytesToStr(): ByteToBigInt fail, oneLen is bigger than allLen, domainBytes:", domainBytes, tmp[0],
				"  oneLen:", oneLen.Int64(), " allLen:", allLen)
			return "", errors.New("bytes cannot conver to domain")
		}
		one := tmp[1 : oneLen.Int64()+1]
		domainStr += string(one)
		tmp = tmp[oneLen.Int64()+1:]
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
