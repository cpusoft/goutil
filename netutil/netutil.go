package netutil

import (
	"net"
	"net/url"
	"strings"

	"github.com/cpusoft/goutil/belogs"
)

// ip network

// just one ipaddr ;
//
//	addr.String()
func ResolveIp(host string) *net.IPAddr {
	// 修复点1：增加输入参数校验，避免空字符串触发无效系统调用
	if strings.TrimSpace(host) == "" {
		belogs.Error("ResolveIp(): input host is empty")
		return nil
	}

	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		// 修复点2：修正日志中的函数名，匹配当前函数
		belogs.Error("ResolveIp(): ResolveIPAddr fail, host:", host, "error:", err)
		return nil
	}
	return ipAddr
}

// []ipaddr
func LookupIp(host string) []net.IP {
	// 修复点1：增加输入参数校验
	if strings.TrimSpace(host) == "" {
		belogs.Error("LookupIp(): input host is empty")
		return nil
	}

	ipAddrs, err := net.LookupIP(host)
	if err != nil {
		// 修复点2：修正日志中的函数名，优化日志参数格式
		belogs.Error("LookupIp(): LookupIP fail, host:", host, "error:", err)
		return nil
	}
	return ipAddrs
}

func LookupIpByUrl(urlStr string) []net.IP { // 变量名改为urlStr，避免与包名冲突
	// 保留原输入参数校验
	if strings.TrimSpace(urlStr) == "" {
		belogs.Error("LookupIpByUrl(): input url is empty")
		return nil
	}

	// 修复点1：正确调用net/url.Parse，变量名避免与包名冲突
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		belogs.Error("LookupIpByUrl(): parse url fail, url:", urlStr, "error:", err)
		return nil
	}

	// 手动提取纯host（自动剥离端口）
	hostWithoutPort := parsedUrl.Hostname()
	if hostWithoutPort == "" {
		belogs.Error("LookupIpByUrl(): extract hostname from url fail, url:", urlStr)
		return nil
	}

	// 调用LookupIP解析纯host
	ipAddrs, err := net.LookupIP(hostWithoutPort)
	if err != nil {
		belogs.Error("LookupIpByUrl(): LookupIP fail, url:", urlStr, "processed host:", hostWithoutPort, "error:", err)
		return nil
	}
	return ipAddrs
}
