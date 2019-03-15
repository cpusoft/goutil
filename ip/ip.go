package ip

import (
	"fmt"
	"strconv"
	"strings"

	belogs "github.com/astaxie/beego/logs"
)

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
