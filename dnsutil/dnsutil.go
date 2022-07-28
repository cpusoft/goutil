package dnsutil

import (
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

	return
}
