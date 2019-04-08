package ip

import (
	"bytes"
	"fmt"
	belogs "github.com/astaxie/beego/logs"
	"strconv"
	"strings"
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
	belogs.Debug("RtrFormatToIp():rtrIp: %+v:", rtrIp, "   len(rtrIp):", len(rtrIp))
	var ip string
	if len(rtrIp) == 4 {
		ip = fmt.Sprintf("%d.%d.%d.%d", rtrIp[0], rtrIp[1], rtrIp[2], rtrIp[3])
		belogs.Debug("RtrFormatToIp():ipv4:ip:", ip)
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
		belogs.Debug("RtrFormatToIp():ipv6:ip:", ip)
		return ip
	}
	belogs.Error("RtrFormatToIp():is not ipv4 or ipv6:", rtrIp)
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
