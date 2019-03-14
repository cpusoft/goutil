package ip

import (
	"fmt"
	"strconv"
	"strings"

	belogs "github.com/astaxie/beego/logs"
)

func RtrFormatToIp(ips []byte) string {

	// ipv4
	belogs.Debug("RtrFormatToIp():ips: %+v:", ips)
	if len(ips) == 8 {
		ip0, _ := strconv.ParseInt(string(ips[0:2]), 16, 0)
		ip1, _ := strconv.ParseInt(string(ips[2:4]), 16, 0)
		ip2, _ := strconv.ParseInt(string(ips[4:6]), 16, 0)
		ip3, _ := strconv.ParseInt(string(ips[6:8]), 16, 0)
		return fmt.Sprintf("%d.%d.%d.%d", ip0, ip1, ip2, ip3)
	} else if len(ips) == 32 {
		ip0 := string(ips[0:4])
		ip1 := string(ips[4:8])
		ip2 := string(ips[8:12])
		ip3 := string(ips[12:16])
		return fmt.Sprintf("%s:%s:%s:%s", ip0, ip1, ip2, ip3)
	}

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
