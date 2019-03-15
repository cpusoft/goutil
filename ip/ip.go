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
	if len(rtrIp) == 4 {
		b0, _ := strconv.ParseInt(string(rtrIp[0:1]), 16, 0)
		b1, _ := strconv.ParseInt(string(rtrIp[1:2]), 16, 0)
		b2, _ := strconv.ParseInt(string(rtrIp[2:3]), 16, 0)
		b3, _ := strconv.ParseInt(string(rtrIp[3:4]), 16, 0)
		ip = fmt.Sprintf("%d.%d.%d.%d", b0, b1, b2, b3)
		belogs.Debug("RtrFormatToIp():ipv4:ip:", ip)
		return ip
	} else if len(rtrIp) == 16 {
		var buffer bytes.Buffer
		buffer.Write(rtrIp[0:2])
		buffer.WriteString(":")
		buffer.Write(rtrIp[2:4])
		buffer.WriteString(":")
		buffer.Write(rtrIp[4:6])
		buffer.WriteString(":")
		buffer.Write(rtrIp[6:8])
		buffer.WriteString(":")
		buffer.Write(rtrIp[8:10])
		buffer.WriteString(":")
		buffer.Write(rtrIp[10:12])
		buffer.WriteString(":")
		buffer.Write(rtrIp[12:14])
		buffer.WriteString(":")
		buffer.Write(rtrIp[14:16])
		ip = (buffer.String())
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
