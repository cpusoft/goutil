package asn1cert

import (
	"encoding/asn1"
	"errors"
	"net"

	"github.com/cpusoft/goutil/asn1util/asn1base"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

// ipv4: ipType==4, ipv6: ipType==6
// data: just ip data, no asn1 header
func ParseToIpNet(data []byte, ipType int) (*net.IPNet, error) {
	belogs.Debug("ParseToIpNet(): ipType:", ipType, "   len(data):", len(data))

	var size int
	if ipType == 1 {
		size = 4
	} else if ipType == 2 {
		size = 16
	} else {
		belogs.Error("ParseToIpNet(): ipType fail:", ipType)
		return nil, errors.New("Not an IP address")
	}

	bi, err := asn1base.ParseBitString(data)
	if err != nil {
		belogs.Error("ParseToIpNet(): ParseBitString fail:", convert.PrintBytesOneLine(data))
		return nil, errors.New("data is not IP address")
	}

	ipAddr := make([]byte, size)
	copy(ipAddr, bi.Bytes)
	mask := net.CIDRMask(bi.BitLength, size*8)
	belogs.Debug("ParseToIpNet(): ipAddr:", convert.PrintBytesOneLine(ipAddr),
		jsonutil.MarshalJson(ipAddr), "  mask:", mask)
	return &net.IPNet{
		IP:   net.IP(ipAddr),
		Mask: mask,
	}, nil

}

// use ParseToIpNet --> 134.144.0.0/16
func ParseToAddressPrefix(data []byte, ipType int) (string, error) {
	net, err := ParseToIpNet(data, ipType)
	if err != nil {
		belogs.Error("ParseToAddressPrefix(): ParseToIpNet fail:", err)
		return "", errors.New("data is not IP address")
	}
	return net.String(), nil
}

// addressPrefix or min/max
type IpAddrBlock struct {
	AddressFamily uint64 `json:"addressFamily"`
	//address prefix: 147.28.83.0/24 '
	AddressPrefix string `json:"addressPrefix"`
	//min address:  99.96.0.0
	Min string `json:"min"`
	//max address:   99.105.127.255
	Max string `json:"max"`
}
type IpAddrAsn1 struct {
	AddressFamily   []byte
	IpAddressChoice asn1.RawValue
}
type IpAddrRangeAsn1 struct {
	Min asn1.BitString
	Max asn1.BitString
}

func ParseToAddressMinMax(data []byte, ipType int) (min, max string, err error) {
	var size int
	if ipType == 1 {
		size = 4
	} else if ipType == 2 {
		size = 16
	} else {
		belogs.Error("ParseToAddressMinMax(): ipType fail:", ipType)
		return "", "", errors.New("Not an IP address")
	}

	var ipAddrRangeAsn1 IpAddrRangeAsn1
	_, err = asn1.Unmarshal(data, &ipAddrRangeAsn1)
	if err != nil {
		belogs.Error("ParseToAddressMinMax():Unmarshal ipAddrRangeAsn1 fail:", convert.PrintBytesOneLine(data), err)
		return "", "", errors.New("data is not IP addresses(min/max)")
	}
	belogs.Debug("ParseToAddressMinMax():ipAddrRangeAsn1:", ipAddrRangeAsn1, err)

	// get min
	ipAddrMin := make([]byte, size)
	copy(ipAddrMin, ipAddrRangeAsn1.Min.Bytes)
	netIpMin := net.IP(ipAddrMin)
	belogs.Info("ParseToAddressMinMax(): netIpMin:", netIpMin.String())

	// get max, and may be set 0xFF
	ipAddrMax := make([]byte, size)
	copy(ipAddrMax, ipAddrRangeAsn1.Max.Bytes)
	for i := ipAddrRangeAsn1.Max.BitLength/8 + 1; i < len(ipAddrMax); i++ {
		ipAddrMax[i] = 0xFF
	}
	if ipAddrRangeAsn1.Max.BitLength/8 > len(ipAddrMax) {
		belogs.Error("ParseToAddressMinMax():max fail, ipAddrRangeAsn1.Max.BitLength/8 > len(ipAddrMax):", convert.PrintBytesOneLine(ipAddrRangeAsn1.Max.Bytes),
			"   ipAddrRangeAsn1.Max.BitLength/8:", ipAddrRangeAsn1.Max.BitLength/8, " len(ipAddrMax):", len(ipAddrMax))
		return "", "", errors.New("get max fail")
	}
	if ipAddrRangeAsn1.Max.BitLength/8 < len(ipAddrMax) {
		ipAddrMax[ipAddrRangeAsn1.Max.BitLength/8] |= 0xFF >> uint(8-(8*(ipAddrRangeAsn1.Max.BitLength/8+1)-ipAddrRangeAsn1.Max.BitLength))
	}
	netIpMax := net.IP(ipAddrMax)
	belogs.Info("ParseToAddressMinMax(): netIpMax:", netIpMax.String())

	return netIpMin.String(), netIpMax.String(), nil

}

func ParseToIpAddressBlock(data []byte) ([]IpAddrBlock, error) {

	belogs.Debug("ParseToIpAddressBlock(): data:", convert.PrintBytesOneLine(data))
	ipAddrBlocks := make([]IpAddrBlock, 0)

	var ipAddrAsn1s []IpAddrAsn1
	_, err := asn1.Unmarshal(data, &ipAddrAsn1s)
	if err != nil {
		belogs.Error("ParseToIpAddressBlock(): Unmarshal data fail:", convert.PrintBytesOneLine(data), err)
		return ipAddrBlocks, err
	}
	belogs.Debug("ParseToIpAddressBlock(): ipAddrAsn1s:", jsonutil.MarshalJson(ipAddrAsn1s))

	for _, ipAddrssAsn1 := range ipAddrAsn1s {
		var family uint64
		if len(ipAddrssAsn1.AddressFamily) == 2 && ipAddrssAsn1.AddressFamily[1] == 1 {
			family = 1
		}
		if len(ipAddrssAsn1.AddressFamily) == 2 && ipAddrssAsn1.AddressFamily[1] == 2 {
			family = 2
		}
		belogs.Debug("ParseToIpAddressBlock(): ipAddrssAsn1.AddressFamily:", ipAddrssAsn1.AddressFamily,
			" family:", family)

		if ipAddrssAsn1.IpAddressChoice.Tag == asn1.TagNull {
			// is null
			ipAddrBlock := IpAddrBlock{AddressFamily: family}
			belogs.Debug("ParseToIpAddressBlock():ipAddrssAsn1 is TagNull:", ipAddrssAsn1.IpAddressChoice.Tag, ipAddrBlock)
			ipAddrBlocks = append(ipAddrBlocks, ipAddrBlock)

		} else if ipAddrssAsn1.IpAddressChoice.Tag == asn1.TagSequence {
			// have ips
			belogs.Debug("ParseToIpAddressBlock():ipAddrssAsn1 is TagSequence:", ipAddrssAsn1.IpAddressChoice.Tag)

			var ipAddrRawValues []asn1.RawValue
			_, err = asn1.Unmarshal(ipAddrssAsn1.IpAddressChoice.FullBytes, &ipAddrRawValues)
			if err != nil {
				belogs.Error("ParseToIpAddressBlock():ipAddrRawValues Unmarshal fail:",
					convert.PrintBytesOneLine(ipAddrssAsn1.IpAddressChoice.FullBytes),
					err)
				return ipAddrBlocks, err
			}
			belogs.Debug("ParseToIpAddressBlock(): len(ipAddrRawValues):", len(ipAddrRawValues))

			for _, ipAddrRawValue := range ipAddrRawValues {
				if ipAddrRawValue.Tag == asn1.TagBitString {

					addressPrefix, err := ParseToAddressPrefix(ipAddrRawValue.Bytes, int(family))
					if err != nil {
						belogs.Error("ParseToIpAddressBlock():TagBitString ParseToAddressPrefix fail:",
							convert.PrintBytesOneLine(ipAddrRawValue.Bytes), family, err)
						return ipAddrBlocks, err
					}
					ipAddrBlock := IpAddrBlock{
						AddressFamily: family,
						AddressPrefix: addressPrefix}
					belogs.Debug("ParseToIpAddressBlock():TagBitString  ipAddrBlock:", ipAddrssAsn1.IpAddressChoice.Tag, ipAddrBlock)
					ipAddrBlocks = append(ipAddrBlocks, ipAddrBlock)

				} else if ipAddrRawValue.Tag == asn1.TagSequence {

					//var ipAddrRangeAsn1 IpAddrRangeAsn1
					//_, err := asn1.Unmarshal(ipAddrRawValue.FullBytes, &ipAddrRangeAsn1)
					//belogs.Debug("ParseToIpAddressBlock():TagSequence ipAddrRangeAsn1:", jsonutil.MarshalJson(ipAddrRangeAsn1))
					min, max, err := ParseToAddressMinMax(ipAddrRawValue.FullBytes, int(family))
					if err != nil {
						belogs.Error("ParseToIpAddressBlock():TagSequence ParseToAddressMinMax fail:",
							convert.PrintBytesOneLine(ipAddrRawValue.FullBytes), family, err)
						return ipAddrBlocks, err
					}
					ipAddrBlock := IpAddrBlock{
						AddressFamily: family,
						Min:           min,
						Max:           max}
					belogs.Debug("ParseToIpAddressBlock():TagSequence ipAddrBlock:", ipAddrssAsn1.IpAddressChoice.Tag, ipAddrBlock)
					ipAddrBlocks = append(ipAddrBlocks, ipAddrBlock)
				}
			}
		}
	}
	belogs.Info("ParseToIpAddressBlock():ipAddrBlocks:", jsonutil.MarshalJson(ipAddrBlocks))
	return ipAddrBlocks, nil
}
