package dnsutil

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type ResourceRecordModel struct {
	DnsName     jsonutil.HexBytes `json:"dnsName"` // 2*n bytes --> n uint16
	DnsNameStr  string            `json:"dnsNameStr"`
	DnsType     uint16            `json:"dnsType"`  // 'type' is keyword in golang, so use dnsType
	DnsClass    uint16            `json:"dnsClass"` //
	DnsTtl      uint32            `json:"dnsTtl"`
	DnsRdLength uint16            `json:"dnsRdLength"`
	DnsRData    jsonutil.HexBytes `json:"dnsRData"`
}

func NewResourceRecordModelByDnsNameBytes(dnsName []byte, dnsType uint16, dnsClass uint16,
	dnsTtl uint32, dnsRdLength uint16, dnsRData []byte) *ResourceRecordModel {
	dnsNameStr, _ := DomainBytesToStr(dnsName)
	c := &ResourceRecordModel{
		DnsName:     dnsName,
		DnsNameStr:  dnsNameStr,
		DnsType:     dnsType,
		DnsClass:    dnsClass,
		DnsTtl:      dnsTtl,
		DnsRdLength: dnsRdLength,
		DnsRData:    dnsRData,
	}
	return c
}

func NewResourceRecordModelByDnsNameStr(dnsNameStr string, dnsType uint16, dnsClass uint16,
	dnsTtl uint32, dnsRdLength uint16, dnsRData []byte) *ResourceRecordModel {
	dnsName, _ := DomainStrToBytes(dnsNameStr)
	c := &ResourceRecordModel{
		DnsName:     dnsName,
		DnsNameStr:  dnsNameStr,
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

func ParseToResourceRecordModels(buf *bytes.Reader) (resourceRecordModels []*ResourceRecordModel, err error) {
	if buf.Len() == 0 {
		return nil, errors.New("buf is empty")
	}
	resourceRecordModels = make([]*ResourceRecordModel, 0)
	for {
		// get dnsName:end with 0x00
		dnsName := make([]byte, 0)
		for buf.Len() > 0 {
			s := make([]byte, 1)
			err = binary.Read(buf, binary.BigEndian, &s)
			if err != nil {
				belogs.Error("ParseToResourceRecordModel():get dnsName, read one byte fail, will ignore: ", buf)
				return nil, err
			}
			belogs.Debug("ParseToResourceRecordModel():compare s is 0x00:", s)
			if bytes.Equal(s, []byte{0x00}) {
				break
			}
			dnsName = append(dnsName, s...)
		}
		belogs.Debug("ParseToResourceRecordModel():  dnsName:", convert.PrintBytesOneLine(dnsName))

		// 0x00 has used in dnsname, so just uint8
		var dnsType uint16
		err = binary.Read(buf, binary.BigEndian, &dnsType)
		if err != nil {
			belogs.Error("ParseToResourceRecordModel():get dnsType fail, will ignore: ", buf)
			return nil, err
		}
		belogs.Debug("ParseToResourceRecordModel(): dnsType:", dnsType)

		var dnsClass uint16
		err = binary.Read(buf, binary.BigEndian, &dnsClass)
		if err != nil {
			belogs.Error("ParseToResourceRecordModel():get dnsClass fail, will ignore: ", buf)
			return nil, err
		}
		belogs.Debug("ParseToResourceRecordModel(): dnsClass:", dnsClass)

		var dnsTtl uint32
		err = binary.Read(buf, binary.BigEndian, &dnsTtl)
		if err != nil {
			belogs.Error("ParseToResourceRecordModel():get dnsTtl fail, will ignore: ", buf)
			return nil, err
		}
		belogs.Debug("ParseToResourceRecordModel(): dnsTtl:", dnsTtl)

		var dnsRdLen uint16
		err = binary.Read(buf, binary.BigEndian, &dnsRdLen)
		if err != nil {
			belogs.Error("ParseToResourceRecordModel():get dnsRdLen fail, will ignore: ", buf)
			return nil, err
		}
		belogs.Debug("ParseToResourceRecordModel(): dnsRdLen:", dnsRdLen)

		// when remove collective rr , dsnRdLen == 0
		var rr *ResourceRecordModel
		if dnsRdLen > 0 {
			dnsRData := make([]byte, dnsRdLen)
			err = binary.Read(buf, binary.BigEndian, &dnsRData)
			if err != nil {
				belogs.Error("ParseToResourceRecordModel():get dnsRData fail, will ignore: ", buf)
				return nil, err
			}
			belogs.Debug("ParseToResourceRecordModel():  dnsRData:", convert.PrintBytesOneLine(dnsRData))
			rr = NewResourceRecordModelByDnsNameBytes(dnsName, dnsType, dnsClass, dnsTtl, dnsRdLen, dnsRData)
		} else {
			rr = NewResourceRecordModelByDnsNameBytes(dnsName, dnsType, dnsClass, dnsTtl, 0, nil)
		}
		belogs.Debug("ParseToResourceRecordModel(): rr:", jsonutil.MarshalJson(rr))
		resourceRecordModels = append(resourceRecordModels, rr)
		if buf.Len() == 0 {
			break
		}
	}
	belogs.Debug("ParseToResourceRecordModel(): resourceRecordModels:", jsonutil.MarshalJson(resourceRecordModels))
	return resourceRecordModels, nil
}
