package ip

import (
	"fmt"
	"testing"

	ip "."
)

func TestIpToRtrFormat(t *testing.T) {
	// dig: 13635B00  str:19.99.91.0
	str := "19.99.91.0"
	di := ip.IpToRtrFormat(str)
	fmt.Println(di)

	dig := "13635B00"
	fmt.Println(dig)

	str = "2001:DB8::"
	di = ip.IpToRtrFormat(str)
	fmt.Println(di)

}

func TestRtrFormatToIp(t *testing.T) {
	dig := "13635B00"
	str := ip.RtrFormatToIp(dig)
	fmt.Println(str)

	dig = "20010DB8000000000000000000000000"
	str = ip.RtrFormatToIp(dig)
	fmt.Println(str)
}
