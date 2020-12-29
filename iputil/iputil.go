package iputil

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	stringutil "github.com/cpusoft/goutil/stringutil"
)

const (
	Ipv4Type = 0x01
	Ipv6Type = 0x02
)

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
		return buffer.String()
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
		return buffer.String()
	}
	return ""
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

//Bad way, still need to find a good way
//19.99.91.0 --> []byte;     2001:DB8::-->[]byte
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
				belogs.Debug("IpToRtrFormatByte():tmp1 Atoi err:", tmp1, err)
				return nil
			}
			byt[bytIndx] = byte(bb)
			bytIndx++

			bb, err = strconv.ParseUint(tmp2, 16, 0)
			if err != nil {
				belogs.Debug("IpToRtrFormatByte():tmp2 Atoi err:", tmp2, err)
				return nil
			}
			byt[bytIndx] = byte(bb)
			bytIndx++
		}
		return byt
	}
	return nil
}

// 210.173.160 --> []byte{0xD2ADA000}
// 2803:d380 --> []byte{0x2803d3800**00}
func AddressToRtrFormatByte(address string) (ipHex []byte, ipType int, err error) {

	ipType = GetIpType(address)
	belogs.Debug("AddressToRtrFormatByte(): after GetIpType  :", ipType)

	addressFill, err := FillAddressWithZero(address, ipType)
	belogs.Debug("AddressToRtrFormatByte(): after FillAddressWithZero  :", addressFill, err)
	if err != nil {
		return nil, 0, err
	}

	if ipType == Ipv4Type {
		ipHex = net.ParseIP(addressFill).To4()
	} else if ipType == Ipv6Type {
		ipHex = net.ParseIP(addressFill).To16()
	}
	belogs.Debug("AddressPrefixToRtrFormatByte(): after ParseIP  :", ipHex)

	return ipHex, ipType, nil

}

// 192.168.0.0/24-->192.168/24    192.168.1.0-->192.168.1
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
		// have no zero in ipv6
		return ip, nil
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
		} else {
			addressFill = address + "::"
		}
		//belogs.Debug("FillAddressWithZero():ipv6  address-->addressFill :", address, addressFill)
		return addressFill, nil

	} else {
		return "", errors.New("illegal ipType")
	}

}

// ip to string with fill zero: ip: 192.168.5.2 --> c0.a8.05.02
func IpStrToHexString(ip string, ipType int) (string, error) {
	ipp := net.IP{}
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
func AddressPrefixToHexRange(ip string, ipType int) (minHex string, maxHex string, err error) {

	network, err := FillAddressPrefixWithZero(ip, ipType)
	if err != nil {
		belogs.Error("AddressPrefixToHexRange(): IpAndCIDRFillWithZero fail,  ip, ipTye :", ip, ipType, err)
		return "", "", err
	}

	_, subnet, err := net.ParseCIDR(network)
	if err != nil {
		belogs.Error("AddressPrefixToHexRange(): ParseCIDR fail,  ip, ipTye :", ip, ipType, err)
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
		belogs.Error("AddressPrefixToHexRange(): IpNetToHexString fail,  ip, ipTye, min :", ip, ipType, min, err)
		return "", "", err
	}
	maxHex, err = IpNetToHexString(max, ipType)
	if err != nil {
		belogs.Error("AddressPrefixToHexRange(): IpNetToHexString fail,  ip, ipTye, max :", ip, ipType, max, err)
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
		belogs.Error("IsAddressPrefix(): IpAndCIDRFillWithZero err:", ip, err)
		return false
	}
	//belogs.Debug("IsAddressPrefix(): network:", network)

	_, _, err = net.ParseCIDR(network)
	if err != nil {
		belogs.Error("IsAddressPrefix(): ParseCIDR err:", ip, err)
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
