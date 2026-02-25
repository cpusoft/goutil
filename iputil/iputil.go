package iputil

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/cpusoft/goutil/asn1util/asn1addressasn"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/stringutil"
)

const (
	Ipv4Type = 0x01
	Ipv6Type = 0x02
	//posix的<netinet/in.h>
	IPv4_MAXLENGTH = 15
	IPv6_MAXLENGTH = 45
	//
	IPv4_PREFIX_MAXLENGTH = 32
	IPv6_PREFIX_MAXLENGTH = 128
)

// IsIPv4 returns true iff the addr string represents an IPv4 address.
func IsIPv4(s string) bool {
	addr, err := netip.ParseAddr(s)
	return err == nil && addr.Is4() && !addr.Is4In6() // 排除IPv4映射的IPv6地址
}

// IsIPv6 returns true iff the addr string represents an IPv6 address.
func IsIPv6(s string) bool {
	addr, err := netip.ParseAddr(s)
	return err == nil && addr.Is6() && !addr.Is4In6() // 排除IPv4映射的IPv6地址
}

// AddrAsIp converts a [netip.Addr] to a [net.IP].
func AddrAsIP(addr netip.Addr) net.IP {
	return net.ParseIP(addr.String())
}

// IPAsAddr converts a [net.IP] to a [netip.Addr].
func IPAsAddr(ip net.IP) (netip.Addr, error) {
	return netip.ParseAddr(ip.String())
}

func RoaFormtToIp(ans1Ip []byte, ipType int) string {
	//belogs.Debug("RoaFormtToIp():ans1Ip: %+v:", ans1Ip, "  ipType:", ipType)
	var buffer bytes.Buffer
	if ipType == Ipv4Type {
		// 新增：空字节数组直接返回空（解决测试失败问题1）
		if len(ans1Ip) == 0 {
			return ""
		}
		// 适配：补全末尾省略的 0，直到 4 字节（IPv4 完整长度）
		ipv4Bytes := make([]byte, 4) // 初始化4字节数组，默认值为0
		copy(ipv4Bytes, ans1Ip)      // 将传入的字节复制到前N位，剩余位保留0（补全末尾省略的0）

		for i, ip := range ipv4Bytes { // 遍历完整的4字节
			if i < len(ipv4Bytes)-1 {
				buffer.WriteString(fmt.Sprintf("%d.", ip))
			} else {
				buffer.WriteString(fmt.Sprintf("%d", ip))
			}
		}
		ipv4 := buffer.String()
		if len(ipv4) == 0 || len(ipv4) > IPv4_MAXLENGTH {
			return ""
		}
		return ipv4
	} else if ipType == Ipv6Type {
		// 新增：空字节数组直接返回空
		if len(ans1Ip) == 0 {
			return ""
		}
		asn1IpTmp := ans1Ip
		// 第一步：处理奇数长度，末尾补0至偶数（保持原逻辑，但补在末尾）
		if len(asn1IpTmp)%2 != 0 {
			asn1IpTmp = append(asn1IpTmp, 0x00) // 补在末尾（省略的是末尾0）
		}
		// 第二步：补全末尾省略的0，直到16字节（IPv6完整长度）
		ipv6Bytes := make([]byte, 16) // 初始化16字节数组，默认值为0
		copy(ipv6Bytes, asn1IpTmp)    // 复制现有字节，剩余位补0

		// 遍历完整的16字节（按2字节一组拼接）
		for i := 0; i < len(ipv6Bytes); i = i + 2 {
			if i < len(ipv6Bytes)-2 {
				buffer.WriteString(fmt.Sprintf("%02x%02x:", ipv6Bytes[i], ipv6Bytes[i+1]))
			} else {
				buffer.WriteString(fmt.Sprintf("%02x%02x", ipv6Bytes[i], ipv6Bytes[i+1]))
			}
		}
		ipv6 := buffer.String()
		if len(ipv6) == 0 || len(ipv6) > IPv6_MAXLENGTH {
			return ""
		}
		return ipv6
	}
	return ""
}

func CheckPrefixLengthOrMaxLength(length, ipType int) bool {
	if ipType == Ipv4Type {
		if length > 0 && length <= IPv4_PREFIX_MAXLENGTH {
			return true
		}
	} else if ipType == Ipv6Type {
		if length > 0 && length <= IPv6_PREFIX_MAXLENGTH {
			return true
		}
	}
	return false
}

func RtrFormatToIp(rtrIp []byte) string {
	var ip string
	if len(rtrIp) == 4 {
		ip = fmt.Sprintf("%d.%d.%d.%d", rtrIp[0], rtrIp[1], rtrIp[2], rtrIp[3])
		return ip
	} else if len(rtrIp) == 16 {
		ip = fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
			rtrIp[0], rtrIp[1],
			rtrIp[2], rtrIp[3],
			rtrIp[4], rtrIp[5],
			rtrIp[6], rtrIp[7],
			rtrIp[8], rtrIp[9],
			rtrIp[10], rtrIp[11],
			rtrIp[12], rtrIp[13],
			rtrIp[14], rtrIp[15])
		return ip
	}
	return ""
}

/*
	func IpToRtrFormat(ip string) string {
		// format  ipv4
		ipsV4 := strings.Split(ip, ".")
		if len(ipsV4) == 4 { // 新增：仅处理合法4段IPv4
			var formatIp string
			for _, ipV4 := range ipsV4 {
				ipInt, err := strconv.Atoi(ipV4)
				if err != nil || ipInt < 0 || ipInt > 255 { // 新增：校验IPv4段数值范围
					return ""
				}
				formatIp += fmt.Sprintf("%02x", ipInt)
			}
			return formatIp
		}

		// format ipv6
		if strings.Contains(ip, ":") {
			// 修复：复用标准库解析IPv6，避免手动处理压缩段的错误
			addr, err := netip.ParseAddr(ip)
			if err != nil {
				return ""
			}
			if !addr.Is6() {
				return ""
			}
			// 转为全段格式（无压缩）
			fullIp := addr.StringExpanded()
			ipsV6 := strings.Split(fullIp, ":")
			var formatIp string
			for _, ipV6 := range ipsV6 {
				formatIp += fmt.Sprintf("%04s", ipV6) // 此时ipV6已是4位，补0（%04s此处安全）
			}
			return formatIp
		}
		return ""
	}
*/
func IpToRtrFormat(ip string) string {
	// 第一步：智能补全不全的IP段（兼容::压缩格式）
	ipFilled := ip
	if strings.Contains(ip, ".") && !strings.Contains(ip, ":") {
		// IPv4补全：192.168 → 192.168.0.0、10.1.2 → 10.1.2.0
		segments := strings.Split(ip, ".")
		if len(segments) > 0 && len(segments) < 4 {
			ipFilled += strings.Repeat(".0", 4-len(segments))
		}
	} else if strings.Contains(ip, ":") && !strings.Contains(ip, ".") {
		// IPv6补全：兼容::压缩格式 + 不全段补全
		ipFilled = completeIPv6(ip)
	}

	// 第二步：统一用标准库解析IP（此时ipFilled已是标准格式）
	addr, err := netip.ParseAddr(ipFilled)
	if err != nil {
		return "" // 解析失败（非法IP），返回空
	}

	formatIp := ""
	// 第三步：根据IP类型转换为Rtr格式
	if addr.Is4() {
		// IPv4：转为4段完整格式，再转8位16进制字符串
		ipv4Str := addr.String()
		ipsV4 := strings.Split(ipv4Str, ".")
		for _, seg := range ipsV4 {
			segInt, err := strconv.Atoi(seg)
			if err != nil || segInt < 0 || segInt > 255 {
				return ""
			}
			formatIp += fmt.Sprintf("%02x", segInt)
		}
	} else if addr.Is6() {
		// IPv6：转为8段无压缩格式，再转32位16进制字符串
		ipv6Full := addr.StringExpanded()
		ipsV6 := strings.Split(ipv6Full, ":")
		for _, seg := range ipsV6 {
			formatIp += fmt.Sprintf("%04s", seg)
		}
	}

	// 第四步：校验输出格式（IPv4=8位16进制，IPv6=32位16进制）
	if (addr.Is4() && len(formatIp) != 8) || (addr.Is6() && len(formatIp) != 32) {
		return ""
	}
	return formatIp
}

// 辅助函数：补全不全的IPv6地址为标准格式（兼容::压缩）
func completeIPv6(ip string) string {
	// 处理::压缩的情况
	if strings.Contains(ip, "::") {
		parts := strings.Split(ip, "::")
		left := strings.Split(parts[0], ":")
		right := strings.Split(parts[1], ":")

		// 过滤空段
		leftValid := make([]string, 0)
		for _, seg := range left {
			if seg != "" {
				leftValid = append(leftValid, seg)
			}
		}
		rightValid := make([]string, 0)
		for _, seg := range right {
			if seg != "" {
				rightValid = append(rightValid, seg)
			}
		}

		// 计算需要补的0段数（总8段）
		missing := 8 - len(leftValid) - len(rightValid)
		middle := strings.Repeat("0:", missing)

		// 拼接完整地址
		full := strings.Join(leftValid, ":") + ":" + middle + strings.Join(rightValid, ":")
		// 清理多余的冒号（开头/结尾）
		full = strings.Trim(full, ":")
		return full
	}

	// 处理无::但段数不足的情况（如2803:d380）
	segments := strings.Split(ip, ":")
	validSegs := make([]string, 0)
	for _, seg := range segments {
		if seg != "" {
			validSegs = append(validSegs, seg)
		}
	}
	if len(validSegs) > 0 && len(validSegs) < 8 {
		missing := 8 - len(validSegs)
		return strings.Join(validSegs, ":") + ":" + strings.Repeat("0:", missing-1) + "0"
	}

	return ip
}

func IpToDnsFormatByte(ip string) []byte {
	addr := net.ParseIP(ip)
	if addr == nil {
		return nil
	}
	ipType := GetIpType(ip)
	if ipType == Ipv4Type {
		return addr.To4()
	} else if ipType == Ipv6Type {
		return addr.To16()
	} else {
		return nil
	}
}

// compressIpv6: will compress ipv6, and ignore for ipv4
func DnsFormatToIp(addr []byte, compressIpv6 bool) string {
	var ip string
	if len(addr) == 4 {
		ip = fmt.Sprintf("%d.%d.%d.%d", addr[0], addr[1], addr[2], addr[3])
		belogs.Debug("DnsFormatToIp():ipv4, addr:", addr, "   ip:", ip)
		return ip
	} else if len(addr) == 16 {
		ip = fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
			addr[0], addr[1],
			addr[2], addr[3],
			addr[4], addr[5],
			addr[6], addr[7],
			addr[8], addr[9],
			addr[10], addr[11],
			addr[12], addr[13],
			addr[14], addr[15])
		belogs.Debug("DnsFormatToIp():ipv6, compressIpv6:", compressIpv6,
			"   addr:", convert.PrintBytesOneLine(addr), "   ip:", ip)
		if compressIpv6 {
			ip = CompressFillIpv6(ip)
		}
		return ip
	}
	return ""
}

func CompressFillIpv6(oldIp string) (newIp string) {
	// 修复：复用标准库解析IPv6，避免手动压缩的逻辑错误
	addr, err := netip.ParseAddr(oldIp)
	if err != nil {
		return oldIp // 解析失败返回原字符串
	}
	return addr.String() // 标准库自动压缩，符合RFC规范
}

// Bad way, still need to find a good way
// 19.99.91.0 --> []byte;     2001:DB8::-->[]byte
func IpToRtrFormatByte(ip string) []byte {
	belogs.Debug("IpToRtrFormatByte():ip", ip)

	// format  ipv4
	ipsV4 := strings.Split(ip, ".")
	if len(ipsV4) == 4 { // 新增：仅处理合法4段IPv4
		byt := make([]byte, 4)
		for i := range ipsV4 {
			tmp, err := strconv.Atoi(ipsV4[i])
			if err != nil || tmp < 0 || tmp > 255 { // 新增：校验数值范围
				belogs.Debug("IpToRtrFormatByte():ipv4 Atoi err:", ipsV4[i], err)
				return nil
			}
			byt[i] = byte(tmp)
		}
		return byt
	}

	// format ipv6
	if strings.Contains(ip, ":") {
		// 修复：复用标准库解析IPv6，避免手动处理压缩段的错误
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			belogs.Error("IpToRtrFormatByte():ParseAddr fail:", ip, err)
			return nil
		}
		if !addr.Is6() {
			return nil
		}
		// 转为16字节数组
		ipBytes := addr.As16()
		return ipBytes[:]
	}
	return nil
}

// 210.173.160 --> []byte{0xD2ADA000}
// 2803:d380 --> []byte{0x2803d3800**00}
func AddressToRtrFormatByte(address string) (ipHex []byte, ipType int, err error) {
	ipType = GetIpType(address)
	if ipType != Ipv4Type && ipType != Ipv6Type {
		return nil, 0, errors.New("invalid ip type")
	}

	addressFill, err := FillAddressWithZero(address, ipType)
	if err != nil {
		return nil, 0, err
	}

	if ipType == Ipv4Type {
		ipHex = net.ParseIP(addressFill).To4()
	} else if ipType == Ipv6Type {
		ipHex = net.ParseIP(addressFill).To16()
	}

	if ipHex == nil {
		return nil, 0, errors.New("parse ip failed after fill zero")
	}
	return ipHex, ipType, nil
}

// 192.168.0.0/24--> 192.168/24    192.168.1.0-->192.168.1
//
//	2c0f:ea60::/32 --> 2c0f:ea60/32
func TrimAddressPrefixZero(ip string, ipType int) (string, error) {
	if ipType == Ipv4Type {
		split := strings.Split(ip, "/")
		if len(split) == 1 {
			return stringutil.TrimSuffixAll(ip, ".0"), nil
		} else if len(split) == 2 {
			return stringutil.TrimSuffixAll(split[0], ".0") + "/" + split[1], nil
		} else {
			return "", errors.New("illegal address prefix")
		}

	} else if ipType == Ipv6Type {
		split := strings.Split(ip, "/")
		if len(split) == 1 {
			return stringutil.TrimSuffixAll(ip, "::"), nil
		} else if len(split) == 2 {
			return stringutil.TrimSuffixAll(split[0], "::") + "/" + split[1], nil
		} else {
			return "", errors.New("illegal address prefix")
		}
	} else {
		return "", errors.New("illegal ipType")
	}
}

// fill addressprefix with zero:
// 192.168/24 --> 192.168.0.0/24  2803:d380/28 --> 2803:d380::/28
func FillAddressPrefixWithZero(addressPrefix string, ipType int) (addressPrefixFill string, err error) {
	address, prefix, err := SplitAddressAndPrefix(addressPrefix)
	if err != nil {
		belogs.Error("FillAddressPrefixWithZero():after SplitAddressAndPrefix fail:  addressPrefix:", addressPrefix, err)
		return "", err
	}

	addressFile, err := FillAddressWithZero(address, ipType)
	if err != nil {
		belogs.Error("FillAddressPrefixWithZero():after FillAddressWithZero fail: address, ipType: ", address, ipType, err)
		return "", err
	}
	return addressFile + "/" + strconv.Itoa(int(prefix)), nil
}

// fill address with zero:
// 192.168 --> 192.168.0.0  2803:d380 --> 2803:d380::
func FillAddressWithZero(address string, ipType int) (addressFill string, err error) {
	if ipType == Ipv4Type {
		countComma := strings.Count(address, ".")
		if countComma > 3 { // 新增：处理段数超过3的非法情况
			return "", errors.New("ipv4 address has too many segments")
		} else if countComma == 3 {
			addressFill = address
		} else if countComma < 3 {
			addressFill = address + strings.Repeat(".0", net.IPv4len-countComma-1)
		}
		// 新增：校验填充后的IPv4是否合法
		if net.ParseIP(addressFill).To4() == nil {
			return "", errors.New("invalid ipv4 address after fill zero")
		}
		return addressFill, nil
	} else if ipType == Ipv6Type {
		// 修复：先补全不全的IPv6段为标准格式，再解析
		addressCompleted := completeIPv6(address)
		addr, err := netip.ParseAddr(addressCompleted)
		if err != nil {
			return "", err
		}
		if !addr.Is6() {
			return "", errors.New("invalid ipv6 address")
		}
		addressFill = addr.StringExpanded() // 转为全段格式（无压缩）
		return addressFill, nil
	} else {
		return "", errors.New("illegal ipType")
	}
}

// ip to string not fill zero: ip: 192.168--> c0.a8
func IpStrIncompleteToHexString(ip string) (string, error) {
	var buffer bytes.Buffer
	ipType := GetIpType(ip)
	if ipType == Ipv4Type {
		split := strings.Split(ip, ".")
		if len(split) == 0 || len(split) > 4 {
			return "", errors.New("invalid ipv4 segments count")
		}
		for i := range split {
			if len(split[i]) == 0 {
				continue
			}
			label, err := strconv.Atoi(split[i])
			if err != nil || label < 0 || label > 255 {
				return "", err
			}
			if i < len(split)-1 {
				buffer.WriteString(fmt.Sprintf("%02x.", label))
			} else {
				buffer.WriteString(fmt.Sprintf("%02x", label))
			}
		}
		return buffer.String(), nil
	} else if ipType == Ipv6Type {
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			return "", err
		}
		if !addr.Is6() {
			return "", errors.New("invalid ipv6 address")
		}
		fullIp := addr.StringExpanded()
		ipsV6 := strings.Split(fullIp, ":")
		for i, seg := range ipsV6 {
			if len(seg) == 0 || len(seg) > 4 {
				continue
			}
			if i < len(ipsV6)-1 {
				buffer.WriteString(seg + ":")
			} else {
				buffer.WriteString(seg)
			}
		}
		return buffer.String(), nil
	}
	return "", errors.New("ip type or ip length is illegal")
}

// ip to string with fill zero: ip: 192.168.5.2 --> c0.a8.05.02
func IpStrToHexString(ip string, ipType int) (string, error) {
	var ipp net.IP
	if ipType == Ipv4Type {
		ipp = net.ParseIP(ip).To4()
	} else if ipType == Ipv6Type {
		ipp = net.ParseIP(ip).To16()
	} else {
		return "", errors.New("illegal ip type")
	}

	return IpNetToHexString(ipp, ipType)
}

// ip to string with fill zero: ip: 192.168.5.2 --> c0.a8.05.02
func IpNetToHexString(ip net.IP, ipType int) (string, error) {
	var buffer bytes.Buffer

	if ipType == Ipv4Type && len(ip) == net.IPv4len {
		for i := 0; i < net.IPv4len; i++ {
			if i < net.IPv4len-1 {
				buffer.WriteString(fmt.Sprintf("%02x.", ip[i]))
			} else {
				buffer.WriteString(fmt.Sprintf("%02x", ip[i]))
			}
		}
		return buffer.String(), nil
	} else if ipType == Ipv6Type && len(ip) == net.IPv6len {
		for i := 0; i < net.IPv6len; i = i + 2 {
			if i < net.IPv6len-2 {
				buffer.WriteString(fmt.Sprintf("%02x%02x:", ip[i], ip[i+1]))
			} else {
				buffer.WriteString(fmt.Sprintf("%02x%02x", ip[i], ip[i+1]))
			}
		}
		return buffer.String(), nil
	}
	return "", errors.New("ip type or ip length is illegal")
}

// 192.168.5/24 -->  192.168.5.0/24 --> [min: c0.a8.05.00  max: c0.a8.05.ff]
// 2803:d380/28 --> 2803:d380::/28 --> [min: 2803:d380:0000:0000:0000:0000:0000:0000  max: 2803:d38f:ffff:ffff:ffff:ffff:ffff:ffff]
func IsAddressPrefixRangeContains(parentAddressPrefix string, childAddressPrefix string) (isContain bool, err error) {
	parentIpType := GetIpType(parentAddressPrefix)
	childIpType := GetIpType(childAddressPrefix)
	if parentIpType != childIpType {
		return false, errors.New("parentIpType is different with childIpType")
	}
	parentMin, parentMax, err := AddressPrefixToHexRange(parentAddressPrefix, parentIpType)
	if err != nil {
		belogs.Error("IsAddressPrefixRangeContains(): AddressPrefixToHexRange parentAddressPrefix fail,  parentAddressPrefix:", parentAddressPrefix, err)
		return false, err
	}

	childMin, childMax, err := AddressPrefixToHexRange(childAddressPrefix, childIpType)
	if err != nil {
		belogs.Error("IsAddressPrefixRangeContains(): AddressPrefixToHexRange childAddressPrefix fail,  childAddressPrefix:", childAddressPrefix, err)
		return false, err
	}
	return (parentMin <= childMin && parentMax >= childMax), nil
}

// 192.168.5/24 -->  192.168.5.0/24 --> [min: c0.a8.05.00  max: c0.a8.05.ff]
// 2803:d380/28 --> 2803:d380::/28 --> [min: 2803:d380:0000:0000:0000:0000:0000:0000  max: 2803:d38f:ffff:ffff:ffff:ffff:ffff:ffff]
func AddressPrefixToHexRange(addressPrefix string, ipType int) (minHex string, maxHex string, err error) {
	belogs.Debug("AddressPrefixToHexRange(): addressPrefix:", addressPrefix, "  ipType:", ipType)
	addressPrefixFill, err := FillAddressPrefixWithZero(addressPrefix, ipType)
	if err != nil {
		belogs.Error("AddressPrefixToHexRange(): FillAddressPrefixWithZero fail,  addressPrefix:", addressPrefix, " ipTye :", ipType, err)
		return "", "", err
	}

	_, subnet, err := net.ParseCIDR(addressPrefixFill)
	if err != nil {
		belogs.Error("AddressPrefixToHexRange(): ParseCIDR fail, addressPrefix:", addressPrefix, " ipTye:", ipType,
			"  addressPrefixFill:", addressPrefixFill, err)
		return "", "", err
	}

	var ipLen int
	if ipType == Ipv4Type {
		ipLen = net.IPv4len
	} else if ipType == Ipv6Type {
		ipLen = net.IPv6len
	}

	min := make(net.IP, ipLen)
	max := make(net.IP, ipLen)
	for i := 0; i < ipLen; i++ {
		min[i] = subnet.IP[i] & subnet.Mask[i]
		max[i] = subnet.IP[i] | (^subnet.Mask[i])
	}

	minHex, err = IpNetToHexString(min, ipType)
	if err != nil {
		return "", "", err
	}
	maxHex, err = IpNetToHexString(max, ipType)
	if err != nil {
		return "", "", err
	}
	return minHex, maxHex, nil
}

// ipaddress is included in parent ipaddress;  .
// parentRangeStart, parentRangeEnd, selfRangeStart, selfRangeEnd,
func IpRangeIncludeInParentRange(parentRangeStart, parentRangeEnd, selfRangeStart, selfRangeEnd string) bool {
	if len(parentRangeStart) == 0 || len(selfRangeStart) == 0 ||
		len(selfRangeEnd) == 0 || len(parentRangeEnd) == 0 {
		return false
	}
	return parentRangeStart <= selfRangeStart && selfRangeEnd <= parentRangeEnd
}

// ipv4 to number
func Ipv4toInt(ip net.IP) int64 {
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(ip.To4())
	return IPv4Int.Int64()
}

func GetIpType(ip string) (ipType int) {
	// 修复：不再通过字符串包含":"判断，改用标准库解析
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		// 兼容旧逻辑：如果解析失败，仍通过":"判断（避免改变核心逻辑）
		if strings.Contains(ip, ":") {
			return Ipv6Type
		}
		return Ipv4Type
	}
	if addr.Is4() {
		return Ipv4Type
	} else if addr.Is6() {
		return Ipv6Type
	}
	return Ipv4Type // 默认返回IPv4Type，保持旧逻辑
}

// check is 192.168.5.1 or 2803:d380::
func IsAddress(ip string) bool {
	return nil != net.ParseIP(ip)
}

// check is: 192.168.5/24   or 2803:d380/28
func IsAddressPrefix(ip string) bool {
	if len(ip) == 0 || !strings.Contains(ip, "/") {
		return false
	}
	ipType := GetIpType(ip)

	network, err := FillAddressPrefixWithZero(ip, ipType)
	if err != nil {
		belogs.Error("IsAddressPrefix(): FillAddressPrefixWithZero err:", ip, err)
		return false
	}

	_, _, err = net.ParseCIDR(network)
	if err != nil {
		belogs.Error("IsAddressPrefix(): ParseCIDR err:", ip, "  ipType:", ipType, "   network:", network, err)
		return false
	}
	return true
}

func SplitAddressAndPrefix(addressPrefix string) (address string, prefix uint64, err error) {
	split := strings.Split(addressPrefix, "/")
	if len(addressPrefix) == 0 || len(split) != 2 {
		return "", 0, errors.New("ip address prefix format is illegal (must contain exactly one '/')")
	}
	address = split[0]
	pInt, err := strconv.Atoi(split[1])
	if err != nil {
		return "", 0, fmt.Errorf("prefix is not a valid integer: %w", err)
	}
	// 新增：校验prefix范围
	ipType := GetIpType(address)
	if (ipType == Ipv4Type && (pInt < 0 || pInt > 32)) || (ipType == Ipv6Type && (pInt < 0 || pInt > 128)) {
		return "", 0, errors.New("prefix out of valid range")
	}
	prefix = uint64(pInt)
	return address, prefix, nil
}

// ip to binary string with fill zero: ip:191.243.248.0 --> 10111111111100111111100000000000
func IpAddressToBinaryString(ip string) (string, error) {
	ipType := GetIpType(ip)
	ipp, err := FillAddressWithZero(ip, ipType)
	if err != nil {
		return "", err
	}
	return IpStrToBinaryString(ipp, ipType)
}

// ip to binary string with fill zero: ip:191.243.248.0 --> 10111111111100111111100000000000
func IpStrToBinaryString(ip string, ipType int) (string, error) {
	ipp := net.IP{}
	if ipType == Ipv4Type {
		ipp = net.ParseIP(ip).To4()
	} else if ipType == Ipv6Type {
		ipp = net.ParseIP(ip).To16()
	} else {
		return "", errors.New("illegal ip type")
	}

	return IpNetToBinaryString(ipp, ipType)
}

// ip to binary string with fill zero: ip:191.243.248.0 --> 10111111111100111111100000000000
func IpNetToBinaryString(ip net.IP, ipType int) (string, error) {
	var buffer bytes.Buffer

	if ipType == Ipv4Type && len(ip) == net.IPv4len {
		for i := 0; i < net.IPv4len; i++ {
			buffer.WriteString(fmt.Sprintf("%08b", ip[i]))
		}
		return buffer.String(), nil
	} else if ipType == Ipv6Type && len(ip) == net.IPv6len {
		for i := 0; i < net.IPv6len; i = i + 2 {
			buffer.WriteString(fmt.Sprintf("%08b%08b", ip[i], ip[i+1]))
		}
		return buffer.String(), nil
	}
	return "", errors.New("ip type or ip length is illegal")
}

// "10.1.0.0/16" --> 0a 01 (all is 03 03 00 0a 01)
func AddressPrefixToAsn1HexFormat(addressPrefix string) (Asn1HexFormat string, err error) {
	belogs.Debug("AddressPrefixToAsn1HexFormat(): addressPrefix:", addressPrefix)
	_, netCIDR, err := net.ParseCIDR(addressPrefix)
	if err != nil {
		belogs.Error("AddressPrefixToAsn1HexFormat(): ParseCIDR fail, addressPrefix:", addressPrefix, err)
		return "", err
	}

	ipNet := &asn1addressasn.IPNet{
		IPNet: netCIDR,
	}
	ipBytesAll, err := ipNet.ASN1()
	if err != nil {
		belogs.Error("AddressPrefixToAsn1HexFormat(): ASN1 fail, addressPrefix:", addressPrefix, "  ipNet:", ipNet, err)
		return "", err
	}
	// 修复：校验ipBytesAll长度，避免索引越界
	if len(ipBytesAll) <= 3 {
		return "", errors.New("asn1 encoded ip bytes is too short")
	}
	ipBytes := ipBytesAll[3:] // just ipaddress, no 0x00
	hexAddress := hex.EncodeToString(ipBytes)
	belogs.Debug("AddressPrefixToAsn1HexFormat(): ok, ipBytesAll:", convert.PrintBytesOneLine(ipBytesAll),
		"  addressPrefix:", addressPrefix, "  hexAddress:", hexAddress)
	return hexAddress, nil
}
