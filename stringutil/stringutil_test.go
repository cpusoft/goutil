package stringutil

import (
	"fmt"
	"strings"
	"testing"
)

func TestTrimSpaceAndNewLine1(t *testing.T) {
	s := `MIIL1gYJKoZIhvcNAQcCoIILxzC   
	CC8MCAQMxDTALBglghkgBZQMEAgEwggMXBgsqhkiG9w0BCRABGqCCAwYEggMCMIIC/gICBd0YDzIwMjMxMjIwMTkyOTAwWhgPMjA
	yMzEyMjExOTM0MDBaBglghkgBZQMEAgEwggLJMGcWQjMyMzYzMDMyM2E2NjY1NjQ2MTNhNjQzNTM4M2EzYTJmMzQzODJkMzQzODI
	wM2QzZTIwMzEzNDMxMzczMTMyLnJvYQMhAFicfryQK0kBExiMpBKZBiSpDqGqYm3INk9U4NhBqBQgMFEWLDczRENFRUMyNUIzM0U
	xQjAwREYyOEQ3RDczQkFCNkQwOEI5Q0FGRkMuY3JsAyEA2gvpQ8nuGDkFNZ1pkVJOYWMk1VliGGHlcgzqdBCVZq0wZxZCMzIzNjM
	   2q0eTI87VZv4CfABTjQ==`
	s1 := strings.Replace(s, "\r", "", -1)
	s1 = strings.Replace(s1, "\n", "", -1)
	s1 = strings.Replace(s1, "\t", "", -1)
	s1 = strings.Replace(s1, " ", "", -1)
	fmt.Println("s1:", s1)

	s2 := TrimSpaceAndNewLine(s)
	fmt.Println("s2:", s2)
}
func TestTrimeSuffixAll(t *testing.T) {
	ips := []string{"16.70.0.0", "16.0.1.0"}

	for _, ip := range ips {
		str := TrimSuffixAll(ip, ".0")
		fmt.Println(ip, " --> ", str)

	}
}

func TestGetValueFromJointStr1(t *testing.T) {
	str := `a=111&b=222&c=333`
	v := GetValueFromJointStr(str, "a", "&")
	fmt.Println(v)
	v = GetValueFromJointStr(str, "b", "&")
	fmt.Println(v)
	v = GetValueFromJointStr(str, "c", "&")
	fmt.Println(v)
}

func TestOmitString1(t *testing.T) {
	str := `0123456789a`
	str1 := OmitString(str, 0)
	fmt.Println(str1)
	str1 = OmitString(str, 1)
	fmt.Println(str1)
	str1 = OmitString(str, 9)
	fmt.Println(str1)
	str1 = OmitString(str, 10)
	fmt.Println(str1)
	str1 = OmitString(str, 11)
	fmt.Println(str1)
	str1 = OmitString(str, 12)
	fmt.Println(str1)
}
func TestStringsToInString1(t *testing.T) {
	ips := []string{"16.70.0.0", "16.0.1.0"}
	s := StringsToInString(ips)
	fmt.Println(s)
}

// -------------------------- 功能测试 --------------------------

// TestContainInSlice 测试 ContainInSlice 函数
func TestContainInSlice(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		one   string
		want  bool
	}{
		{
			name:  "正常匹配",
			slice: []string{"a", "b", "c"},
			one:   "b",
			want:  true,
		},
		{
			name:  "不匹配",
			slice: []string{"a", "b", "c"},
			one:   "d",
			want:  false,
		},
		{
			name:  "空切片",
			slice: []string{},
			one:   "a",
			want:  false,
		},
		{
			name:  "空匹配字符串",
			slice: []string{"a", "b"},
			one:   "",
			want:  false,
		},
		{
			name:  "切片含空字符串",
			slice: []string{"", "b"},
			one:   "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainInSlice(tt.slice, tt.one); got != tt.want {
				t.Errorf("ContainInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTrimNewLine 测试 TrimNewLine 函数
func TestTrimNewLine(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "含\\r\\n",
			str:  "a\r\nb\nc\r",
			want: "abc",
		},
		{
			name: "仅换行符",
			str:  "\r\n\r\n",
			want: "",
		},
		{
			name: "空字符串",
			str:  "",
			want: "",
		},
		{
			name: "无换行符",
			str:  "test",
			want: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimNewLine(tt.str); got != tt.want {
				t.Errorf("TrimNewLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestTrimSpace 测试 TrimSpace 函数
func TestTrimSpace(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "含空格和制表符",
			str:  "a \tb\t c ",
			want: "abc",
		},
		{
			name: "仅空格制表符",
			str:  " \t \t ",
			want: "",
		},
		{
			name: "空字符串",
			str:  "",
			want: "",
		},
		{
			name: "无空格制表符",
			str:  "test",
			want: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimSpace(tt.str); got != tt.want {
				t.Errorf("TrimSpace() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestTrimSpaceAndNewLine 测试 TrimSpaceAndNewLine 函数
func TestTrimSpaceAndNewLine(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "含空格、制表符、换行符",
			str:  "a \tb\r\nc\t \r",
			want: "abc",
		},
		{
			name: "混合空白字符",
			str:  "\r\n \t \r\n",
			want: "",
		},
		{
			name: "空字符串",
			str:  "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimSpaceAndNewLine(tt.str); got != tt.want {
				t.Errorf("TrimSpaceAndNewLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestTrimSuffixAll 测试 TrimSuffixAll 函数
func TestTrimSuffixAll(t *testing.T) {
	tests := []struct {
		name string
		str  string
		trim string
		want string
	}{
		{
			name: "移除多个后缀",
			str:  "test///",
			trim: "/",
			want: "test",
		},
		{
			name: "移除单个后缀",
			str:  "test/",
			trim: "/",
			want: "test",
		},
		{
			name: "无匹配后缀",
			str:  "test",
			trim: "/",
			want: "test",
		},
		{
			name: "空原字符串",
			str:  "",
			trim: "/",
			want: "",
		},
		{
			name: "空trim字符串",
			str:  "test///",
			trim: "",
			want: "test///",
		},
		{
			name: "多字符后缀",
			str:  "testabcabc",
			trim: "abc",
			want: "test",
		},
		{
			name: "边界：trim长度等于原字符串",
			str:  "abc",
			trim: "abc",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimSuffixAll(tt.str, tt.trim); got != tt.want {
				t.Errorf("TrimSuffixAll(%q, %q) = %q, want %q", tt.str, tt.trim, got, tt.want)
			}
		})
	}
}

// TestGetValueFromJointStr 测试 GetValueFromJointStr 函数
func TestGetValueFromJointStr(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		key       string
		separator string
		want      string
	}{
		{
			name:      "正常提取",
			line:      "a=1&b=2&c=3",
			key:       "b",
			separator: "&",
			want:      "2",
		},
		{
			name:      "key在开头",
			line:      "a=1&b=2",
			key:       "a",
			separator: "&",
			want:      "1",
		},
		{
			name:      "key在结尾",
			line:      "a=1&b=2",
			key:       "b",
			separator: "&",
			want:      "2",
		},
		{
			name:      "无匹配key",
			line:      "a=1&b=2",
			key:       "c",
			separator: "&",
			want:      "",
		},
		{
			name:      "空line",
			line:      "",
			key:       "a",
			separator: "&",
			want:      "",
		},
		{
			name:      "空key",
			line:      "a=1&b=2",
			key:       "",
			separator: "&",
			want:      "",
		},
		{
			name:      "空separator",
			line:      "a=1&b=2",
			key:       "a",
			separator: "",
			want:      "",
		},
		{
			name:      "separator非&",
			line:      "a=1;b=2;c=3",
			key:       "c",
			separator: ";",
			want:      "3",
		},
		{
			name:      "值含等号",
			line:      "a=1=2&b=3",
			key:       "a",
			separator: "&",
			want:      "1=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetValueFromJointStr(tt.line, tt.key, tt.separator); got != tt.want {
				t.Errorf("GetValueFromJointStr(%q, %q, %q) = %q, want %q", tt.line, tt.key, tt.separator, got, tt.want)
			}
		})
	}
}

// TestOmitString 测试 OmitString 函数
func TestOmitString(t *testing.T) {
	tests := []struct {
		name string
		str  string
		end  uint64
		want string
	}{
		{
			name: "end小于字符串长度",
			str:  "1234567890",
			end:  5,
			want: "12345",
		},
		{
			name: "end等于字符串长度",
			str:  "12345",
			end:  5,
			want: "12345",
		},
		{
			name: "end大于字符串长度",
			str:  "12345",
			end:  10,
			want: "12345",
		},
		{
			name: "空字符串",
			str:  "",
			end:  5,
			want: "",
		},
		{
			name: "end为0",
			str:  "12345",
			end:  0,
			want: "",
		},
		{
			name: "边界：end为int最大值",
			str:  "test",
			end:  uint64(int(^uint(0) >> 1)),
			want: "test",
		},
		{
			name: "边界：end超过int最大值",
			str:  "1234567890",
			end:  uint64(int(^uint(0)>>1)) + 100,
			want: "1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OmitString(tt.str, tt.end); got != tt.want {
				t.Errorf("OmitString(%q, %d) = %q, want %q", tt.str, tt.end, got, tt.want)
			}
		})
	}
}

// TestInt64sToInString 测试 Int64sToInString 函数
func TestInt64sToInString(t *testing.T) {
	tests := []struct {
		name string
		s    []int64
		want string
	}{
		{
			name: "正常切片",
			s:    []int64{1, 2, 3},
			want: "(1,2,3)",
		},
		{
			name: "空切片",
			s:    []int64{},
			want: "",
		},
		{
			name: "单元素切片",
			s:    []int64{100},
			want: "(100)",
		},
		{
			name: "含负数",
			s:    []int64{-1, 0, 1},
			want: "(-1,0,1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int64sToInString(tt.s); got != tt.want {
				t.Errorf("Int64sToInString() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestStringsToInString 测试 StringsToInString 函数
func TestStringsToInString(t *testing.T) {
	tests := []struct {
		name string
		s    []string
		want string
	}{
		{
			name: "正常切片",
			s:    []string{"a", "b", "c"},
			want: "(\"a\",\"b\",\"c\")",
		},
		{
			name: "空切片",
			s:    []string{},
			want: "",
		},
		{
			name: "单元素切片",
			s:    []string{"test"},
			want: "(\"test\")",
		},
		{
			name: "含空字符串",
			s:    []string{"", "b"},
			want: "(\"\",\"b\")",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringsToInString(tt.s); got != tt.want {
				t.Errorf("StringsToInString() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestStringsToInSqlString 测试 StringsToInSqlString 函数（重点测试SQL注入防护）
func TestStringsToInSqlString(t *testing.T) {
	tests := []struct {
		name string
		s    []string
		want string
	}{
		{
			name: "正常切片",
			s:    []string{"a", "b", "c"},
			want: "('a','b','c')",
		},
		{
			name: "空切片",
			s:    []string{},
			want: "",
		},
		{
			name: "单元素切片",
			s:    []string{"test"},
			want: "('test')",
		},
		{
			name: "含单引号（SQL注入防护）",
			s:    []string{"user'123", "b' OR '1'='1"},
			want: "('user''123','b'' OR ''1''=''1')",
		},
		{
			name: "含空字符串",
			s:    []string{"", "b"},
			want: "('','b')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringsToInSqlString(tt.s); got != tt.want {
				t.Errorf("StringsToInSqlString() = %q, want %q", got, tt.want)
			}
		})
	}
}

// -------------------------- 性能测试 --------------------------

// BenchmarkContainInSlice 性能测试 ContainInSlice
func BenchmarkContainInSlice(b *testing.B) {
	slice := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	// 预热
	_ = ContainInSlice(slice, "j")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ContainInSlice(slice, "j") // 最坏情况：匹配最后一个元素
	}
}

// BenchmarkTrimNewLine 性能测试 TrimNewLine
func BenchmarkTrimNewLine(b *testing.B) {
	str := "a\r\nb\nc\r\nd\r\ne\nf\r"
	// 预热
	_ = TrimNewLine(str)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TrimNewLine(str)
	}
}

// BenchmarkTrimSuffixAll 性能测试 TrimSuffixAll
func BenchmarkTrimSuffixAll(b *testing.B) {
	str := "test////////////////////"
	trim := "/"
	// 预热
	_ = TrimSuffixAll(str, trim)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TrimSuffixAll(str, trim)
	}
}

// BenchmarkGetValueFromJointStr 性能测试 GetValueFromJointStr
func BenchmarkGetValueFromJointStr(b *testing.B) {
	line := "a=1&b=2&c=3&d=4&e=5&f=6&g=7&h=8&i=9&j=10"
	key := "j"
	separator := "&"
	// 预热
	_ = GetValueFromJointStr(line, key, separator)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetValueFromJointStr(line, key, separator)
	}
}

// BenchmarkOmitString 性能测试 OmitString
func BenchmarkOmitString(b *testing.B) {
	str := "1234567890123456789012345678901234567890"
	end := uint64(20)
	// 预热
	_ = OmitString(str, end)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OmitString(str, end)
	}
}

// BenchmarkStringsToInSqlString 性能测试 StringsToInSqlString
func BenchmarkStringsToInSqlString(b *testing.B) {
	s := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o"}
	// 预热
	_ = StringsToInSqlString(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StringsToInSqlString(s)
	}
}

// BenchmarkInt64sToInString 性能测试 Int64sToInString
func BenchmarkInt64sToInString(b *testing.B) {
	s := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	// 预热
	_ = Int64sToInString(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Int64sToInString(s)
	}
}
