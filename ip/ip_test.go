package ip

import (
	"fmt"
	"testing"

	ip "."
)

func TestRoaFormtToIp(t *testing.T) {
	b := []byte{0xb0, 0x10}
	di := ip.RoaFormtToIp(b, 0x01)
	fmt.Println(di)

	b = []byte{0x03, 0x05, 0x00, 0x28, 0x03, 0xEA, 0x80}
	di = ip.RoaFormtToIp(b, 0x02)
	fmt.Println(di)

	b = []byte{0x2a, 0x0, 0x15, 0x28, 0xaa, 0x0, 0xd0}
	di = ip.RoaFormtToIp(b, 0x02)
	fmt.Println(di)
}

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
	dig := []byte{80, 128, 0, 0}
	fmt.Println(len([]byte(dig)))
	str := ip.RtrFormatToIp(dig)
	fmt.Println(str)

	dig = []byte{32, 1, 7, 248, 0, 25, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	str = ip.RtrFormatToIp(dig)
	fmt.Println(str)
}
