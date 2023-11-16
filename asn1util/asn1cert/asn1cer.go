package asn1cert

import (
	"encoding/asn1"
	"errors"
	"net"

	"github.com/cpusoft/goutil/asn1util/asn1base"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/guregu/null"
)

// addressPrefix or min/max
type IpAddrBlock struct {
	AddressFamily uint64 `json:"addressFamily" asn1:"optional"`
	//address prefix: 147.28.83.0/24 '
	AddressPrefix string `json:"addressPrefix"`
	//min address:  99.96.0.0
	Min string `json:"min"`
	//max address:   99.105.127.255
	Max string `json:"max"`
}
type IpAddrRaw struct {
	AddressFamily   []byte
	IpAddressChoice asn1.RawValue
}
type IpAddrRange struct {
	Min asn1.BitString
	Max asn1.BitString
}

// ipv4: size==4, ipv6: size==6
func ParseBitStringToIpNet(bi asn1.BitString, ipType int) (ipNet *net.IPNet, err error) {
	var size int
	if ipType == 1 {
		size = 4
	} else if ipType == 2 {
		size = 16
	} else {
		belogs.Error("ParseBitStringToAddressPrefix(): ipType fail:", ipType)
		return nil, errors.New("Not an IP type")
	}

	ipAddr := make([]byte, size)
	copy(ipAddr, bi.Bytes)
	mask := net.CIDRMask(bi.BitLength, size*8)
	belogs.Debug("ParseBitStringToAddressPrefix(): ipAddr:", convert.PrintBytesOneLine(ipAddr),
		jsonutil.MarshalJson(ipAddr), "  mask:", mask)
	return &net.IPNet{
		IP:   net.IP(ipAddr),
		Mask: mask,
	}, nil
}

// data: just ip data, no asn1 header
func ParseBytesToIpNet(data []byte, ipType int) (*net.IPNet, error) {
	belogs.Debug("ParseBytesToIpNet(): ipType:", ipType, "   len(data):", len(data))

	bi, err := asn1base.ParseBitString(data)
	if err != nil {
		belogs.Error("ParseBytesToIpNet(): ParseBitString fail:", convert.PrintBytesOneLine(data))
		return nil, errors.New("data is not IP address")
	}
	bitString := asn1.BitString{
		Bytes:     bi.Bytes,
		BitLength: bi.BitLength,
	}
	return ParseBitStringToIpNet(bitString, ipType)
}

// use ParseBytesToIpNet --> 134.144.0.0/16
func ParseBitStringToAddressPrefix(bi asn1.BitString, ipType int) (addressPrefix string, err error) {
	net, err := ParseBitStringToIpNet(bi, ipType)
	if err != nil {
		belogs.Error("ParseBitStringToAddressPrefix(): ParseBitStringToIpNet fail:", err)
		return "", errors.New("data is not IP address")
	}
	return net.String(), nil
}

// use ParseBytesToIpNet --> 134.144.0.0/16
func ParseBytesToAddressPrefix(data []byte, ipType int) (addressPrefix string, err error) {
	net, err := ParseBytesToIpNet(data, ipType)
	if err != nil {
		belogs.Error("ParseBytesToAddressPrefix(): ParseBytesToIpNet fail:", err)
		return "", errors.New("data is not IP address")
	}
	return net.String(), nil
}

func ParseToAddressMinMax(data []byte, ipType int) (min, max string, err error) {
	belogs.Debug("ParseToAddressMinMax():data:", convert.PrintBytesOneLine(data), "  ipType:", ipType)

	var size int
	if ipType == 1 {
		size = 4
	} else if ipType == 2 {
		size = 16
	} else {
		belogs.Error("ParseToAddressMinMax(): ipType fail:", ipType)
		return "", "", errors.New("Not an IP address")
	}

	var ipAddrRange IpAddrRange
	_, err = asn1.Unmarshal(data, &ipAddrRange)
	if err != nil {
		belogs.Error("ParseToAddressMinMax():Unmarshal ipAddrRange fail:", convert.PrintBytesOneLine(data), err)
		return "", "", errors.New("data is not IP addresses(min/max)")
	}
	belogs.Debug("ParseToAddressMinMax():ipAddrRange:", ipAddrRange, err)

	// get min
	ipAddrMin := make([]byte, size)
	copy(ipAddrMin, ipAddrRange.Min.Bytes)
	netIpMin := net.IP(ipAddrMin)
	belogs.Info("ParseToAddressMinMax(): netIpMin:", netIpMin.String())

	// get max, and may be set 0xFF
	ipAddrMax := make([]byte, size)
	copy(ipAddrMax, ipAddrRange.Max.Bytes)
	for i := ipAddrRange.Max.BitLength/8 + 1; i < len(ipAddrMax); i++ {
		ipAddrMax[i] = 0xFF
	}
	if ipAddrRange.Max.BitLength/8 > len(ipAddrMax) {
		belogs.Error("ParseToAddressMinMax():max fail, ipAddrRange.Max.BitLength/8 > len(ipAddrMax):", convert.PrintBytesOneLine(ipAddrRange.Max.Bytes),
			"   ipAddrRange.Max.BitLength/8:", ipAddrRange.Max.BitLength/8, " len(ipAddrMax):", len(ipAddrMax))
		return "", "", errors.New("get max fail")
	}
	if ipAddrRange.Max.BitLength/8 < len(ipAddrMax) {
		ipAddrMax[ipAddrRange.Max.BitLength/8] |= 0xFF >> uint(8-(8*(ipAddrRange.Max.BitLength/8+1)-ipAddrRange.Max.BitLength))
	}
	netIpMax := net.IP(ipAddrMax)
	belogs.Info("ParseToAddressMinMax(): netIpMax:", netIpMax.String())

	return netIpMin.String(), netIpMax.String(), nil

}

func ParseToIpAddressBlocks(data []byte) ([]IpAddrBlock, error) {

	belogs.Debug("ParseToIpAddressBlocks(): data:", convert.PrintBytesOneLine(data))
	ipAddrBlocks := make([]IpAddrBlock, 0)

	var ipAddrRaws []IpAddrRaw
	_, err := asn1.Unmarshal(data, &ipAddrRaws)
	if err != nil {
		belogs.Error("ParseToIpAddressBlocks(): Unmarshal data fail:", convert.PrintBytesOneLine(data), err)
		return ipAddrBlocks, err
	}
	belogs.Debug("ParseToIpAddressBlocks(): ipAddrRaws:", jsonutil.MarshalJson(ipAddrRaws))

	for _, ipAddrRaw := range ipAddrRaws {
		// compatible
		var family uint64
		if len(ipAddrRaw.AddressFamily) == 2 {
			if ipAddrRaw.AddressFamily[1] == 1 {
				family = 1
			} else if ipAddrRaw.AddressFamily[1] == 2 {
				family = 2
			}
		} else if len(ipAddrRaw.AddressFamily) == 1 {
			if ipAddrRaw.AddressFamily[0] == 1 {
				family = 1
			} else if ipAddrRaw.AddressFamily[0] == 2 {
				family = 2
			}
		}
		if family == 0 {
			belogs.Error("ParseToIpAddressBlocks(): family fail:", family)
			return ipAddrBlocks, errors.New("family is error")
		}
		belogs.Debug("ParseToIpAddressBlocks(): ipAddrRaw.AddressFamily:", ipAddrRaw.AddressFamily,
			" family:", family)

		if ipAddrRaw.IpAddressChoice.Tag == asn1.TagNull {
			// is null
			ipAddrBlock := IpAddrBlock{AddressFamily: family}
			belogs.Debug("ParseToIpAddressBlocks():ipAddrRaw is TagNull:", ipAddrRaw.IpAddressChoice.Tag, ipAddrBlock)
			ipAddrBlocks = append(ipAddrBlocks, ipAddrBlock)

		} else if ipAddrRaw.IpAddressChoice.Tag == asn1.TagSequence {
			// have ips
			belogs.Debug("ParseToIpAddressBlocks():ipAddrRaw is TagSequence:", ipAddrRaw.IpAddressChoice.Tag)

			var ipAddrRawValues []asn1.RawValue
			_, err = asn1.Unmarshal(ipAddrRaw.IpAddressChoice.FullBytes, &ipAddrRawValues)
			if err != nil {
				belogs.Error("ParseToIpAddressBlocks():ipAddrRawValues Unmarshal fail:",
					convert.PrintBytesOneLine(ipAddrRaw.IpAddressChoice.FullBytes),
					err)
				return ipAddrBlocks, err
			}
			belogs.Debug("ParseToIpAddressBlocks(): len(ipAddrRawValues):", len(ipAddrRawValues))

			for _, ipAddrRawValue := range ipAddrRawValues {
				if ipAddrRawValue.Tag == asn1.TagBitString {

					addressPrefix, err := ParseBytesToAddressPrefix(ipAddrRawValue.Bytes, int(family))
					if err != nil {
						belogs.Error("ParseToIpAddressBlocks():TagBitString ParseBytesToAddressPrefix fail:",
							convert.PrintBytesOneLine(ipAddrRawValue.Bytes), family, err)
						return ipAddrBlocks, err
					}
					ipAddrBlock := IpAddrBlock{
						AddressFamily: family,
						AddressPrefix: addressPrefix}
					belogs.Debug("ParseToIpAddressBlocks():TagBitString  ipAddrBlock:", ipAddrRaw.IpAddressChoice.Tag, ipAddrBlock)
					ipAddrBlocks = append(ipAddrBlocks, ipAddrBlock)

				} else if ipAddrRawValue.Tag == asn1.TagSequence {

					//var ipAddrRange IpAddrRange
					//_, err := asn1.Unmarshal(ipAddrRawValue.FullBytes, &ipAddrRange)
					//belogs.Debug("ParseToIpAddressBlocks():TagSequence ipAddrRange:", jsonutil.MarshalJson(ipAddrRange))
					min, max, err := ParseToAddressMinMax(ipAddrRawValue.FullBytes, int(family))
					if err != nil {
						belogs.Error("ParseToIpAddressBlocks():TagSequence ParseToAddressMinMax fail:",
							convert.PrintBytesOneLine(ipAddrRawValue.FullBytes), family, err)
						return ipAddrBlocks, err
					}
					ipAddrBlock := IpAddrBlock{
						AddressFamily: family,
						Min:           min,
						Max:           max}
					belogs.Debug("ParseToIpAddressBlocks():TagSequence ipAddrBlock:", ipAddrRaw.IpAddressChoice.Tag, ipAddrBlock)
					ipAddrBlocks = append(ipAddrBlocks, ipAddrBlock)
				}
			}
		}
	}
	belogs.Info("ParseToIpAddressBlocks():ipAddrBlocks:", jsonutil.MarshalJson(ipAddrBlocks))
	return ipAddrBlocks, nil
}

type FileAndHash struct {
	File string `asn1:"ia5" json:"file"`
	Hash []byte `json:"hash"`
}

func ParseToFileAndHashs(data []byte) ([]FileAndHash, error) {
	belogs.Debug("ParseToFileAndHashs(): data:", convert.PrintBytesOneLine(data))
	fileAndHashs := make([]FileAndHash, 0)
	_, err := asn1.Unmarshal(data, &fileAndHashs)
	if err != nil {
		belogs.Error("ParseToFileAndHashs(): Unmarshal data fail:", convert.PrintBytesOneLine(data), err)
		return fileAndHashs, err
	}
	belogs.Debug("ParseToFileAndHashs(): fileAndHashs:", jsonutil.MarshalJson(fileAndHashs))
	return fileAndHashs, nil
}

type AsBlock struct {
	As  null.Int `json:"as"`
	Min null.Int `json:"min"`
	Max null.Int `json:"Max"`
}
type AsRaw struct {
	AsChoice asn1.RawValue `asn1:"explicit,optional,tag:0`
	Rdi      asn1.RawValue `asn1:"explicit,optional,tag:1`
}

type ASIdentifiersModel struct {
	Asnum ASIdentifierChoiceModel `asn1:"explicit,optional,tag:0`
	Rdi   ASIdentifierChoiceModel `asn1:"explicit,optional,tag:1`
}
type ASIdentifierChoiceModel struct {
	AsIdsOrRanges []AsnOrRangeModel
}
type AsnOrRangeModel struct {
	Asn          AsnModel     `asn1:"optional"`
	AsRangeModel AsRangeModel `asn1:"optional"`
}
type AsRangeModel struct {
	Min AsnModel
	Max AsnModel
}

type AsnModel int

func ParseToAsBlocks(data []byte) (asBlocks []AsBlock, err error) {

	belogs.Debug("ParseToAsBlocks(): data:", convert.PrintBytesOneLine(data))
	var asIdentifiersModel ASIdentifiersModel
	_, err = asn1.Unmarshal(data, &asIdentifiersModel)
	if err != nil {
		belogs.Error("ParseToAsBlocks(): Unmarshal data fail:", convert.PrintBytesOneLine(data), err)
		return asBlocks, err
	}
	/*


		asBlocks = make([]AsBlock, 0)
		var asRaws AsRaw
		_, err = asn1.Unmarshal(data, &asRaws)
		if err != nil {
			belogs.Error("ParseToAsBlocks(): Unmarshal data fail:", convert.PrintBytesOneLine(data), err)
			return asBlocks, err
		}
		belogs.Debug("ParseToAsBlocks(): asRaws:", asRaws, jsonutil.MarshalJson(asRaws))



			for _, asRaw := range asRaws {

				if asRaw.AsChoice.Tag == asn1.TagNull {
					// is null
					belogs.Debug("ParseToAsBlocks():asRaw is TagNull:", asRaw.AsChoice.Tag, asRaw)

				} else if asRaw.AsChoice.Tag == asn1.TagSequence {
					// have ips
					belogs.Debug("ParseToAsBlocks():asRaw is TagSequence:", asRaw.AsChoice.Tag)

					var asRawValues []asn1.RawValue
					_, err = asn1.Unmarshal(asRaw.AsChoice.FullBytes, &asRawValues)
					if err != nil {
						belogs.Error("ParseToAsBlocks():asRawValues Unmarshal fail:",
							convert.PrintBytesOneLine(asRaw.AsChoice.FullBytes),
							err)
						return asBlocks, err
					}
					belogs.Debug("ParseToAsBlocks(): len(asRawValues):", len(asRawValues))

					for _, asRawValue := range asRawValues {
						if asRawValue.Tag == asn1.TagInteger {
							as, err := asn1base.ParseBigInt(asRawValue.Bytes)
							if err != nil {
								belogs.Error("ParseToAsBlocks():TagInteger ParseBigInt fail:",
									convert.PrintBytesOneLine(asRawValue.Bytes), err)
								return asBlocks, err
							}
							asUint64 := as.Uint64()
							asBlock := AsBlock{As: null.IntFrom(int64(asUint64))}
							belogs.Debug("ParseToAsBlocks():TagInteger  AsBlock:", asRaw.AsChoice.Tag, asBlock)
							asBlocks = append(asBlocks, asBlock)

						} else if asRawValue.Tag == asn1.TagSequence {

							min, max, err := ParseToAsMinMax(asRawValue.FullBytes)
							if err != nil {
								belogs.Error("ParseToAsBlocks():TagSequence ParseToAsMinMax fail:",
									convert.PrintBytesOneLine(asRawValue.FullBytes), err)
								return asBlocks, err
							}
							asBlock := AsBlock{
								Min: null.IntFrom(int64(min)),
								Max: null.IntFrom(int64(max))}
							belogs.Debug("ParseToAsBlocks():TagSequence AsBlock:", asRaw.AsChoice.Tag, asBlock)
							asBlocks = append(asBlocks, asBlock)
						}
					}
				}
			}
	*/
	belogs.Info("ParseToAsBlocks():asBlocks:", jsonutil.MarshalJson(asBlocks))

	return
}

func ParseToAsMinMax(data []byte) (min, max int, err error) {
	belogs.Debug("ParseToAsMinMax():data:", convert.PrintBytesOneLine(data))

	var asRange AsRangeModel
	_, err = asn1.Unmarshal(data, &asRange)
	if err != nil {
		belogs.Error("ParseToAsMinMax():Unmarshal asRange fail:", convert.PrintBytesOneLine(data), err)
		return 0, 0, errors.New("data is not As(min/max)")
	}
	belogs.Debug("ParseToAsMinMax():asRange:", asRange)
	return int(asRange.Min), int(asRange.Max), nil
}

/* ok, but only ASId, not support min-max asn
type AsnPoint struct {
	AsnPointName AnsPointName `asn1:"optional,tag:0"`
}

type AnsPointName struct {
	AsnNames []RawValue //`asn1:"optional,tag:0"`
}

func GetAsns(value []byte) (AsnPoint, error) {
	fmt.Println("GetAsns(): value:", convert.PrintBytesOneLine(value))

	var asnPoint AsnPoint
	_, err := Unmarshal(value, &asnPoint)
	if err != nil {
		fmt.Println("GetAsns(): Unmarshal fail:", err)
		return asnPoint, err
	}

	fmt.Println("GetAsns(): asnPoint:", jsonutil.MarshalJson(asnPoint))
	for _, asnName := range asnPoint.AsnPointName.AsnNames {
		b := asnName.Bytes
		asn := big.NewInt(0).SetBytes(b)
		fmt.Println(asn)
	}
	return asnPoint, nil
}
*/
