package iputil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/convert"
)

func TestIpRangeIncludeInParentRange(t *testing.T) {

	self := make([]ChainIpAddress, 0)
	self1 := ChainIpAddress{
		RangeStart: "7a.09.00.00",
		RangeEnd:   "7a.09.00.ff",
	}
	self2 := ChainIpAddress{
		RangeStart: "7b.3c.00.00",
		RangeEnd:   "7b.3c.00.ff",
	}
	self = append(self, self1)
	self = append(self, self2)

	parent := make([]ChainIpAddress, 0)
	parent1 := ChainIpAddress{
		RangeStart: "7b.3c.00.00",
		RangeEnd:   "7b.3c.ff.ff",
	}
	parent2 := ChainIpAddress{
		RangeStart: "2407:1d00:0000:0000:0000:0000:0000:0000",
		RangeEnd:   "2407:1d00:ffff:ffff:ffff:ffff:ffff:ffff",
	}
	parent = append(parent, parent1)
	parent = append(parent, parent2)

	inRange := ipAddressesIncludeInParent(parent, self)
	fmt.Println(inRange)

	//7b.3c.00.00	7b.3c.00.ff
}

type ChainIpAddress struct {
	//min address range from addressPrefix or min/max, in hex:  63.60.00.00'
	RangeStart string `json:"rangeStart" xorm:"rangeStart varchar(512)"`
	//max address range from addressPrefix or min/max, in hex:  63.69.7f.ff'
	RangeEnd string `json:"rangeEnd" xorm:"rangeEnd varchar(512)"`
}

func ipAddressesIncludeInParent(parents []ChainIpAddress, self []ChainIpAddress) bool {
	fmt.Println(parents, self)
	for _, s := range self {
		include := false
		for _, p := range parents {
			fmt.Println("compare:", s, p)
			include = IpRangeIncludeInParentRange(p.RangeStart, p.RangeEnd, s.RangeStart, s.RangeEnd)
			if include {
				fmt.Println("include:", s, p)
				break
			}
		}
		if !include {
			fmt.Println("not include:", s)
			return false
		}
	}
	fmt.Println(true)
	return true
}

func TestFillAddressPrefixWithZero1(t *testing.T) {
	ipss := `16.7/16`
	ips, _ := FillAddressPrefixWithZero(ipss, GetIpType(ipss))
	fmt.Println(ips)
}

func TestRoaFormtToIp(t *testing.T) {
	b := []byte{0xb0, 0x10}
	di := RoaFormtToIp(b, 0x01)
	fmt.Println(di)

	b = []byte{0x03, 0x05, 0x00, 0x28, 0x03, 0xEA, 0x80}
	di = RoaFormtToIp(b, 0x02)
	fmt.Println(di)

	b = []byte{0x2a, 0x0, 0x15, 0x28, 0xaa, 0x0, 0xd0}
	di = RoaFormtToIp(b, 0x02)
	fmt.Println(di)
}

func TestIpToRtrFormat(t *testing.T) {
	// dig: 13635B00  str:19.99.91.0
	str := "19.99.91.0"
	di := IpToRtrFormat(str)
	fmt.Println(di)

	dig := "13635B00"
	fmt.Println(dig)

	str = "2001:DB8::"
	di = IpToRtrFormat(str)
	fmt.Println(di)

}

func TestIpToRtrFormatByte(t *testing.T) {
	// dig: 13635B00  str:19.99.91.0
	str := "19.99.91.0"
	di := IpToRtrFormatByte(str)
	fmt.Println(di)
	fmt.Println(convert.Bytes2String(di))

	str = "2001:DB8::"
	di2 := IpToRtrFormatByte(str)
	fmt.Println(di2)
	fmt.Println(convert.Bytes2String(di2))

}

func TestRtrFormatToIp(t *testing.T) {
	dig := []byte{80, 128, 0, 0}
	fmt.Println(len([]byte(dig)))
	str := RtrFormatToIp(dig)
	fmt.Println(str)

	dig = []byte{32, 1, 7, 248, 0, 25, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	str = RtrFormatToIp(dig)
	fmt.Println(str)
}

func TestTrimAddressPrefixZero(t *testing.T) {
	ips := []string{"16.70.0.0/16", "16.0.1.0/16"}

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

	ips := []string{"16.7", "16.7/16"}

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

	ips := []string{"45.5.68.0/24", "45.5.68.0/16"}
	for _, ip := range ips {
		ipp, err := FillAddressPrefixWithZero(ip, Ipv4Type)
		fmt.Println(ip, " --> ", ipp, err)
		min, max, err := AddressPrefixToHexRange(ipp, Ipv4Type)
		fmt.Println(ip, " --> ", min, max, err)

	}
	ips = []string{"45.5.69.0/24"}
	for _, ip := range ips {
		ipp, err := FillAddressPrefixWithZero(ip, Ipv4Type)
		fmt.Println(ip, " --> ", ipp, err)
		min, max, err := AddressPrefixToHexRange(ipp, Ipv4Type)
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

func TestIpStrToBinaryString(t *testing.T) {
	ip := "191.243.248"
	ipType := GetIpType(ip)
	ip1, _ := FillAddressWithZero(ip, ipType)
	binary, _ := IpStrToBinaryString(ip1, ipType)
	fmt.Println(ip, binary)
	binary, _ = IpAddressToBinaryString(ip)
	fmt.Println(ip, binary)

	ip = "2803:5360::"
	ipType = GetIpType(ip)
	ip1, _ = FillAddressWithZero(ip, ipType)
	binary, _ = IpStrToBinaryString(ip1, ipType)
	fmt.Println(ip, binary)
	binary, _ = IpAddressToBinaryString(ip)
	fmt.Println(ip, binary)
}
