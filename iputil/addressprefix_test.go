package iputil

import (
	"fmt"
	"net"
	"testing"
)

func TestSummarize(t *testing.T) {
	pss := IpRangeToAddressPrefixRanges("194.193.128.0", "194.193.223.255")
	for _, p := range pss {
		fmt.Println(p)
	}

	ps := Summarize(net.ParseIP("2001:db8::1"), net.ParseIP("2001:db8::8000"))
	for _, p := range ps {
		fmt.Println(p)
	}
	/*
	   2001:db8::1/128
	   2001:db8::2/127
	   2001:db8::4/126
	   2001:db8::8/125
	   2001:db8::10/124
	   2001:db8::20/123
	   2001:db8::40/122
	   2001:db8::80/121
	   2001:db8::100/120
	   2001:db8::200/119
	   2001:db8::400/118
	   2001:db8::800/117
	   2001:db8::1000/116
	   2001:db8::2000/115
	   2001:db8::4000/114
	   2001:db8::8000/128
	*/
	ps = Summarize(net.ParseIP("194.193.128.0"), net.ParseIP("194.193.223.255"))
	for _, p := range ps {
		fmt.Println(p)
	}
	/*
		194.193.128.0/18
		194.193.192.0/19
	*/
	ps = Summarize(net.ParseIP("194.223.0.0"), net.ParseIP("194.223.95.255"))
	for _, p := range ps {
		fmt.Println(p)
	}
	/*
		194.223.0.0/18
		194.223.64.0/19
	*/
	ps = Summarize(net.ParseIP("2001:7fa:9::"), net.ParseIP("2001:7fa:e:ffff:ffff:ffff:ffff:ffff"))
	for _, p := range ps {
		fmt.Println(p)
	}
	/*
		2001:7fa:9::/48
		2001:7fa:a::/47
		2001:7fa:c::/47
		2001:7fa:e::/48
	*/

}

func TestContains(t *testing.T) {
	_, ipNet1, err := net.ParseCIDR("204.2.135.0/24")
	fmt.Println(ipNet1, err)

	_, ipNet2, err := net.ParseCIDR("204.2.226.240/28")
	fmt.Println(ipNet2, err)

	ipPrefix1 := Prefix{*ipNet1}
	fmt.Println(ipPrefix1)
	ipPrefix2 := Prefix{*ipNet2}
	fmt.Println(ipPrefix2)
}
