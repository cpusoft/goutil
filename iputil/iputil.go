package iputil

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
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
	return err == nil && addr.Is4()
}

// IsIPv6 returns true iff the addr string represents an IPv6 address.
func IsIPv6(s string) bool {
	addr, err := netip.ParseAddr(s)
	return err == nil && addr.Is6()
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
		for i, ip := range ans1Ip {
			if i < len(ans1Ip)-1 {
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
		asn1IpTmp := ans1Ip
		if len(ans1Ip)%2 != 0 {
			// Insufficient digits, fill in 0
			asn1IpTmp = append(ans1Ip, 0x00)
		}
		for i := 0; i < len(asn1IpTmp); i = i + 2 {
			if i < len(asn1IpTmp)-2 {
				buffer.WriteString(fmt.Sprintf("%02x%02x:", asn1IpTmp[i], asn1IpTmp[i+1]))
			} else {
				buffer.WriteString(fmt.Sprintf("%02x%02x", asn1IpTmp[i], asn1IpTmp[i+1]))
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

	// ipv4
	//belogs.Debug("RtrFormatToIp():rtrIp: %+v:", rtrIp, "   len(rtrIp):", len(rtrIp))
	var ip string
	if len(rtrIp) == 4 {
		ip = fmt.Sprintf("%d.%d.%d.%d", rtrIp[0], rtrIp[1], rtrIp[2], rtrIp[3])
		//belogs.Debug("RtrFormatToIp():ipv4:ip:", ip)
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
		//belogs.Debug("RtrFormatToIp():ipv6:ip:", ip)
		return ip
	}
	//belogs.Error("RtrFormatToIp():is not ipv4 or ipv6:", rtrIp)
	return ""
}

func IpToRtrFormat(ip string) string {
	formatIp := ""

	// format  ipv4
	ipsV4 := strings.Split(ip, ".")
	if len(ipsV4) > 1 {
		for _, ipV4 := range ipsV4 {
			ip, _ := strconv.Atoi(ipV4)
			formatIp += fmt.Sprintf("%02x", ip)
		}
		return formatIp
	}

	// format ipv6
	count := strings.Count(ip, ":")
	if count > 0 {
		count := strings.Count(ip, ":")
		if count < 7 { // total colon is 8
			needCount := 7 - count + 2 //2 is current "::", need add
			colon := strings.Repeat(":", needCount)
			ip = strings.Replace(ip, "::", colon, -1)

		}
		ipsV6 := strings.Split(ip, ":")

		for _, ipV6 := range ipsV6 {
			formatIp += fmt.Sprintf("%04s", ipV6)
		}
		return formatIp
	}
	return ""
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
	//belogs.Error("RtrFormatToIp():is not ipv4 or ipv6:", rtrIp)
	return ""
}

func CompressFillIpv6(oldIp string) (newIp string) {

	// 8: 0000
	// 0000:0000:0000:0000:0000:0000:0000:0000 --> ::
	if oldIp == "0000:0000:0000:0000:0000:0000:0000:0000" {
		return "::"
	}
	zeros := []string{"0000:0000:0000:0000:0000:0000:0000",
		"0000:0000:0000:0000:0000:0000",
		"0000:0000:0000:0000:0000",
		"0000:0000:0000:0000",
		"0000:0000:0000",
		"0000:0000"}
	zeros1 := []string{"000", "00", "0"}
	zeros2 := []string{":000", ":00"}
	for i := range zeros {
		//belogs.Debug("CompressFillIpv6():oldIp:", oldIp, "   zeros[i]:", zeros[i])
		if strings.Contains(oldIp, zeros[i]) {
			// 2001:0000:0000:0000:0000:0000:0000:1 --> 2001::1
			// 2001:1:0000:0000:0000:0000:0000:0000 --> 2001:1::
			// 0000:0000:0000:0000:0000:0000:2001:1 --> ::2001:1
			newIp = strings.Replace(oldIp, zeros[i], ":", -1) //
			belogs.Debug("CompressFillIpv6():oldIp:", oldIp, "   zeros[i]:", zeros[i], "   newIp", newIp)
			if strings.HasPrefix(newIp, ":") {
				newIp = ":" + newIp
			} else if strings.HasSuffix(newIp, ":") {
				newIp = newIp + ":"
			}
			newIp = strings.Replace(newIp, ":::", "::", -1) //
			belogs.Debug("CompressFillIpv6(): HasPrefix or HasSuffix,newIp:", newIp)
			// 000*:****
			for j := range zeros1 {
				if strings.HasPrefix(newIp, zeros1[j]) {
					newIp = strings.Replace(newIp, zeros1[j], "", 1) // just one
					belogs.Debug("CompressFillIpv6(): newIp:", newIp, " zeros1[j]:", zeros1[j])
				}
			}
			// **:000*:** --> **:*:**
			// **:00**:** --> **:**:**
			for k := range zeros2 {
				newIp = strings.Replace(newIp, zeros2[k], ":", -1) //
				belogs.Debug("CompressFillIpv6(): newIp:", newIp, " zeros2[j]:", zeros2[k])
			}
			// **:0*:** --> **:*:**
			// but *:0:*, cannot replace ":0:"
			// 2001:0db8:00:00:1:: --> 2001:db8:0:0:1::
			for m := 0; m < 6; m++ {
				newIp = strings.Replace(newIp, ":0:", ":00:", -1) // *:0:* --> *:00:*
			}

			belogs.Debug("CompressFillIpv6(): Replace :0: to :00:", newIp)
			newIp = strings.Replace(newIp, ":0", ":", -1) // *:00:* --> *:0:* , *:0* --> *:*
			belogs.Debug("CompressFillIpv6(): Replace :0 to :", newIp)
			return newIp
		}
	}
	return oldIp
}

// Bad way, still need to find a good way
// 19.99.91.0 --> []byte;     2001:DB8::-->[]byte
func IpToRtrFormatByte(ip string) []byte {
	belogs.Debug("IpToRtrFormatByte():ip", ip)

	// format  ipv4
	ipsV4 := strings.Split(ip, ".")
	if len(ipsV4) > 1 {
		byt := make([]byte, 4)
		for i := range ipsV4 {
			tmp, err := strconv.Atoi(ipsV4[i])
			belogs.Debug("IpToRtrFormatByte():ipv6 Atoi i:", i, " ipsV4[i]:", ipsV4[i], "   tmp:", tmp)
			if err != nil {
				belogs.Debug("IpToRtrFormatByte():ipv4 Atoi err:", ipsV4[i], err)
				return nil
			}
			byt[i] = byte(tmp)
		}
		return byt
	}

	// format ipv6
	count := strings.Count(ip, ":")
	if count > 0 {
		count := strings.Count(ip, ":")
		if count < 7 { // total colon is 8
			needCount := 7 - count + 2 //2 is current "::", need add
			colon := strings.Repeat(":", needCount)
			ip = strings.Replace(ip, "::", colon, -1)
		}
		belogs.Debug("IpToRtrFormatByte():new ip", ip)

		ipsV6 := strings.Split(ip, ":")
		byt := make([]byte, 16)
		bytIndx := 0
		for i := range ipsV6 {
			tmpV6 := fmt.Sprintf("%04s", ipsV6[i])
			tmp1 := tmpV6[0:2]
			tmp2 := tmpV6[2:4]
			belogs.Debug("IpToRtrFormatByte():tmpV6:", tmpV6, "  tmp1:", tmp1, "  tmp2:", tmp2)

			bb, err := strconv.ParseUint(tmp1, 16, 0)
			if err != nil {
				belogs.Error("IpToRtrFormatByte():tmp1 ParseUint fail, tmp1:", tmp1, err)
				return nil
			}
			if bb > math.MaxUint16 {
				belogs.Error("IpToRtrFormatByte():tmp1 > math.MaxUint16 fail, tmp1:", tmp1)
				return nil
			} else {
				byt[bytIndx] = byte(bb)
				bytIndx++
			}

			bb, err = strconv.ParseUint(tmp2, 16, 0)
			if err != nil {
				belogs.Error("IpToRtrFormatByte():tmp2 ParseUint fail, tmp2:", tmp2, err)
				return nil
			}
			if bb > math.MaxUint16 {
				belogs.Error("IpToRtrFormatByte():tmp2 > math.MaxUint16 fail, tmp2:", tmp2)
				return nil
			} else {
				byt[bytIndx] = byte(bb)
				bytIndx++
			}
		}
		return byt
	}
	return nil
}

// 210.173.160 --> []byte{0xD2ADA000}
// 2803:d380 --> []byte{0x2803d3800**00}
func AddressToRtrFormatByte(address string) (ipHex []byte, ipType int, err error) {

	ipType = GetIpType(address)
	//belogs.Debug("AddressToRtrFormatByte(): after GetIpType  :", ipType)

	addressFill, err := FillAddressWithZero(address, ipType)
	//belogs.Debug("AddressToRtrFormatByte(): after FillAddressWithZero  :", addressFill, err)
	if err != nil {
		return nil, 0, err
	}

	if ipType == Ipv4Type {
		ipHex = net.ParseIP(addressFill).To4()
	} else if ipType == Ipv6Type {
		ipHex = net.ParseIP(addressFill).To16()
	}
	//belogs.Debug("AddressPrefixToRtrFormatByte(): after ParseIP  :", ipHex)

	return ipHex, ipType, nil

}

// 192.168.0.0/24--> 192.168/24    192.168.1.0-->192.168.1
//
//	2c0f:ea60::/32 --> 2c0f:ea60/32
func TrimAddressPrefixZero(ip string, ipType int) (string, error) {
	if ipType == Ipv4Type {
		split := strings.Split(ip, "/")
		if len(split) == 1 {
			return stringutil.TrimeSuffixAll(ip, ".0"), nil
		} else if len(split) == 2 {
			return stringutil.TrimeSuffixAll(split[0], ".0") + "/" + split[1], nil
		} else {
			return "", errors.New("illegal address prefix")
		}

	} else if ipType == Ipv6Type {
		split := strings.Split(ip, "/")
		if len(split) == 1 {
			return stringutil.TrimeSuffixAll(ip, "::"), nil
		} else if len(split) == 2 {
			return stringutil.TrimeSuffixAll(split[0], "::") + "/" + split[1], nil
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
		if countComma == 3 {
			addressFill = address
		} else if countComma < 3 {
			addressFill = address + strings.Repeat(".0", net.IPv4len-countComma-1)
		}
		//belogs.Debug("FillAddressWithZero():ipv4  address-->addressFill :", address, addressFill)
		return addressFill, nil
	} else if ipType == Ipv6Type {
		countColon := strings.Count(address, ":")
		if countColon == 7 {
			addressFill = address
		} else if strings.HasSuffix(address, "::") {
			addressFill = address
		} else if strings.Contains(address, "::") {
			addressFill = address
		} else {
			addressFill = address + "::"
		}
		//belogs.Debug("FillAddressWithZero():ipv6  address-->addressFill :", address, addressFill)
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
		for i := range split {
			if len(split[i]) == 0 {
				continue
			}
			label, err := strconv.Atoi(split[i])
			if err != nil {
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
		split := strings.Split(ip, ":")
		for i := range split {
			var s string
			if len(split[i]) == 0 || len(split[i]) > 4 {
				continue
			} else if len(split[i]) == 1 {
				s = "000" + split[i]
			} else if len(split[i]) == 2 {
				s = "00" + split[i]
			} else if len(split[i]) == 3 {
				s = "0" + split[i]
			} else if len(split[i]) == 4 {
				s = split[i]
			}
			belogs.Debug("IpStrIncompleteToHexString(): split[i]:", split[i], " s:", s)
			buffer.WriteString(s + ":")
		}
		b := buffer.String()
		return string(b[:len(b)-1]), nil
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
		//belogs.Debug("IsAddressPrefixRangeContains(): parentIpType is different with childIpType, fail,  parentAddressPrefix:", parentAddressPrefix, " parentIpType :", parentIpType,
		//	"   childAddressPrefix:", childAddressPrefix, "   childIpType:", childIpType)
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
	//belogs.Debug("IsAddressPrefixRangeContains(): parentMin:", parentMin, "   parentMax:", parentMax, "   childMin:", childMin, " childMax:", childMax)
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
	belogs.Debug("AddressPrefixToHexRange(): addressPrefix:", addressPrefix, "  ipType:", ipType, "  addressPrefixFill:", addressPrefixFill)

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
		belogs.Error("AddressPrefixToHexRange(): IpNetToHexString fail,  addressPrefix:", addressPrefix, " ipTye:", ipType, "   min:", min, err)
		return "", "", err
	}
	maxHex, err = IpNetToHexString(max, ipType)
	if err != nil {
		belogs.Error("AddressPrefixToHexRange(): IpNetToHexString fail,  addressPrefix:", addressPrefix, " ipTye:", ipType, "   max:", max, err)
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

	// parent.RangeStart <--- c.RangeStart <---------> c.RangeEnd ---> parent.RangeEnd
	if parentRangeStart <= selfRangeStart && selfRangeEnd <= parentRangeEnd {
		return true
	}
	return false
}

// ipv4 to number
func Ipv4toInt(ip net.IP) int64 {
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(ip.To4())
	return IPv4Int.Int64()
}

func GetIpType(ip string) (ipType int) {
	ipType = Ipv4Type
	if strings.Contains(ip, ":") {
		ipType = Ipv6Type
	}
	return ipType
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
	//belogs.Debug("IsAddressPrefix(): network:", network)

	_, _, err = net.ParseCIDR(network)
	if err != nil {
		belogs.Error("IsAddressPrefix(): ParseCIDR err:", ip, "  ipType:", ipType, "   network:", network, err)
		return false
	}
	//belogs.Debug("IsAddressPrefix(): subnet:", subnet)
	return true
}

func SplitAddressAndPrefix(addressPrefix string) (address string, prefix uint64, err error) {
	if len(addressPrefix) == 0 || len(strings.Split(addressPrefix, "/")) != 2 {
		return "", 0, errors.New("ip length or format is illegal")
	}
	split := strings.Split(addressPrefix, "/")
	address = split[0]
	p, err := strconv.Atoi(split[1])
	prefix = uint64(p)
	return address, prefix, err
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
	belogs.Debug("AddressPrefixToAsn1HexFormat(): ipNet:", ipNet, "  ipBytesAll:", convert.PrintBytesOneLine(ipBytesAll))
	ipBytes := ipBytesAll[3:] // just ipaddress, no 0x00
	hexAddress := hex.EncodeToString(ipBytes)
	belogs.Debug("AddressPrefixToAsn1HexFormat(): ok, ipBytesAll:", convert.PrintBytesOneLine(ipBytesAll),
		"  addressPrefix:", addressPrefix, "  hexAddress:", hexAddress)
	return hexAddress, nil
}
