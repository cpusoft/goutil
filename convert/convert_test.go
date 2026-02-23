package convert

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"testing"
	"time"
)

/*
go test -bench=. -benchmem ./convert
*/

// -------------------------- 功能测试 & 临界值测试 --------------------------
func TestBytes2Uint64(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    uint64
		wantErr bool // 该函数无返回错误，仅用于标记场景
	}{
		{
			name:  "空字节切片",
			input: []byte{},
			want:  0,
		},
		{
			name:  "1字节（0x01）",
			input: []byte{0x01},
			want:  1,
		},
		{
			name:  "8字节（0x0102030405060708）",
			input: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			want:  0x0102030405060708,
		},
		{
			name:  "9字节（截断最后8字节）",
			input: []byte{0xFF, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			want:  0x0102030405060708,
		},
		{
			name:  "临界值（uint64最大值）",
			input: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			want:  math.MaxUint64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Bytes2Uint64(tt.input)
			if got != tt.want {
				t.Errorf("Bytes2Uint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestByteToDigit(t *testing.T) {
	tests := []struct {
		name  string
		input byte
		want  int
		ok    bool
	}{
		{name: "数字0", input: '0', want: 0, ok: true},
		{name: "数字9", input: '9', want: 9, ok: true},
		{name: "字母a", input: 'a', want: 0, ok: false},
		{name: "符号-", input: '-', want: 0, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ByteToDigit(tt.input)
			if got != tt.want || ok != tt.ok {
				t.Errorf("ByteToDigit() = (%v, %v), want (%v, %v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestDigitToByte(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  byte
		ok    bool
	}{
		{name: "数字0", input: 0, want: '0', ok: true},
		{name: "数字9", input: 9, want: '9', ok: true},
		{name: "数字10", input: 10, want: 0, ok: false},
		{name: "数字-1", input: -1, want: 0, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := DigitToByte(tt.input)
			if got != tt.want || ok != tt.ok {
				t.Errorf("DigitToByte() = (%v, %v), want (%v, %v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestBytesToBigInt(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  *big.Int
	}{
		{
			name:  "空字节",
			input: []byte{},
			want:  big.NewInt(0),
		},
		{
			name:  "单字节0x01",
			input: []byte{0x01},
			want:  big.NewInt(1),
		},
		{
			name:  "多字节0x0102",
			input: []byte{0x01, 0x02},
			want:  big.NewInt(258),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BytesToBigInt(tt.input)
			if got.Cmp(tt.want) != 0 {
				t.Errorf("BytesToBigInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBytesToInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    int64
		wantErr bool
	}{
		{
			name:    "空字节",
			input:   []byte{},
			want:    0,
			wantErr: false,
		},
		{
			name:    "int64最大值",
			input:   []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			want:    math.MaxInt64,
			wantErr: false,
		},
		{
			name:    "超过int64范围",
			input:   []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BytesToInt64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("BytesToInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BytesToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntToBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    []byte
		wantErr bool
	}{
		{name: "nil输入", input: nil, want: nil, wantErr: true},
		{name: "int8(127)", input: int8(127), want: []byte{0x7F}, wantErr: false},
		{name: "uint8(255)", input: uint8(255), want: []byte{0xFF}, wantErr: false},
		{name: "int16(32767)", input: int16(32767), want: []byte{0x7F, 0xFF}, wantErr: false},
		{name: "uint16(65535)", input: uint16(65535), want: []byte{0xFF, 0xFF}, wantErr: false},
		{name: "int32(2147483647)", input: int32(2147483647), want: []byte{0x7F, 0xFF, 0xFF, 0xFF}, wantErr: false},
		{name: "uint32(4294967295)", input: uint32(4294967295), want: []byte{0xFF, 0xFF, 0xFF, 0xFF}, wantErr: false},
		{name: "int64(9223372036854775807)", input: int64(math.MaxInt64), want: []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, wantErr: false},
		{name: "uint64(18446744073709551615)", input: uint64(math.MaxUint64), want: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, wantErr: false},
		{name: "int(100)", input: int(100), want: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64}, wantErr: false},
		{name: "uint(100)", input: uint(100), want: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64}, wantErr: false},
		{name: "非数字类型", input: "test", want: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IntToBytes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("IntToBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IntToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{name: "nil输入", input: nil, want: ""},
		{name: "字符串test", input: "test", want: "test"},
		{name: "字节切片[]byte{0x31,0x32}", input: []byte("12"), want: "12"},
		{name: "int(100)", input: 100, want: "100"},
		{name: "int8(-128)", input: int8(-128), want: "-128"},
		{name: "uint64最大值", input: uint64(math.MaxUint64), want: "18446744073709551615"},
		{name: "float64(3.14)", input: 3.14, want: "3.14"},
		{name: "bool(true)", input: true, want: "true"},
		{name: "time.Time", input: time.Date(2024, 1, 1, 12, 0, 0, 0, time.Local), want: "2024-01-01 12:00:00"},
		{name: "time.Duration", input: time.Second * 10, want: "10s"},
		{name: "结构体（默认格式）", input: struct{ Name string }{"test"}, want: "{test}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.input); got != tt.want {
				t.Errorf("ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterfaces2String(t *testing.T) {
	tests := []struct {
		name  string
		input []interface{}
		want  string
	}{
		{name: "空参数", input: nil, want: ""},
		{name: "包含nil参数", input: []interface{}{1, nil, "test"}, want: "1 test"},
		{name: "多类型参数", input: []interface{}{100, "abc", true}, want: "100 abc true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Interfaces2String(tt.input...); got != tt.want {
				t.Errorf("Interfaces2String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterface2String(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{name: "字符串类型", input: "test", want: "test", wantErr: false},
		{name: "字节切片类型", input: []byte("test"), want: "test", wantErr: false},
		{name: "int类型（不兼容）", input: 100, want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Interface2String(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Interface2String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Interface2String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterface2Bytes(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    []byte
		wantErr bool
	}{
		{name: "字节切片类型", input: []byte{0x01, 0x02}, want: []byte{0x01, 0x02}, wantErr: false},
		{name: "字符串类型（不兼容）", input: "test", want: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Interface2Bytes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Interface2Bytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Interface2Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterface2Uint64(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    uint64
		wantErr bool
	}{
		{name: "uint64类型", input: uint64(100), want: 100, wantErr: false},
		{name: "int类型", input: int(100), want: 100, wantErr: false},
		{name: "int64类型", input: int64(100), want: 100, wantErr: false},
		{name: "字符串类型（不兼容）", input: "100", want: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Interface2Uint64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Interface2Uint64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Interface2Uint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStruct2Map(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
		addr string // 未导出字段
	}
	tests := []struct {
		name  string
		input interface{}
		want  map[string]interface{}
	}{
		{
			name:  "nil输入",
			input: nil,
			want:  make(map[string]interface{}),
		},
		{
			name:  "结构体指针",
			input: &TestStruct{Name: "test", Age: 18, addr: "xxx"},
			want: map[string]interface{}{
				"Name": "test",
				"Age":  18,
			},
		},
		{
			name:  "非结构体类型",
			input: 100,
			want:  make(map[string]interface{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Struct2Map(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Struct2Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString2Time(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "空字符串", input: "", wantErr: true},
		{name: "标准格式", input: "2024-01-01 12:00:00", wantErr: false},
		{name: "带Z格式", input: "2024-01-01 12:00:00Z", wantErr: false},
		{name: "T分隔格式", input: "2024-01-01T12:00:00", wantErr: false},
		{name: "不支持的格式", input: "2024/01/01 12:00:00", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := String2Time(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("String2Time() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// -------------------------- 性能基准测试 --------------------------
func BenchmarkBytes2Uint64(b *testing.B) {
	// 测试数据：8字节（常规场景）、9字节（截断场景）、空字节（边界场景）
	testData := [][]byte{
		[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		[]byte{0xFF, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		[]byte{},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range testData {
			Bytes2Uint64(data)
		}
	}
}

func BenchmarkIntToBytes(b *testing.B) {
	// 覆盖不同数值类型
	testData := []interface{}{
		int8(127), uint8(255), int16(32767), uint16(65535),
		int32(2147483647), uint32(4294967295), int64(math.MaxInt64), uint64(math.MaxUint64),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range testData {
			IntToBytes(data)
		}
	}
}

func BenchmarkToString(b *testing.B) {
	// 高频场景：uint64、字符串、字节切片、时间
	testData := []interface{}{
		uint64(math.MaxUint64),
		"test string",
		[]byte("test bytes"),
		time.Now(),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range testData {
			ToString(data)
		}
	}
}

func BenchmarkStruct2Map(b *testing.B) {
	type LargeStruct struct {
		Field1 string
		Field2 int
		Field3 bool
		Field4 float64
		Field5 time.Time
	}
	testData := &LargeStruct{
		Field1: "test",
		Field2: 100,
		Field3: true,
		Field4: 3.14,
		Field5: time.Now(),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Struct2Map(testData)
	}
}

func BenchmarkCloneBytes(b *testing.B) {
	// 测试不同长度：小切片（10字节）、大切片（1024字节）
	smallData := make([]byte, 10)
	largeData := make([]byte, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CloneBytes(smallData)
		CloneBytes(largeData)
	}
}

func BenchmarkMapKeysToSlice(b *testing.B) {
	// 测试不同大小的map
	smallMap := make(map[int]string, 10)
	for i := 0; i < 10; i++ {
		smallMap[i] = strconv.Itoa(i)
	}
	largeMap := make(map[int]string, 1000)
	for i := 0; i < 1000; i++ {
		largeMap[i] = strconv.Itoa(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MapKeysToSlice(smallMap)
		MapKeysToSlice(largeMap)
	}
}

func TestIntToBytes1(t *testing.T) {
	i := int8(113)
	b, err := IntToBytes(i)
	fmt.Println(b, err)
	s := PrintBytes(b, 8)
	fmt.Println(s)
}

func TestInterface2Map(t *testing.T) {
	m := make(map[string]string, 0)
	m["aa"] = "11"
	m["bb"] = "bb"
	fmt.Println(m)

	var z interface{}
	z = m
	fmt.Println(z)

	m1, _ := Interface2Map(z)
	fmt.Println(m1)

	v1 := reflect.ValueOf(m1)
	fmt.Println("v1:", v1.Kind())

	bb := []byte{0x01, 0x02, 0x03}
	s, err := GetInterfaceType(bb)
	fmt.Println("type:", s, err)

}

func TestBytes2String(t *testing.T) {
	byt := []byte{0x01, 0x02, 0x03, 0xa1, 0xfc, 0x7c}
	s := Bytes2String(byt)
	fmt.Println(s)
}

func TestBytes2StringSection(t *testing.T) {
	byt := []byte{0x01, 0x02, 0x03, 0xa1, 0xfc, 0x7c, 0x01, 0x02, 0x03, 0xa1, 0xfc, 0x7c}
	s := Bytes2StringSection(byt, 8)
	fmt.Println(s)
}

type User struct {
	Id         int         `json:"id"`
	Username   string      `json:"username"`
	Password   string      `json:"password"`
	StateItems []StateItem `json:"stateItems"`
}

type StateItem struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func TestStruct2Map1(t *testing.T) {
	s1 := StateItem{"state", "valid"}
	s2 := StateItem{"error", "errorssss"}
	s3 := StateItem{"warning", "warningsss"}
	ss := make([]StateItem, 0)
	ss = append(ss, s1)
	ss = append(ss, s2)
	ss = append(ss, s3)
	user := User{Id: 5, Username: "zhangsan", Password: "password", StateItems: ss}

	data := Struct2Map(user)
	fmt.Printf("%v", data)
}

func TestTime2String(t *testing.T) {
	tt := time.Now()
	s := Time2String(tt)
	fmt.Println(s)

	s = Time2StringZone(tt)
	fmt.Println(s)
}

func TestByteToDigit1(t *testing.T) {
	b := []byte{0x00, 0x10}
	fmt.Println(b)
	bb := Bytes2Uint64(b)
	fmt.Println(bb)

	bb1 := BytesToBigInt(b)
	fmt.Println("BytesToBigInt:", bb1, bb1.Int64())

	var bbb byte
	bbb = 0x01
	bb1 = ByteToBigInt(bbb)
	fmt.Println("ByteToBigInt:", bb1, bb1.Int64())

}
