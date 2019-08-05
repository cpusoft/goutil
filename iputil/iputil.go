package ip

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"

	belogs "github.com/astaxie/beego/logs"
)

const (
	Ipv4Type = 0x01
	Ipv6Type = 0x02
)

func RoaFormtToIp(ans1Ip []byte, ipType int) string {
	belogs.Debug("RoaFormtToIp():ans1Ip: %+v:", ans1Ip, "  ipType:", ipType)
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

// 192.168.0.0/24-->192.168/24    192.168.1.0-->192.168.1
func TrimAddressPrefixZero(ip string, ipType int) (string, error) {
	if ipType == Ipv4Type {
		return strings.Replace(ip, ".0", "", -1), nil
	} else if ipType == Ipv6Type {
		// have no zero in ipv6
		return ip, nil
	} else {
		return "", errors.New("illegal ipType")
	}
}

// fill ip with zero:
// 192.168.1 --> 192.168.1.0;   192.168/24 --> 192.168.0.0/24
func FillAddressPrefixWithZero(ip string, ipType int) (string, error) {

	prefix := ""
	ipp := ip
	pos := strings.Index(ip, "/")
	if pos > 0 {
		prefix = string(ip[pos:])
		ipp = string(ip[:pos])
	}
	belogs.Debug("FillAddressPrefixWithZero():ip:", ip, "     ipType:", ipType, " --> ipp:", ipp, "   prefix:", prefix, "   pos:", pos)

	if ipType == Ipv4Type {
		countComma := strings.Count(ipp, ".")
		if countComma == 3 {
			return ipp + prefix, nil
		} else if countComma < 3 {
			return ipp + strings.Repeat(".0", net.IPv4len-countComma-1) + prefix, nil
		} else {
			return "", errors.New("illegal ipv4")
		}
	} else if ipType == Ipv6Type {
		countColon := strings.Count(ipp, ":")
		if countColon == 7 {
			return ipp + prefix, nil
		} else if strings.HasSuffix(ipp, "::") {
			return ipp + prefix, nil
		} else {
			return ipp + "::" + prefix, nil
		}

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
		belogs.Error("AddressPrefixToHexRange(): IpAndCIDRFillWithZero err:", err)
		return "", "", err
	}
	belogs.Debug("AddressPrefixToHexRange(): network:", network)

	_, subnet, err := net.ParseCIDR(network)
	if err != nil {
		belogs.Error("AddressPrefixToHexRange(): ParseCIDR err:", err)
		return "", "", err
	}
	belogs.Debug("AddressPrefixToHexRange(): subnet:", subnet)

	var ipLen int
	if ipType == Ipv4Type {
		ipLen = net.IPv4len
	} else if ipType == Ipv6Type {
		ipLen = net.IPv6len
	}
	belogs.Debug("AddressPrefixToHexRange(): ipLen:", ipLen)

	min := make(net.IP, ipLen)
	max := make(net.IP, ipLen)
	for i := 0; i < ipLen; i++ {
		min[i] = subnet.IP[i] & subnet.Mask[i]
		max[i] = subnet.IP[i] | (^subnet.Mask[i])
	}
	belogs.Debug("AddressPrefixToHexRange(): min:", min, " max:", max)

	minHex, err = IpNetToHexString(min, ipType)
	if err != nil {
		return "", "", err
	}
	maxHex, err = IpNetToHexString(max, ipType)
	if err != nil {
		return "", "", err
	}
	belogs.Debug("AddressPrefixToHexRange(): minHex:", minHex, " maxHex:", maxHex)
	return minHex, maxHex, nil
}

// ipv4 to number
func Ipv4toInt(ip net.IP) int64 {
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(ip.To4())
	return IPv4Int.Int64()
}

// check is: 192.168.5/24   or 2803:d380/28
func IsAddressPrefix(ip string) bool {
	if len(ip) == 0 || !strings.Contains(ip, "/") {
		return false
	}
	ipType := Ipv4Type
	if strings.Contains(ip, ":") {
		ipType = Ipv6Type
	}
	network, err := FillAddressPrefixWithZero(ip, ipType)
	if err != nil {
		belogs.Error("IsAddressPrefix(): IpAndCIDRFillWithZero err:", err)
		return false
	}
	belogs.Debug("IsAddressPrefix(): network:", network)

	_, subnet, err := net.ParseCIDR(network)
	if err != nil {
		belogs.Error("IsAddressPrefix(): ParseCIDR err:", err)
		return false
	}
	belogs.Debug("IsAddressPrefix(): subnet:", subnet)
	return true
}

func SplitAddressAndPrefix(addressPrefix string) (address string, prefix int, err error) {
	if len(addressPrefix) == 0 || len(strings.Split(addressPrefix, "/")) != 2 {
		return "", 0, errors.New("ip length or format is illegal")
	}
	split := strings.Split(addressPrefix, "/")
	address = split[0]
	prefix, err = strconv.Atoi(split[1])
	return address, prefix, err
}