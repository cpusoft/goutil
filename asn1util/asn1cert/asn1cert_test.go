package asn1cert

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseToIpNet(t *testing.T) {
	hexStr := `002001067C208C`
	by, err := hex.DecodeString(hexStr)
	fmt.Println(convert.PrintBytesOneLine(by), err)
	ipNet, err := ParseToIpNet(by, 2)
	fmt.Println(jsonutil.MarshalJson(ipNet), err)
	addressPrefix, err := ParseToAddressPrefix(by, 2)
	fmt.Println("addressPrefix:", addressPrefix, err)
	newAdd, err := iputil.TrimAddressPrefixZero(addressPrefix, 2)
	fmt.Println("newAdd:", newAdd, err)

	hexStr = `074E8200`
	by, err = hex.DecodeString(hexStr)
	fmt.Println(convert.PrintBytesOneLine(by), err)
	ipNet, err = ParseToIpNet(by, 1)
	ones, bits := ipNet.Mask.Size()
	fmt.Println(jsonutil.MarshalJson(ipNet), ones, bits, err)
	addressPrefix, err = ParseToAddressPrefix(by, 1)
	fmt.Println("addressPrefix:", addressPrefix, err)
	newAdd, err = iputil.TrimAddressPrefixZero(addressPrefix, 1)
	fmt.Println("newAdd:", newAdd, err)
}

func TestParseToAddressMinMax(t *testing.T) {
	hexStr := `300e030500c0a80001030500c0a80003`
	by, err := hex.DecodeString(hexStr)
	fmt.Println(convert.PrintBytesOneLine(by), err)
	min, max, err := ParseToAddressMinMax(by, 1)
	fmt.Println(min, max, err)
}

func TestParseToIpAddressBlock(t *testing.T) {
	hexStr := `3031302004020001301A0304022D40B8030402671BC8300C03040067F5A503040067F5A6300D04020002300703050024077900`
	by, err := hex.DecodeString(hexStr)
	fmt.Println(convert.PrintBytesOneLine(by), err)

	ipAddrBlocks, err := ParseToIpAddressBlock(by)
	fmt.Println(jsonutil.MarshalJson(ipAddrBlocks), err)
}
