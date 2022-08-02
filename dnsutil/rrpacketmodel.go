package dnsutil

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

// for packet
type RrPacketModel struct {

	//  domain = name + origin (will remove the end ".")  //
	// NameStr: dwn.roo.bo --> Name: 03 64 77 6e 03 72 6f 6f 02 62 6f 00
	RrPacketDomain jsonutil.HexBytes `json:"rrPacketDomain"`
	RrPacketType   uint16            `json:"rrPacketType"`  // 'type' is keyword in golang, so use dnsType
	RrPacketClass  uint16            `json:"rrPacketClass"` //
	RrPacketTtl    uint32            `json:"rrPacketTtl"`

	// == len(RrPacketRData)
	RrPacketDataLength uint16            `json:"rrPacketDataLength"`
	RrPacketData       jsonutil.HexBytes `json:"rrPacketData"`
}

func NewRrPacketModel(rrPacketDomain []byte, rrPacketType uint16, rrPacketClass uint16, rrPacketTtl uint32,
	rrPacketDataLength uint16, rrPacketData []byte) (rrPacketModel *RrPacketModel) {
	c := &RrPacketModel{
		RrPacketDomain:     rrPacketDomain,
		RrPacketType:       rrPacketType,
		RrPacketClass:      rrPacketClass,
		RrPacketTtl:        rrPacketTtl,
		RrPacketDataLength: rrPacketDataLength,
		RrPacketData:       rrPacketData,
	}
	return c
}
func ConvertRrToPacketModel(originModel OriginModel, rrModel RrModel) (rrPacketModel *RrPacketModel, err error) {
	belogs.Debug("ConvertRrModelToRrPacketModel(): originModel:", originModel, "  rrModel:", jsonutil.MarshalJson(rrModel))

	domain := rrModel.RrName + "." + originModel.Origin
	rrPacketDomain, err := DomainStrToBytes(domain)
	if err != nil {
		belogs.Error("ConvertRrToPacketModel(): DomainStrToBytes fail, RrName:", rrModel.RrName, "  origin:", originModel.Origin, err)
		return nil, err
	}

	rrPacketType, ok := DnsStrTypes[rrModel.RrType]
	if !ok {
		belogs.Error("ConvertRrModelToRrPacketModel(): DnsStrTypes fail, RrType:", rrModel.RrType)
		return nil, errors.New("RrType is illegal")
	}
	rrPacketClass, ok := DnsStrTypes[rrModel.RrClass]
	if !ok {
		belogs.Error("ConvertRrModelToRrPacketModel(): DnsStrTypes fail, RrClass:", rrModel.RrClass)
		return nil, errors.New("RrClass is illegal")
	}
	var rrPacketTtl uint32
	if !rrModel.RrTtl.IsZero() {
		rrPacketTtl = uint32(rrModel.RrTtl.ValueOrZero())
	} else {
		if !originModel.Ttl.IsZero() {
			rrPacketTtl = uint32(originModel.Ttl.ValueOrZero())
		}
	}
	rrPacketData := []byte(rrModel.RrData)
	c := NewRrPacketModel(rrPacketDomain, rrPacketType, rrPacketClass, rrPacketTtl,
		uint16(len(rrPacketData)), rrPacketData)
	belogs.Debug("ConvertRrModelToRrPacketModel(): rrPacketModel:", c)
	return c, nil
}

func (c *RrPacketModel) Length() uint16 {
	//   type(2)+class(2)+ttl(4)+rdlen(2)
	return uint16(len(c.RrPacketDomain) + 2 + 2 + 4 + 2 + len(c.RrPacketData))
}
func (c *RrPacketModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.RrPacketDomain)
	binary.Write(wr, binary.BigEndian, c.RrPacketType)
	binary.Write(wr, binary.BigEndian, c.RrPacketClass)
	binary.Write(wr, binary.BigEndian, c.RrPacketTtl)
	binary.Write(wr, binary.BigEndian, c.RrPacketDataLength)
	if len(c.RrPacketData) > 0 {
		binary.Write(wr, binary.BigEndian, c.RrPacketData)
	}
	return wr.Bytes()
}

// rrAllLen:>0,will compare; if <=0, no use
func ParseToRrPacketModels(buf *bytes.Reader, rrAllLen int) (rrPacketModels []*RrPacketModel, err error) {
	if buf.Len() == 0 {
		return nil, errors.New("buf is empty")
	}
	rrPacketModels = make([]*RrPacketModel, 0)
	rrSumLen := 0
	for {
		// get dnsName:end with 0x00
		rrPacketDomain := make([]byte, 0)
		for buf.Len() > 0 {
			s := make([]byte, 1)
			err = binary.Read(buf, binary.BigEndian, &s)
			if err != nil {
				belogs.Error("ParseToRrPacketModels():get RrPacketDomain, read one byte fail, will ignore: ", buf)
				return nil, err
			}
			belogs.Debug("ParseToRrPacketModels():compare s is 0x00:", s)
			if bytes.Equal(s, []byte{0x00}) {
				break
			}
			rrPacketDomain = append(rrPacketDomain, s...)
		}
		belogs.Debug("ParseToRrPacketModels():  dnsName:", convert.PrintBytesOneLine(rrPacketDomain))

		//
		var rrPacketType uint16
		err = binary.Read(buf, binary.BigEndian, &rrPacketType)
		if err != nil {
			belogs.Error("ParseToRrPacketModels():get rrPacketType fail, will ignore: ", buf)
			return nil, err
		}
		belogs.Debug("ParseToRrPacketModels(): rrPacketType:", rrPacketType)

		var rrPacketClass uint16
		err = binary.Read(buf, binary.BigEndian, &rrPacketClass)
		if err != nil {
			belogs.Error("ParseToRrPacketModels():get rrPacketClass fail, will ignore: ", buf)
			return nil, err
		}
		belogs.Debug("ParseToRrPacketModels(): rrPacketClass:", rrPacketClass)

		var rrPacketTtl uint32
		err = binary.Read(buf, binary.BigEndian, &rrPacketTtl)
		if err != nil {
			belogs.Error("ParseToRrPacketModels():get rrPacketTtl fail, will ignore: ", buf)
			return nil, err
		}
		belogs.Debug("ParseToRrPacketModels(): rrPacketTtl:", rrPacketTtl)

		var rrPacketDataLength uint16
		err = binary.Read(buf, binary.BigEndian, &rrPacketDataLength)
		if err != nil {
			belogs.Error("ParseToRrPacketModels():get rrPacketDataLength fail, will ignore: ", buf)
			return nil, err
		}
		belogs.Debug("ParseToRrPacketModels(): rrPacketDataLength:", rrPacketDataLength)

		// when remove collective rr , dsnRdLen == 0
		var rrPacketModel *RrPacketModel
		if rrPacketDataLength > 0 {
			rrPacketData := make([]byte, rrPacketDataLength)
			err = binary.Read(buf, binary.BigEndian, &rrPacketData)
			if err != nil {
				belogs.Error("ParseToRrPacketModels():get RrPacketModel fail, will ignore: ", buf)
				return nil, err
			}
			belogs.Debug("ParseToRrPacketModels():  rrPacketData:", convert.PrintBytesOneLine(rrPacketData))
			rrPacketModel = NewRrPacketModel(rrPacketDomain, rrPacketType, rrPacketClass, rrPacketTtl, rrPacketDataLength, rrPacketData)
		} else {
			rrPacketModel = NewRrPacketModel(rrPacketDomain, rrPacketType, rrPacketClass, rrPacketTtl, 0, nil)
		}
		belogs.Debug("ParseToRrPacketModels(): rr:", jsonutil.MarshalJson(rrPacketModel))
		rrPacketModels = append(rrPacketModels, rrPacketModel)

		if buf.Len() == 0 {
			break
		}
		rrSumLen += int(rrPacketModel.Length())
		belogs.Debug("ParseToRrPacketModels(): rrPacketModel.Length():", rrPacketModel.Length(), " rrSumLen:", rrSumLen)
		if rrAllLen > 0 {
			if rrSumLen >= rrAllLen {
				break
			}
		}
	}
	belogs.Debug("ParseToRrPacketModels(): rrPacketModels:", jsonutil.MarshalJson(rrPacketModels))
	return rrPacketModels, nil
}
