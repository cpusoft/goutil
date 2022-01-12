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
	// ipv4+ipv6 range
	hexStr := `3031302004020001301A0304022D40B8030402671BC8300C03040067F5A503040067F5A6300D04020002300703050024077900`
	by, err := hex.DecodeString(hexStr)
	fmt.Println(convert.PrintBytesOneLine(by), err)
	ipAddrBlocks, err := ParseToIpAddressBlocks(by)
	fmt.Println(jsonutil.MarshalJson(ipAddrBlocks), err)
	fmt.Println("-------------------------")

	// only ipv6
	hexStr = `300F300D04020002300703050028041F54`
	by, err = hex.DecodeString(hexStr)
	fmt.Println(convert.PrintBytesOneLine(by), err)
	ipAddrBlocks, err = ParseToIpAddressBlocks(by)
	fmt.Println(jsonutil.MarshalJson(ipAddrBlocks), err)
	fmt.Println("-------------------------")

	// from sig
	hexStr = `3010300E04010230090307002001067C208C` // error: 040102 --> 04020002
	//hexStr = `3011300F0402000230090307002001067C208C`
	by, err = hex.DecodeString(hexStr)
	fmt.Println(convert.PrintBytesOneLine(by), err)
	ipAddrBlocks, err = ParseToIpAddressBlocks(by)
	fmt.Println(jsonutil.MarshalJson(ipAddrBlocks), err)
}

func TestParseToFileAndHashs(t *testing.T) {
	hexStr := `3077303416106234325F697076365F6C6F612E706E6704209516DD64BE7C1725B9FCA117120E58E8D842A5206873399B3DDFFC91C4B6ACF0303F161B6234325F736572766963655F646566696E6974696F6E2E6A736F6E04200AE1394722005CD92F4C6AA024D5D6B3E2E67D629F11720D9478A633A117A1C7`
	by, err := hex.DecodeString(hexStr)
	fmt.Println(convert.PrintBytesOneLine(by), err)
	fs, err := ParseToFileAndHashs(by)
	fmt.Println(jsonutil.MarshalJson(fs), err)
}
