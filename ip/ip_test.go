package ip

import (
	"fmt"
	"testing"

	ip "."
)

func TestStandardFormatToDigital(t *testing.T) {
	// dig: 13635B00  str:19.99.91.0
	str := "19.99.91.0"
	di := ip.StandardFormatToDigital(str)
	fmt.Println(di)

	dig := "13635B00"
	fmt.Println(dig)

	str = "2001:DB8::"
	di = ip.StandardFormatToDigital(str)
	fmt.Println(di)

}

func TestDigitalFormatToStandard(t *testing.T) {
	dig := "13635B00"
	str := ip.DigitalFormatToStandard(dig)
	fmt.Println(str)

	dig = "20010DB8000000000000000000000000"
	str = ip.DigitalFormatToStandard(dig)
	fmt.Println(str)
}
