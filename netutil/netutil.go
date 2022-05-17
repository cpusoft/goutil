package netutil

import (
	"net"

	"github.com/cpusoft/goutil/belogs"
)

// ip network

// just one ipaddr ;
//  addr.String()
func ResolveIp(host string) *net.IPAddr {
	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		belogs.Error("GetIpByHost():  ResolveIPAddr fail:", err)
		return nil
	}
	return ipAddr
}

// []ipaddr
func LookupIp(host string) []net.IP {
	ipAddrs, err := net.LookupIP(host)
	if err != nil {
		belogs.Error("GetIpByHost():  ResolveIPAddr fail:", err)
		return nil
	}
	return ipAddrs
}
