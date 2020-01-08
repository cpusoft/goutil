package iputil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/convert"

	ip "."
)

func TestFillAddressPrefixWithZero1(t *testing.T) {
	ipss := `190.232/21`
	ips, _ := FillAddressPrefixWithZero(ipss, GetIpType(ipss))
	fmt.Println(ips)
}

func TestRoaFormtToIp(t *testing.T) {
	b := []byte{0xb0, 0x10}
	di := ip.RoaFormtToIp(b, 0x01)
	fmt.Println(di)

	b = []byte{0x03, 0x05, 0x00, 0x28, 0x03, 0xEA, 0x80}
	di = ip.RoaFormtToIp(b, 0x02)
	fmt.Println(di)

	b = []byte{0x2a, 0x0, 0x15, 0x28, 0xaa, 0x0, 0xd0}
	di = ip.RoaFormtToIp(b, 0x02)
	fmt.Println(di)
}

func TestIpToRtrFormat(t *testing.T) {
	// dig: 13635B00  str:19.99.91.0
	str := "19.99.91.0"
	di := ip.IpToRtrFormat(str)
	fmt.Println(di)

	dig := "13635B00"
	fmt.Println(dig)

	str = "2001:DB8::"
	di = ip.IpToRtrFormat(str)
	fmt.Println(di)

}

func TestIpToRtrFormatByte(t *testing.T) {
	// dig: 13635B00  str:19.99.91.0
	str := "19.99.91.0"
	di := ip.IpToRtrFormatByte(str)
	fmt.Println(di)
	fmt.Println(convert.Bytes2String(di))

	str = "2001:DB8::"
	di2 := ip.IpToRtrFormatByte(str)
	fmt.Println(di2)
	fmt.Println(convert.Bytes2String(di2))

}

func TestRtrFormatToIp(t *testing.T) {
	dig := []byte{80, 128, 0, 0}
	fmt.Println(len([]byte(dig)))
	str := ip.RtrFormatToIp(dig)
	fmt.Println(str)

	dig = []byte{32, 1, 7, 248, 0, 25, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	str = ip.RtrFormatToIp(dig)
	fmt.Println(str)
}

func TestTrimAddressPrefixZero(t *testing.T) {
	ips := []string{"19.0.1.0", "0.1.0.0/16"}

	for _, ip := range ips {
		str, err := TrimAddressPrefixZero(ip, Ipv4Type)
		fmt.Println(ip, " --> ", str, err)

	}
	ips = []string{"2803:d380::/28", "2803::/28"}
	for _, ip := range ips {
		str, err := TrimAddressPrefixZero(ip, Ipv6Type)
		fmt.Println(ip, " --> ", str, err)

	}
}

func TestFillAddressPrefixWithZero(t *testing.T) {

	ips := []string{"19.99.91", "19.99/16"}

	for _, ip := range ips {
		str, err := FillAddressPrefixWithZero(ip, Ipv4Type)
		fmt.Println(ip, " --> ", str, err)

	}
	ips = []string{"2803:d380/28", "2803/28"}
	for _, ip := range ips {
		str, err := FillAddressPrefixWithZero(ip, Ipv6Type)
		fmt.Println(ip, " --> ", str, err)

		strFill, err := IpStrToHexString(str, Ipv6Type)
		fmt.Println(str, " --> ", strFill, err)

	}
}

func TestIpStrToHexString(t *testing.T) {

	ips := []string{"19.99.91.3"}
	for _, ip := range ips {
		str, err := IpStrToHexString(ip, Ipv4Type)
		fmt.Println(ip, " --> ", str, err)

	}
	ips = []string{"2803:d380::", "2803:d380::5512"}
	for _, ip := range ips {
		str, err := IpStrToHexString(ip, Ipv6Type)
		fmt.Println(ip, " --> ", str, err)

	}
}

func TestAddressToRtrFormatByte(t *testing.T) {
	ips := []string{"192.216", "19.99.91.0", "2803:d380"}
	for _, ip := range ips {
		ipHex, ipType, err := AddressToRtrFormatByte(ip)
		fmt.Println(ip, " --> ", ipHex, ipType, err)
	}
}

func TestAddressPrefixToHexRange(t *testing.T) {

	ips := []string{"192.236/23"}
	for _, ip := range ips {
		ipp, err := FillAddressPrefixWithZero(ip, Ipv4Type)
		min, max, err := AddressPrefixToHexRange(ipp, Ipv4Type)
		fmt.Println(ip, " --> ", min, max, err)

	}
	ips = []string{"2803:d380/28"}
	for _, ip := range ips {
		ipp, err := FillAddressPrefixWithZero(ip, Ipv6Type)
		min, max, err := AddressPrefixToHexRange(ipp, Ipv6Type)
		fmt.Println(ip, " --> ", min, max, err)

	}
}

func TestSplitAddressAndPrefix(t *testing.T) {
	ap := "87.247.164/24"
	ip, pr, er := SplitAddressAndPrefix(ap)
	fmt.Println(ip, pr, er)
}

func TestIsAddressPrefix(t *testing.T) {
	ip := "182.18.223/24"
	is := IsAddressPrefix(ip)
	fmt.Println(is)
}
func TestIsAddress(t *testing.T) {
	ip := "182.18.223.1"
	is := IsAddress(ip)
	fmt.Println(is)

	ip = "182.18.223"
	is = IsAddress(ip)
	fmt.Println(is)

}
