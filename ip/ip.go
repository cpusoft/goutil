package ip

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	belogs "github.com/astaxie/beego/logs"
)

func RtrFormatToIp(rtrIp []byte) string {

	// ipv4
	belogs.Debug("RtrFormatToIp():rtrIp: %+v:", rtrIp, "   len(rtrIp):", len(rtrIp))
	var ip string
	if len(rtrIp) == 8 {
		b0, _ := strconv.ParseInt(string(rtrIp[0:2]), 16, 0)
		b1, _ := strconv.ParseInt(string(rtrIp[2:4]), 16, 0)
		b2, _ := strconv.ParseInt(string(rtrIp[4:6]), 16, 0)
		b3, _ := strconv.ParseInt(string(rtrIp[6:8]), 16, 0)
		ip = fmt.Sprintf("%d.%d.%d.%d", b0, b1, b2, b3)
		belogs.Debug("RtrFormatToIp():ipv4:ip:", ip)
		return ip
	} else if len(rtrIp) == 32 {
		var buffer bytes.Buffer
		buffer.Write(rtrIp[0:4])
		buffer.WriteString(":")
		buffer.Write(rtrIp[4:8])
		buffer.WriteString(":")
		buffer.Write(rtrIp[8:12])
		buffer.WriteString(":")
		buffer.Write(rtrIp[12:16])
		buffer.WriteString(":")
		buffer.Write(rtrIp[16:20])
		buffer.WriteString(":")
		buffer.Write(rtrIp[20:24])
		buffer.WriteString(":")
		buffer.Write(rtrIp[24:28])
		buffer.WriteString(":")
		buffer.Write(rtrIp[28:32])
		ip = (buffer.String())
		belogs.Debug("RtrFormatToIp():ipv6:ip:", ip)
		return ip
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
