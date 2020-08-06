package bitutil

import ()

// 0x01(0000 0001) --> left 2 --> 0x(0000 0111)
// rfc3779 max : usine |, right bits are one
func LeftAndFillOne(bits uint8) (a uint8) {
	a = 1
	for i := uint8(0); i < bits; i++ {
		a = a | a<<1
	}
	return a
}

// 0x01(1111 1111) --> left 6 --> 0x(1100 0000)
// rfc3779 min : usine &, right bits are zero
func LeftAndFillZero(bits uint8) (a uint8) {
	a = 0xff
	for i := uint8(0); i < bits; i++ {
		a = a << 1
	}
	return a
}
