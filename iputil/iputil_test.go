package iputil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/convert"
)

// -------------------------- 基础函数测试 --------------------------
// TestIsIPv4_IsIPv6 测试IP类型判断函数
func TestIsIPv4_IsIPv6(t *testing.T) {
	tests := []struct {
		name  string
		ip    string
		want4 bool
		want6 bool
	}{
		// IPv4 正常场景
		{"IPv4 normal", "192.168.1.1", true, false},
		{"IPv4 min", "0.0.0.0", true, false},
		{"IPv4 max", "255.255.255.255", true, false},
		// IPv6 正常场景
		{"IPv6 normal", "2001:db8::1", false, true},
		{"IPv6 min", "::", false, true},
		{"IPv6 max", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", false, true},
		// 异常/边界场景
		{"IPv4 invalid segment", "192.168.1.256", false, false},
		{"IPv6 invalid char", "2001:db8::g", false, false},
		{"IPv4 mapped IPv6", "::ffff:192.168.1.1", false, false}, // 排除IPv4映射的IPv6
		{"empty string", "", false, false},
		{"random string", "abc123", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIPv4(tt.ip); got != tt.want4 {
				t.Errorf("IsIPv4(%q) = %v, want %v", tt.ip, got, tt.want4)
			}
			if got := IsIPv6(tt.ip); got != tt.want6 {
				t.Errorf("IsIPv6(%q) = %v, want %v", tt.ip, got, tt.want6)
			}
		})
	}
}

// TestCheckPrefixLengthOrMaxLength 测试前缀长度校验
func TestCheckPrefixLengthOrMaxLength(t *testing.T) {
	tests := []struct {
		name   string
		length int
		ipType int
		want   bool
	}{
		// IPv4 前缀
		{"IPv4 prefix 0", 0, Ipv4Type, false},
		{"IPv4 prefix 1", 1, Ipv4Type, true},
		{"IPv4 prefix 32 (max)", 32, Ipv4Type, true},
		{"IPv4 prefix 33 (over)", 33, Ipv4Type, false},
		// IPv6 前缀
		{"IPv6 prefix 0", 0, Ipv6Type, false},
		{"IPv6 prefix 1", 1, Ipv6Type, true},
		{"IPv6 prefix 128 (max)", 128, Ipv6Type, true},
		{"IPv6 prefix 129 (over)", 129, Ipv6Type, false},
		// 非法IP类型
		{"invalid ip type", 16, 0x03, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPrefixLengthOrMaxLength(tt.length, tt.ipType); got != tt.want {
				t.Errorf("CheckPrefixLengthOrMaxLength(%d, %d) = %v, want %v", tt.length, tt.ipType, got, tt.want)
			}
		})
	}
}

// -------------------------- RoaFormtToIp 测试 --------------------------
func TestRoaFormtToIp(t *testing.T) {
	tests := []struct {
		name   string
		ans1Ip []byte
		ipType int
		want   string
	}{
		// IPv4 场景（补全末尾省略的0）
		{"IPv4 2 bytes", []byte{210, 173}, Ipv4Type, "210.173.0.0"},
		{"IPv4 3 bytes", []byte{192, 168, 1}, Ipv4Type, "192.168.1.0"},
		{"IPv4 4 bytes (max)", []byte{255, 255, 255, 255}, Ipv4Type, "255.255.255.255"},
		{"IPv4 empty bytes", []byte{}, Ipv4Type, ""},
		// IPv6 场景（补全末尾省略的0 + 奇数长度补0）
		{"IPv6 2 bytes", []byte{0x28, 0x03}, Ipv6Type, "2803:0000:0000:0000:0000:0000:0000:0000"},
		{"IPv6 3 bytes (odd)", []byte{0x28, 0x03, 0xd3}, Ipv6Type, "2803:d300:0000:0000:0000:0000:0000:0000"},
		{"IPv6 16 bytes (full)", []byte{
			0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		}, Ipv6Type, "2001:0db8:0000:0000:0000:0000:0000:0001"},
		{"IPv6 empty bytes", []byte{}, Ipv6Type, ""},
		// 非法IP类型
		{"invalid ip type", []byte{192, 168}, 0x03, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RoaFormtToIp(tt.ans1Ip, tt.ipType); got != tt.want {
				t.Errorf("RoaFormtToIp(%v, %d) = %q, want %q", tt.ans1Ip, tt.ipType, got, tt.want)
			}
		})
	}
}

// -------------------------- IpToRtrFormat 测试 --------------------------
func TestIpToRtrFormat(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		// IPv4 场景（补全末尾省略的0）
		{"IPv4 2 segments", "192.168", "c0a80000"},
		{"IPv4 3 segments", "10.1.2", "0a010200"},
		{"IPv4 full (normal)", "192.168.1.1", "c0a80101"},
		{"IPv4 full (max)", "255.255.255.255", "ffffffff"},
		{"IPv4 invalid segment", "192.168.256.0", ""},
		// IPv6 场景（补全末尾省略的0）
		{"IPv6 2 segments", "2803:d380", "2803d380000000000000000000000000"},
		{"IPv6 compressed", "fe80::1", "fe800000000000000000000000000001"},
		{"IPv6 full (no compress)", "2001:0db8:0000:0000:0000:0000:0000:0001", "20010db8000000000000000000000001"},
		{"IPv6 invalid char", "2001:db8::g", ""},
		// 非法场景
		{"empty string", "", ""},
		{"random string", "abc123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IpToRtrFormat(tt.ip); got != tt.want {
				t.Errorf("IpToRtrFormat(%q) = %q, want %q", tt.ip, got, tt.want)
			}
		})
	}
}

// -------------------------- CompressFillIpv6 测试 --------------------------
func TestCompressFillIpv6(t *testing.T) {
	tests := []struct {
		name  string
		oldIp string
		want  string
	}{
		{"IPv6 full zero", "0000:0000:0000:0000:0000:0000:0000:0000", "::"},
		{"IPv6 partial zero", "2001:0db8:0000:0000:0000:0000:0000:0001", "2001:db8::1"},
		{"IPv6 end zero", "2001:0db8:1:0000:0000:0000:0000:0000", "2001:db8:1::"},
		{"IPv6 start zero", "0000:0000:0000:0000:0000:0000:2001:db8", "::2001:db8"},
		{"IPv6 no compress needed", "2001:db8:1:2:3:4:5:6", "2001:db8:1:2:3:4:5:6"},
		{"IPv6 invalid", "2001:db8::g", "2001:db8::g"}, // 解析失败返回原字符串
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompressFillIpv6(tt.oldIp); got != tt.want {
				t.Errorf("CompressFillIpv6(%q) = %q, want %q", tt.oldIp, got, tt.want)
			}
		})
	}
}

// -------------------------- FillAddressWithZero 测试 --------------------------
func TestFillAddressWithZero(t *testing.T) {
	tests := []struct {
		name    string
		address string
		ipType  int
		want    string
		wantErr bool
	}{
		// IPv4 场景
		{"IPv4 2 segments", "192.168", Ipv4Type, "192.168.0.0", false},
		{"IPv4 3 segments", "10.1.2", Ipv4Type, "10.1.2.0", false},
		{"IPv4 full", "255.255.255.255", Ipv4Type, "255.255.255.255", false},
		{"IPv4 over segments", "192.168.1.2.3", Ipv4Type, "", true},
		// IPv6 场景
		{"IPv6 2 segments", "2803:d380", Ipv6Type, "2803:d380:0000:0000:0000:0000:0000:0000", false},
		{"IPv6 compressed", "::", Ipv6Type, "0000:0000:0000:0000:0000:0000:0000:0000", false},
		{"IPv6 invalid", "2001:db8::g", Ipv6Type, "", true},
		// 非法IP类型
		{"invalid ip type", "192.168", 0x03, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FillAddressWithZero(tt.address, tt.ipType)
			if (err != nil) != tt.wantErr {
				t.Errorf("FillAddressWithZero(%q, %d) error = %v, wantErr %v", tt.address, tt.ipType, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FillAddressWithZero(%q, %d) = %q, want %q", tt.address, tt.ipType, got, tt.want)
			}
		})
	}
}

// -------------------------- IsAddressPrefix 测试 --------------------------
func TestIsAddressPrefix(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		// 合法前缀
		{"IPv4 prefix normal", "192.168.1/24", true},
		{"IPv4 prefix max", "0.0.0.0/32", true},
		{"IPv6 prefix normal", "2001:db8::/32", true},
		{"IPv6 prefix max", "::/128", true},
		// 非法前缀
		{"IPv4 prefix over", "192.168.1/33", false},
		{"IPv6 prefix over", "2001:db8::/129", false},
		{"no slash", "192.168.1.1", false},
		{"invalid ip", "192.168.256/24", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAddressPrefix(tt.ip); got != tt.want {
				t.Errorf("IsAddressPrefix(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

// -------------------------- 性能测试（Benchmark） --------------------------
// BenchmarkIsIPv4 IP类型判断性能
func BenchmarkIsIPv4(b *testing.B) {
	ipList := []string{"192.168.1.1", "0.0.0.0", "255.255.255.255", "2001:db8::1"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ip := range ipList {
			IsIPv4(ip)
		}
	}
}

// BenchmarkIpToRtrFormat IP转Rtr格式性能
func BenchmarkIpToRtrFormat(b *testing.B) {
	ipList := []string{"192.168", "2803:d380", "192.168.1.1", "fe80::1"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ip := range ipList {
			IpToRtrFormat(ip)
		}
	}
}

// BenchmarkCompressFillIpv6 IPv6压缩性能
func BenchmarkCompressFillIpv6(b *testing.B) {
	ipList := []string{
		"0000:0000:0000:0000:0000:0000:0000:0000",
		"2001:0db8:0000:0000:0000:0000:0000:0001",
		"fe80:0000:0000:0000:0000:0000:0000:0001",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ip := range ipList {
			CompressFillIpv6(ip)
		}
	}
}

// BenchmarkRoaFormtToIp Roa格式转IP性能
func BenchmarkRoaFormtToIp(b *testing.B) {
	tests := []struct {
		ans1Ip []byte
		ipType int
	}{
		{[]byte{192, 168}, Ipv4Type},
		{[]byte{0x28, 0x03}, Ipv6Type},
		{[]byte{255, 255, 255, 255}, Ipv4Type},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tt := range tests {
			RoaFormtToIp(tt.ans1Ip, tt.ipType)
		}
	}
}

/////////////////////////////////////////////
//////////////////////////////////////////////////////////////
//////////////////////

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

func TestRoaFormtToIp1(t *testing.T) {
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

func TestIpToRtrFormat1(t *testing.T) {
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

	ips := []string{"19.99.91.3", "192.168.0.0", "192.168"}
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
func TestIpStrIncompleteToHexString(t *testing.T) {

	ips := []string{"19.99.91.3", "192.168.0.0", "192.168"}
	for _, ip := range ips {
		str, err := IpStrIncompleteToHexString(ip)
		fmt.Println(ip, " --> ", str, err)
	}
	ips = []string{"2803:d380::", "2803:d380::5512", "03:d380::12"}
	for _, ip := range ips {
		str, err := IpStrIncompleteToHexString(ip)
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

func TestIsAddressPrefix1(t *testing.T) {
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

func TestIpToDnsFormatByte(t *testing.T) {
	ip := `1.1.1.1`
	b := IpToDnsFormatByte(ip)
	fmt.Println(convert.PrintBytesOneLine(b))
	ip1 := DnsFormatToIp(b, true)
	fmt.Println(ip1)

	ip = `2001:67c:1562::1c`
	b = IpToDnsFormatByte(ip)
	fmt.Println(convert.PrintBytesOneLine(b))
	ip1 = DnsFormatToIp(b, true)
	fmt.Println(ip1)

	ip = `200Z:67c:1562::1c`
	b = IpToDnsFormatByte(ip)
	fmt.Println(convert.PrintBytesOneLine(b))
	ip1 = DnsFormatToIp(b, true)
	fmt.Println(ip1)
}

func TestCompressFillIpv61(t *testing.T) {
	ips := []string{"2001:0db8:0000:0000:0001:0000:0000:0000",
		`2001:0000:0000:0000:0000:0000:0000:0001`,
		`2001:0001:0000:0000:0000:0000:0000:0000`,
		`0000:0000:0000:0000:0000:0000:2001:0001`,
		`240e:0014:6000:0000:0000:0000:0000:0001`}
	for i := range ips {
		fmt.Println(i, ":")
		newIp := CompressFillIpv6(ips[i])
		fmt.Println(ips[i], newIp)
	}
}

func TestIsAddressPrefixRangeContains(t *testing.T) {
	p := `192.168/16`
	c := `192.169.5/24`
	is, err := IsAddressPrefixRangeContains(p, c)
	fmt.Println(is, err)

	p = `2803:d380/28`
	c = `2803:d380/29`
	is, err = IsAddressPrefixRangeContains(p, c)
	fmt.Println(is, err)
}

func TestAddressPrefixToAsn1HexFormat(t *testing.T) {
	addressPrefix := `2001:0:200:3::/64` //`203.0.113.0/27` //`10.1.0.0/16`
	hex, err := AddressPrefixToAsn1HexFormat(addressPrefix)
	fmt.Println(hex, err)

}
