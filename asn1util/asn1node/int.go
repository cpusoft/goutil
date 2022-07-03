package asn1node

import (
	"encoding/binary"
	"reflect"

	"github.com/cpusoft/goutil/asn1util/asn1base"
)

var byteOrder = binary.BigEndian

func intBytesCrop(data []byte) []byte {

	if size := len(data); size > 0 {

		sign := data[0] & 0x80

		var b byte
		if sign != 0 {
			b = 0xFF
		}

		pos := 0
		for pos+1 < size {

			if data[pos] != b {
				break
			}

			if (data[pos+1] & 0x80) != sign {
				break
			}

			pos++
		}

		data = data[pos:]
	}

	return data
}

func intBytesComplete(data []byte, n int) []byte {

	if size := len(data); size < n {

		newData := make([]byte, n)

		var b byte
		if (data[0] & 0x80) != 0 {
			b = 0xFF
		}

		pos := 0
		for pos+size < n {
			newData[pos] = b
			pos++
		}

		copy(newData[pos:], data)
		data = newData
	}

	return data
}

func IntEncode(x int64) []byte {
	data := make([]byte, sizeOfUint64)
	byteOrder.PutUint64(data, uint64(x))
	return intBytesCrop(data)
}

func UintEncode(x uint64) []byte {
	data := make([]byte, sizeOfUint64+1)
	data[0] = 0
	byteOrder.PutUint64(data[1:], x)
	return intBytesCrop(data)
}

func IntDecode(data []byte) (int64, error) {
	//completedData := intBytesComplete(data, sizeOfUint64)
	//if len(completedData) == sizeOfUint64 {
	//	return int64(byteOrder.Uint64(completedData)), nil
	//}
	//return 0, ErrorUnmarshalBytes{data, reflect.Int}
	return asn1base.ParseInt64(data)
}

func UintDecode(data []byte) (uint64, error) {
	completedData := intBytesComplete(data, sizeOfUint64+1)
	if len(completedData) == sizeOfUint64+1 {
		if completedData[0] == 0 {
			return byteOrder.Uint64(completedData[1:]), nil
		}
	}
	return 0, ErrorUnmarshalBytes{data, reflect.Uint}
}
