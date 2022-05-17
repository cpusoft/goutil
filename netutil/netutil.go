package netutil

import (
	"net"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/urlutil"
)

// ip network

// just one ipaddr ;
//  addr.String()
func ResolveIp(host string) *net.IPAddr {
	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		belogs.Error("GetIpByHost():  ResolveIPAddr fail:", host, err)
		return nil
	}
	return ipAddr
}

// []ipaddr
func LookupIp(host string) []net.IP {
	ipAddrs, err := net.LookupIP(host)
	if err != nil {
		belogs.Error("GetIpByHost(): ResolveIPAddr fail:", host, err)
		return nil
	}
	return ipAddrs
}

func LookupIpByUrl(url string) []net.IP {
	host, err := urlutil.Host(url)
	if err != nil {
		belogs.Error("GetIpByHost():  Host fail:", url, err)
		return nil
	}
	ipAddrs, err := net.LookupIP(host)
	if err != nil {
		belogs.Error("GetIpByHost():  ResolveIPAddr fail:", url, host, err)
		return nil
	}
	return ipAddrs
}
