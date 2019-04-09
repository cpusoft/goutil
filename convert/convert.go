package convert

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
)

func Bytes2Uint64(bytes []byte) uint64 {
	lens := 8 - len(bytes)
	bb := make([]byte, lens)
	bb = append(bb, bytes...)
	return binary.BigEndian.Uint64(bb)
}

func Bytes2String(byt []byte) string {
	return hex.EncodeToString(byt)
}

func GetInterfaceType(v interface{}) (string, error) {
	switch v.(type) {
	case int:
		return "int", nil
	case string:
		return "string", nil
	case ([]byte):
		return "[]byte", nil
	default:
		return "unknown", nil
	}
}

func Interface2String(v interface{}) (string, error) {
	if str, ok := v.(string); ok {
		return str, nil
	} else {
		return "", errors.New("an interface{} cannot convert to a string")
	}
}
func Interface2Bytes(v interface{}) ([]byte, error) {
	if by, ok := v.([]byte); ok {
		return by, nil
	} else {
		return nil, errors.New("an interface{} cannot convert to []byte")
	}
}
func Interface2Uint64(v interface{}) (uint64, error) {
	if ui, ok := v.(uint64); ok {
		return ui, nil
	} else {
		return 0, errors.New("an interface{} cannot convert to a uint64")
	}
}
