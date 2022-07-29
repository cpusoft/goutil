package dnsutil

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestRr(t *testing.T) {
	dnsNameStr1 := `dwn1.roo.bo`
	dnsRData1 := []byte("1.1.1.1")
	rr1 := NewResourceRecordModelByDnsNameStr(dnsNameStr1, uint16(DNS_TYPE_A), uint16(DNS_CLASS_IN),
		uint32(100), uint16(len(dnsRData1)), dnsRData1)
	fmt.Println(jsonutil.MarshalJson(rr1))
	b1 := rr1.Bytes()
	fmt.Println(convert.PrintBytesOneLine(b1))

	dnsNameStr2 := `dwn2.roo.bo`
	dnsRData2 := []byte("2.2.2.2")
	rr2 := NewResourceRecordModelByDnsNameStr(dnsNameStr2, uint16(DNS_TYPE_A), uint16(DNS_CLASS_IN),
		uint32(200), uint16(len(dnsRData2)), dnsRData2)
	fmt.Println(jsonutil.MarshalJson(rr2))
	b2 := rr2.Bytes()
	fmt.Println(convert.PrintBytesOneLine(b2))

	dnsNameStr3 := `dwn3.roo.bo`
	dnsRData3 := []byte("3.3.3.")
	rr3 := NewResourceRecordModelByDnsNameStr(dnsNameStr3, uint16(DNS_TYPE_A), uint16(DNS_CLASS_IN),
		uint32(300), uint16(len(dnsRData3)), dnsRData3)
	fmt.Println(jsonutil.MarshalJson(rr3))
	b3 := rr3.Bytes()
	fmt.Println(convert.PrintBytesOneLine(b3))

	allRr := make([]*ResourceRecordModel, 0)
	allRr = append(allRr, rr1)
	allRr = append(allRr, rr2)
	allRr = append(allRr, rr3)
	fmt.Println(jsonutil.MarshalJson(allRr))

	allByte := make([]byte, 0)
	allByte = append(allByte, b1...)
	allByte = append(allByte, b2...)
	allByte = append(allByte, b3...)

	buf := bytes.NewReader(allByte)
	rrs, err := ParseToResourceRecordModels(buf)
	fmt.Println(jsonutil.MarshalJson(rrs), err)
}
