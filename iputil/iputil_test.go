package iputil

import (
	"fmt"
	"net"
	"testing"

	"github.com/cpusoft/goutil/convert"

	ip "."
)

func TestFillAddressPrefixWithZero1(t *testing.T) {
	ipss := `91.228.164/24`
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
	ips := []string{"19.99.91.0", "19.99.0.0/16"}

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

func TestSummarize(t *testing.T) {
	pss := IpRangeToAddressPrefixRanges("194.193.128.0", "194.193.223.255")
	for _, p := range pss {
		fmt.Println(p)
	}

	ps := Summarize(net.ParseIP("2001:db8::1"), net.ParseIP("2001:db8::8000"))
	for _, p := range ps {
		fmt.Println(p)
	}
	/*
	   2001:db8::1/128
	   2001:db8::2/127
	   2001:db8::4/126
	   2001:db8::8/125
	   2001:db8::10/124
	   2001:db8::20/123
	   2001:db8::40/122
	   2001:db8::80/121
	   2001:db8::100/120
	   2001:db8::200/119
	   2001:db8::400/118
	   2001:db8::800/117
	   2001:db8::1000/116
	   2001:db8::2000/115
	   2001:db8::4000/114
	   2001:db8::8000/128
	*/
	ps = Summarize(net.ParseIP("194.193.128.0"), net.ParseIP("194.193.223.255"))
	for _, p := range ps {
		fmt.Println(p)
	}
	/*
		194.193.128.0/18
		194.193.192.0/19
	*/
	ps = Summarize(net.ParseIP("194.223.0.0"), net.ParseIP("194.223.95.255"))
	for _, p := range ps {
		fmt.Println(p)
	}
	/*
		194.223.0.0/18
		194.223.64.0/19
	*/
	ps = Summarize(net.ParseIP("2001:7fa:9::"), net.ParseIP("2001:7fa:e:ffff:ffff:ffff:ffff:ffff"))
	for _, p := range ps {
		fmt.Println(p)
	}
	/*
		2001:7fa:9::/48
		2001:7fa:a::/47
		2001:7fa:c::/47
		2001:7fa:e::/48
	*/

}
