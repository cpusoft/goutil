package convert

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/osutil"
)

// Deprecated
func Bytes2Uint64(bytes []byte) uint64 {
	// 修复：处理长度超过8的情况，截断（取最后8字节）；长度不足8时补0到前面
	var bb []byte
	if len(bytes) > 8 {
		bb = bytes[len(bytes)-8:] // 超过8字节时取最后8个（符合uint64存储逻辑）
	} else {
		bb = make([]byte, 8-len(bytes)) // 补0到前面
		bb = append(bb, bytes...)
	}
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
	if n == nil {
		return nil, errors.New("n is nil")
	}
	bytesBuffer := bytes.NewBuffer([]byte{})
	var err error
	switch v := n.(type) {
	case int8:
		belogs.Debug("IntToBytes():int8 tmp:", v)
		err = binary.Write(bytesBuffer, binary.BigEndian, v)
	case uint8:
		belogs.Debug("IntToBytes():uint8 tmp:", v)
		err = binary.Write(bytesBuffer, binary.BigEndian, v)
	case int16:
		belogs.Debug("IntToBytes():int16 tmp:", v)
		err = binary.Write(bytesBuffer, binary.BigEndian, v)
	case uint16:
		belogs.Debug("IntToBytes():uint16 tmp:", v)
		err = binary.Write(bytesBuffer, binary.BigEndian, v)
	case int32:
		belogs.Debug("IntToBytes():int32 tmp:", v)
		err = binary.Write(bytesBuffer, binary.BigEndian, v)
	case uint32:
		belogs.Debug("IntToBytes():uint32 tmp:", v)
		err = binary.Write(bytesBuffer, binary.BigEndian, v)
	case int64:
		belogs.Debug("IntToBytes():int64 tmp:", v)
		err = binary.Write(bytesBuffer, binary.BigEndian, v)
	case uint64:
		belogs.Debug("IntToBytes():uint64 tmp:", v)
		err = binary.Write(bytesBuffer, binary.BigEndian, v)
	case int:
		tmp := int64(v)
		belogs.Debug("IntToBytes():int tmp:", tmp)
		err = binary.Write(bytesBuffer, binary.BigEndian, tmp)
	case uint:
		tmp := uint64(v)
		belogs.Debug("IntToBytes():uint tmp:", tmp)
		err = binary.Write(bytesBuffer, binary.BigEndian, tmp)
	default:
		return nil, errors.New("n is not digital")
	}
	if err != nil {
		return nil, err
	}
	return bytesBuffer.Bytes(), nil
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
	if data == nil {
		return ""
	}
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
	if v == nil {
		return "nil", nil
	}
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
	if a == nil {
		return ""
	}
	if v, p := a.(string); p {
		return v
	}
	if v, p := a.([]byte); p {
		return string(v)
	}

	// 修复：uint64 强转 int 溢出问题，改用 FormatUint
	switch v := a.(type) {
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.Itoa(int(v))
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64) // 修复：float64 应使用 64 位精度
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Local().Format("2006-01-02 15:04:05")
	case time.Duration:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", a)
	}
}

// when arg in args is nil, will ignore
func Interfaces2String(args ...interface{}) string {
	var buf strings.Builder
	for _, arg := range args {
		if arg == nil {
			continue
		}
		buf.WriteString(fmt.Sprintf("%v", arg) + " ")
	}
	return strings.TrimSpace(buf.String()) // 修复：去除末尾多余空格
}

func Interface2String(v interface{}) (string, error) {
	if str, ok := v.(string); ok {
		return str, nil
	}
	// 兼容 []byte 转 string（常见场景，不破坏原有返回值逻辑）
	if by, ok := v.([]byte); ok {
		return string(by), nil
	}
	return "", errors.New("an interface{} cannot convert to a string")
}

func Interface2Bytes(v interface{}) ([]byte, error) {
	if by, ok := v.([]byte); ok {
		return CloneBytes(by), nil // 修复：返回拷贝，避免原切片被修改
	}
	return nil, errors.New("an interface{} cannot convert to []byte")
}

func Interface2Uint64(v interface{}) (uint64, error) {
	if ui, ok := v.(uint64); ok {
		return ui, nil
	}
	// 兼容其他数值类型（不破坏原有返回值逻辑）
	switch val := v.(type) {
	case int:
		return uint64(val), nil
	case int64:
		return uint64(val), nil
	case uint:
		return uint64(val), nil
	case int32:
		return uint64(val), nil
	case uint32:
		return uint64(val), nil
	case int8:
		return uint64(val), nil
	case uint8:
		return uint64(val), nil
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
	return time.Time{}, errors.New("an interface{} cannot convert to time.Time") // 修复：返回零值而非当前时间
}

func String2Int(s string) (int, error) {
	if s == "" {
		return 0, errors.New("string is empty")
	}
	return strconv.Atoi(s)
}

func String2Uint64(s string) (uint64, error) {
	if s == "" {
		return 0, errors.New("string is empty")
	}
	return strconv.ParseUint(s, 10, 64)
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
	// 修复：兼容更多常见时间格式，避免解析失败
	formats := []string{
		"2006-01-02 15:04:05Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05 MST",
	}
	for _, format := range formats {
		tm, e = time.Parse(format, t)
		if e == nil {
			return tm.Local(), nil
		}
	}
	return tm, errors.New("unsupported time format: " + t)
}

// struct --> map
func Struct2Map(obj interface{}) map[string]interface{} {
	if obj == nil {
		return make(map[string]interface{})
	}
	// 修复：处理结构体指针类型
	val := reflect.ValueOf(obj)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return make(map[string]interface{})
	}
	t := val.Type()

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// 跳过未导出字段（避免 panic）
		if field.PkgPath != "" {
			continue
		}
		data[field.Name] = val.Field(i).Interface()
	}
	return data
}

func StringIsDigit(s string) bool {
	if s == "" {
		return false
	}
	reg := regexp.MustCompile("^[0-9]+$")
	return reg.MatchString(s)
}

func ByteIsDigit(b byte) bool {
	return ('0' <= b) && (b <= '9')
}

func CloneBytes(a []byte) []byte {
	if a == nil {
		return nil
	}
	b := make([]byte, len(a))
	copy(b, a)
	return b
}

func PrintBytesOneLine(data []byte) (ret string) {
	if data == nil {
		return ""
	}
	for _, b := range data {
		ret += fmt.Sprintf("%02x ", b)
	}
	return strings.TrimSpace(ret)
}

func MapKeysToSlice[K comparable, V any](m map[K]V) []K {
	if m == nil {
		return nil
	}
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func MapValuesToSlice[K comparable, V any](m map[K]V) []V {
	if m == nil {
		return nil
	}
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
