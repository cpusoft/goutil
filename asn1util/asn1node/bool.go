package asn1node

import (
	"github.com/cpusoft/goutil/asn1util/asn1base"
)

func BoolEncode(x bool) []byte {
	data := []byte{0}
	if x {
		data[0] = 0xFF
	}
	return data
}

func BoolDecode(data []byte) (bool, error) {
	//	if len(data) != 1 {
	//		return false, ErrorUnmarshalBytes{data, reflect.Bool}
	//	}
	//	return (data[0] != 0), nil
	return asn1base.ParseBool(data)
}
