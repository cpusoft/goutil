package bitutil

import (
	"fmt"
	"testing"
)

func TestXYZ(t *testing.T) {
	var z uint8 = 0x80
	a := LeftAndFillOne(3)
	fmt.Printf("%x,%d,%b\n", a, a, a)
	z = z | a
	fmt.Printf("%x,%d,%b\n", z, z, z)

	z = 0xff
	a = LeftAndFillZero(5)
	fmt.Printf("%x,%d,%b\n", a, a, a)
	z = z & a
	fmt.Printf("%x,%d,%b\n", z, z, z)
}
