package convert

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cpusoft/goutil/osutil"
)

func Bytes2Uint64(bytes []byte) uint64 {
	lens := 8 - len(bytes)
	bb := make([]byte, lens)
	bb = append(bb, bytes...)
	return binary.BigEndian.Uint64(bb)
}

//0102abc1
func Bytes2String(byt []byte) string {
	return hex.EncodeToString(byt)
}

// print bytes in section to show,  num==8
func PrintBytes(data []byte, num int) (ret string) {
	return Bytes2StringSection(data, num)
}

// print bytes in section to show
func Bytes2StringSection(data []byte, num int) (ret string) {
	var buffer bytes.Buffer
	for i, b := range data {
		buffer.WriteString(fmt.Sprintf("%02x ", b))
		if i > 0 && num > 0 && (i+1)%num == 0 {
			buffer.WriteString(osutil.GetNewLineSep())
		}
	}
	return buffer.String()
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
		var r = reflect.TypeOf(v)
		return fmt.Sprintf("%v", r), nil
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
	if v, p := a.(int64); p {
		return strconv.Itoa(int(v))
	}

	if v, p := a.(uint); p {
		return strconv.Itoa(int(v))
	}
	if v, p := a.(uint16); p {
		return strconv.Itoa(int(v))
	}
	if v, p := a.(uint32); p {
		return strconv.Itoa(int(v))
	}
	if v, p := a.(uint64); p {
		return strconv.Itoa(int(v))
	}

	if v, p := a.(float32); p {
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	if v, p := a.(float64); p {
		return strconv.FormatFloat(v, 'f', -1, 32)
	}
	if v, p := a.(bool); p {
		return strconv.FormatBool(v)
	}

	return ""
}

func Interface2String(v interface{}) (string, error) {
	if str, ok := v.(string); ok {
		return str, nil
	}
	return "", errors.New("an interface{} cannot convert to a string")
}

func Interface2Bytes(v interface{}) ([]byte, error) {
	if by, ok := v.([]byte); ok {
		return by, nil
	}
	return nil, errors.New("an interface{} cannot convert to []byte")
}

func Interface2Uint64(v interface{}) (uint64, error) {
	if ui, ok := v.(uint64); ok {
		return ui, nil
	}
	return 0, errors.New("an interface{} cannot convert to a uint64")
}
func Interface2Map(v interface{}) (map[string]string, error) {
	m := make(map[string]string, 0)
	if v, p := v.(map[string]string); p {
		for key, value := range v {
			m[key] = value
		}
		return m, nil
	}
	return m, errors.New("an interface{} cannot convert to a map")
}

func Interface2Time(v interface{}) (time.Time, error) {
	if by, ok := v.(time.Time); ok {
		return by, nil
	}
	return time.Now(), errors.New("an interface{} cannot convert to time.Time")
}

func Time2String(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02 15:04:05")
}

func Time2StringZone(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02 15:04:05 MST")
}

func String2Time(t string) (tm time.Time, e error) {
	if len(t) == 0 {
		return tm, errors.New("string is empty")
	}
	if strings.LastIndex(t, "z") >= 0 || strings.LastIndex(t, "Z") >= 0 {
		tm, e = time.Parse("2006-01-02 15:04:05Z", t)
	} else {
		tm, e = time.Parse("2006-01-02 15:04:05", t)
	}
	return tm.Local(), e
}

// struct --> map
func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	return data
}
