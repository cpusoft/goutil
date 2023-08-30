package bitutil

import (
	"fmt"
	"testing"
)

func TestXYZ(t *testing.T) {
	var z uint8 = 0x80
	a := Shift0x00LeftFillOne(3)
	fmt.Printf("%02x,%d,%b\n", a, a, a)
	z = z | a
	fmt.Printf("%02x,%d,%b\n", z, z, z)

	z = 0xff
	a = Shift0xffLeftFillZero(5)
	fmt.Printf("%02x,%d,%b\n", a, a, a)
	z = z & a
	fmt.Printf("%02x,%d,%b\n", z, z, z)

	newB := Shift0x00LeftFillOne(2)
	fmt.Println(newB)

	newB = Shift0xffLeftFillZero(2)
	fmt.Println(newB)
}
