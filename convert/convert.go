package convert

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/osutil"
)

// Deprecated
func Bytes2Uint64(bytes []byte) uint64 {
	lens := 8 - len(bytes)
	bb := make([]byte, lens)
	bb = append(bb, bytes...)
	return binary.BigEndian.Uint64(bb)
}

// Deprecated
/*
func IntToBytes(n int) ([]byte, error) {
	data := int64(n)
	bytebuf := bytes.NewBuffer([]byte{})
	err := binary.Write(bytebuf, binary.BigEndian, data)
	if err != nil {
		return nil, err
	}
	return bytebuf.Bytes(), nil
}
*/

// Deprecated
func ByteToDigit(b byte) (digit int, ok bool) {
	if ByteIsDigit(b) {
		digit = int(b - '0')
		return digit, true
	}
	return 0, false
}

// Deprecated
func DigitToByte(digit int) (b byte, ok bool) {
	if (0 <= digit) && (digit <= 9) {
		b = byte('0' + digit)
		return b, true
	}
	return 0, false
}

func BytesToBigInt(bytes []byte) *big.Int {
	return big.NewInt(0).SetBytes(bytes)
}

func BytesToInt64(bytes []byte) (int64, error) {
	i := BytesToBigInt(bytes)
	if i != nil && i.IsInt64() {
		return i.Int64(), nil
	}
	return 0, errors.New("is not int")
}

func ByteToBigInt(b byte) *big.Int {
	bytes := make([]byte, 0)
	bytes = append(bytes, b)
	return BytesToBigInt(bytes)
}

// int8/int16/int32/int64 uint8/uint16/uint32/uint64
// int/uint as int64/uint64
// https://blog.csdn.net/whatday/article/details/97967180
func IntToBytes(n interface{}) ([]byte, error) {
	switch n.(type) {
	case int8:
		tmp := n.(int8)
		belogs.Debug("IntToBytes():int8 tmp:", tmp)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(), nil
	case uint8:
		tmp := n.(uint8)
		belogs.Debug("IntToBytes():uint8 tmp:", tmp)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(), nil
	case int16:
		tmp := n.(int16)
		belogs.Debug("IntToBytes():int16 tmp:", tmp)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(), nil
	case uint16:
		tmp := n.(uint16)
		belogs.Debug("IntToBytes():uint16 tmp:", tmp)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(), nil
	case int32:
		tmp := n.(int32)
		belogs.Debug("IntToBytes():int32 tmp:", tmp)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(), nil
	case uint32:
		tmp := n.(uint32)
		belogs.Debug("IntToBytes():uint32 tmp:", tmp)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(), nil
	case int64:
		tmp := n.(int64)
		belogs.Debug("IntToBytes():int64 tmp:", tmp)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(), nil
	case uint64:
		tmp := n.(uint64)
		belogs.Debug("IntToBytes():uint64 tmp:", tmp)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(), nil
	case int:
		tmp1 := n.(int)
		tmp := int64(tmp1)
		belogs.Debug("IntToBytes():int tmp:", tmp)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(), nil
	case uint:
		tmp1 := n.(uint)
		tmp := uint64(tmp1)
		belogs.Debug("IntToBytes():uint tmp:", tmp)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(), nil
	}
	return nil, errors.New("n is not digital")
}

// 0102abc1
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
	if v, p := a.(int8); p {
		return strconv.Itoa(int(v))
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
	if v, p := a.(uint8); p {
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
	if v, p := a.(time.Time); p {
		return v.Local().Format("2006-01-02 15:04:05")
	}
	if v, p := a.(time.Duration); p {
		return fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("%v", a)
}
func Interfaces2String(args ...interface{}) string {
	var buf strings.Builder
	for _, arg := range args {
		buf.WriteString(fmt.Sprintf("%v", arg) + " ")
	}
	return buf.String()
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

func ByteIsDigit(b byte) bool {
	return ('0' <= b) && (b <= '9')
}

func CloneBytes(a []byte) []byte {
	b := make([]byte, len(a))
	copy(b, a)
	return b
}

func PrintBytesOneLine(data []byte) (ret string) {
	for _, b := range data {
		ret += fmt.Sprintf("%02x ", b)
	}
	return strings.TrimSpace(ret)
}

func MapKeysToSlice[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func MapValuesToSlice[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
