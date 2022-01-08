package asn1parse

import (
	"errors"
	"io"

	"github.com/cpusoft/goutil/belogs"
)

// Length
func EncodeLength(data []byte, size int) ([]byte, error) {

	if size < 0 {
		belogs.Error("EncodeLength(): size < 0:", size)
		return data, errors.New("length is negative")
	}

	// Short form of length

	if size < 0x80 {
		b := byte(size)
		data = append(data, b)
		return data, nil
	}

	// Long form of length

	var k int
	x := size
	for x > 0 {
		k++
		x >>= 8
	}

	bs := make([]byte, 1+k)
	bs[0] = 0x80 | byte(k)

	x = size
	for k > 0 {
		bs[k] = byte(x & 0xFF)
		x >>= 8
		k--
	}

	data = append(data, bs...)

	return data, nil
}

func DecodeLength(data []byte, size *int) (rest []byte, err error) {

	if len(data) == 0 {
		return data, io.EOF
	}

	b := data[0]
	data = data[1:]

	// Short form of length

	if (b & 0x80) == 0 {
		*size = int(b)
		return data, nil
	}

	// Long form of length

	count := int(b & 0x7F)
	if (count < 1) || (8 < count) {
		belogs.Error("DecodeLength():fail count <1 or >8:", count)
		return data, errors.New("decode length is error")
	}

	var x int
	for i := 0; i < count; i++ {
		if len(data) == 0 {
			belogs.Error("DecodeLength():fail len(data) == 0:", count)
			return data, errors.New("Insufficient data length")
		}
		b := data[0]
		data = data[1:]

		x = (x << 8) | int(b)
	}

	*size = x

	return data, nil
}
