package convert

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strconv"
	"time"
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
	case map[string]string:
		return "map[string]string", nil
	default:
		return "unknown", nil
	}
}

func ToString(a interface{}) string {
	if v, p := a.(string); p {
		return v
	}
	if v, p := a.([]byte); p {
		return string(v)
	}
	if v, p := a.(int); p {
		return strconv.Itoa(v)
	}
	if v, p := a.(int16); p {
		return strconv.Itoa(int(v))
	}
	if v, p := a.(int32); p {
		return strconv.Itoa(int(v))
	}
	if v, p := a.(uint); p {
		return strconv.Itoa(int(v))
	}
	if v, p := a.(float32); p {
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	if v, p := a.(float64); p {
		return strconv.FormatFloat(v, 'f', -1, 32)
	}

	return ""
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
func Interface2Map(v interface{}) (map[string]string, error) {
	m := make(map[string]string, 0)
	if v, p := v.(map[string]string); p {
		for key, value := range v {
			m[key] = value
		}
		return m, nil
	} else {
		return m, errors.New("an interface{} cannot convert to a map")
	}
}

func Time2String(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04:05")
}
func String2Time(t string) (time.Time, error) {
	tm, e := time.Parse("2006-01-02 15:04:05", t)
	return tm.Local(), e
}
