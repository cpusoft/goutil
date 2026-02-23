package asn1util

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ======================== 基础功能测试 ========================

// TestGetDNFromName 测试从 pkix.Name 解析 DN
func TestGetDNFromName(t *testing.T) {
	tests := []struct {
		name    string
		nameObj pkix.Name
		sep     string
		want    string
		wantErr bool
	}{
		{
			name: "标准DN字段解析",
			nameObj: pkix.Name{
				CommonName:         "test.cn",
				Country:            []string{"CN"},
				Organization:       []string{"test-org"},
				OrganizationalUnit: []string{"test-ou"},
			},
			sep:     ",",
			want:    "C=CN,O=test-org,OU=test-ou,CN=test.cn", // 修正字段顺序
			wantErr: false,
		},
		{
			name: "空Name对象",
			nameObj: pkix.Name{
				ExtraNames: []pkix.AttributeTypeAndValue{},
			},
			sep:     ",",
			want:    "",
			wantErr: false,
		},
		{
			name: "自定义分隔符",
			nameObj: pkix.Name{
				CommonName: "test",
			},
			sep:     "|",
			want:    "CN=test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDNFromName(tt.nameObj, tt.sep)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGetDNFromRDNSeq 测试从 RDNSequence 解析 DN（核心修复验证）
func TestGetDNFromRDNSeq(t *testing.T) {
	tests := []struct {
		name    string
		rdns    pkix.RDNSequence
		sep     string
		want    string
		wantErr bool
	}{
		{
			name: "标准字符串值OID映射",
			rdns: pkix.RDNSequence{
				{
					{Type: asn1.ObjectIdentifier{2, 5, 4, 3}, Value: "test.cn"}, // CN
					{Type: asn1.ObjectIdentifier{2, 5, 4, 6}, Value: "CN"},      // C
				},
			},
			sep:     ",",
			want:    "CN=test.cn,C=CN",
			wantErr: false,
		},
		{
			name: "非字符串值OID映射（serialNumber）",
			rdns: pkix.RDNSequence{
				{
					{Type: asn1.ObjectIdentifier{2, 5, 4, 5}, Value: 12345}, // serialNumber
				},
			},
			sep:     ",",
			want:    "serialNumber=12345",
			wantErr: false,
		},
		{
			name: "未知OID解析",
			rdns: pkix.RDNSequence{
				{
					{Type: asn1.ObjectIdentifier{1, 2, 3, 4}, Value: "unknown"},
				},
			},
			sep:     ",",
			want:    "1.2.3.4=unknown",
			wantErr: false,
		},
		{
			name: "混合类型值",
			rdns: pkix.RDNSequence{
				{
					{Type: asn1.ObjectIdentifier{2, 5, 4, 3}, Value: "test"},       // string
					{Type: asn1.ObjectIdentifier{2, 5, 4, 5}, Value: uint64(6789)}, // uint64
				},
			},
			sep:     "|",
			want:    "CN=test|serialNumber=6789",
			wantErr: false,
		},
		{
			name:    "空RDN序列",
			rdns:    pkix.RDNSequence{},
			sep:     ",",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDNFromRDNSeq(tt.rdns, tt.sep)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestDecodeString 测试字符串解码（UTF8/IA5通用）
func TestDecodeString(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "正常UTF8字符串",
			data: []byte("hello 测试"),
			want: "hello 测试",
		},
		{
			name: "空字节切片",
			data: []byte{},
			want: "",
		},
		{
			name: "特殊字符",
			data: []byte("!@#$%^&*()"),
			want: "!@#$%^&*()",
		},
		{
			name: "单字节",
			data: []byte("a"),
			want: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证三个字符串解码函数行为一致
			assert.Equal(t, tt.want, DecodeString(tt.data))
			assert.Equal(t, tt.want, DecodeUTF8String(tt.data))
			assert.Equal(t, tt.want, DecodeIA5String(tt.data))
		})
	}
}

// TestDecodeBool 测试布尔值解码（边界值）
func TestDecodeBool(t *testing.T) {
	tests := []struct {
		name string
		data byte
		want bool
	}{
		{
			name: "布尔值true（0x01）",
			data: 0x01,
			want: true,
		},
		{
			name: "布尔值false（0x00）",
			data: 0x00,
			want: false,
		},
		{
			name: "非零值均为true（0xFF）",
			data: 0xFF,
			want: true,
		},
		{
			name: "边界值（0x02）",
			data: 0x02,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, DecodeBool(tt.data))
		})
	}
}

// TestDecodeUTCTime 测试UTC时间解码（临界值+异常场景）
func TestDecodeUTCTime(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{
			name:    "标准13字节UTC时间",
			data:    []byte("240223123456Z"), // 2024-02-23 12:34:56Z
			want:    "2024-02-23 12:34:56Z",
			wantErr: false,
		},
		{
			name:    "超长字节（截断）",
			data:    []byte("240223123456Zextra"),
			want:    "2024-02-23 12:34:56Z",
			wantErr: false,
		},
		{
			name:    "临界值：12字节（不足13）",
			data:    []byte("24022312345"),
			want:    "",
			wantErr: true,
		},
		{
			name:    "空字节",
			data:    []byte{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "边界值：刚好13字节",
			data:    []byte("991231235959Z"), // 2099-12-31 23:59:59Z
			want:    "2099-12-31 23:59:59Z",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeUTCTime(tt.data)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestDecodeGeneralizedTime 测试通用时间解码（临界值+异常场景）
func TestDecodeGeneralizedTime(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{
			name:    "标准15字节通用时间",
			data:    []byte("20240223123456Z"),
			want:    "2024-02-23 12:34:56Z",
			wantErr: false,
		},
		{
			name:    "超长字节（截断）",
			data:    []byte("20240223123456Zextra"),
			want:    "2024-02-23 12:34:56Z",
			wantErr: false,
		},
		{
			name:    "临界值：14字节（不足15）",
			data:    []byte("2024022312345"),
			want:    "",
			wantErr: true,
		},
		{
			name:    "空字节",
			data:    []byte{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "边界值：9999年（最大4位年份）",
			data:    []byte("99991231235959Z"),
			want:    "9999-12-31 23:59:59Z",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeGeneralizedTime(tt.data)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestDecodeOid 测试OID解码（正常+溢出+边界场景）
func TestDecodeOid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "标准OID（2.5.4.3）",
			data: []byte{0x55, 0x04, 0x03}, // 2*40+5=85=0x55, 4=0x04, 3=0x03
			want: "2.5.4.3",
		},
		{
			name: "长OID（1.2.840.113549.1.9.1）",
			data: []byte{0x2A, 0x86, 0x48, 0x86, 0xF7, 0x0D, 0x01, 0x09, 0x01},
			want: "1.2.840.113549.1.9.1",
		},
		{
			name: "空字节",
			data: []byte{},
			want: "",
		},
		{
			name: "溢出风险OID（全0xFF）",
			data: []byte{0xFF, 0xFF, 0xFF, 0xFF},
			want: "2.175", // 0xFF=255 → 255-80=175
		},
		{
			name: "边界值：最小OID（0.0）",
			data: []byte{0x00},
			want: "0.0",
		},
		{
			name: "溢出保护场景",
			data: []byte{0xFF, 0xFF, 0x7F}, // 0xFF(255) → 2.175, 0xFF<<7+0x7F=16383
			want: "2.175.16383",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, DecodeOid(tt.data))
		})
	}
}

// TestDecodeBytes 测试字节解码为十六进制字符串（无多余空格）
func TestDecodeBytes(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "正常多字节",
			data: []byte{0x01, 0x02, 0xFF},
			want: "01 02 ff",
		},
		{
			name: "空字节",
			data: []byte{},
			want: "",
		},
		{
			name: "单字节",
			data: []byte{0x00},
			want: "00",
		},
		{
			name: "边界值：0x00-0xFF",
			data: []byte{0x00, 0xFF},
			want: "00 ff",
		},
		{
			name: "大字节切片（1024字节）",
			data: func() []byte {
				b := make([]byte, 1024)
				for i := 0; i < 1024; i++ {
					b[i] = byte(i % 256)
				}
				return b
			}(),
			want: func() string {
				var s string
				for i := 0; i < 1024; i++ {
					if i > 0 {
						s += " "
					}
					s += fmt.Sprintf("%02x", byte(i%256))
				}
				return s
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, DecodeBytes(tt.data))
		})
	}
}

// TestEncodeDecodeInteger 测试整数编解码双向验证（边界值+溢出）
func TestEncodeDecodeInteger(t *testing.T) {
	tests := []struct {
		name       string
		val        uint64
		wantEncode []byte
		wantDecode uint64
		wantWarn   bool // 是否预期溢出警告
	}{
		{
			name:       "零值",
			val:        0,
			wantEncode: []byte{0x00},
			wantDecode: 0,
			wantWarn:   false,
		},
		{
			name:       "正常整数（123456）",
			val:        123456,
			wantEncode: []byte{0x01, 0xe2, 0x40}, // 123456 = 0x1E240
			wantDecode: 123456,
			wantWarn:   false,
		},
		{
			name:       "最大值（math.MaxUint64）",
			val:        math.MaxUint64,
			wantEncode: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			wantDecode: math.MaxUint64,
			wantWarn:   false,
		},
		{
			name:       "溢出场景（9字节输入）",
			val:        0,
			wantEncode: []byte{0x00},
			wantDecode: 0,
			wantWarn:   true,
		},
		{
			name:       "边界值（1字节）",
			val:        255,
			wantEncode: []byte{0xff},
			wantDecode: 255,
			wantWarn:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试编码
			encoded := EncodeInteger(tt.val)
			assert.Equal(t, tt.wantEncode, encoded)

			// 测试解码
			if tt.wantWarn {
				// 构造9字节溢出输入
				longData := make([]byte, 9)
				decoded := DecodeInteger(longData)
				assert.Equal(t, tt.wantDecode, decoded)
			} else {
				decoded := DecodeInteger(encoded)
				assert.Equal(t, tt.wantDecode, decoded)
			}
		})
	}
}

// TestDecodeFiniteLen 测试有限长度解码（正常+数据不足+边界场景）
func TestDecodeFiniteLen(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantLen uint64
		wantPos uint64
		wantErr bool
	}{
		{
			name:    "短长度（0x03）",
			data:    []byte{0x30, 0x03, 0x01, 0x02, 0x03}, // 类型0x30，长度0x03
			wantLen: 3,
			wantPos: 2,
			wantErr: false,
		},
		{
			name:    "长长度（0x82 0x00 0x05）",
			data:    []byte{0x30, 0x82, 0x00, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05}, // 长度5
			wantLen: 5,
			wantPos: 4,
			wantErr: false,
		},
		{
			name:    "数据不足-长度字段不完整",
			data:    []byte{0x30, 0x82, 0x00}, // 0x82需要2字节长度字段，仅1字节
			wantLen: 0,
			wantPos: 0,
			wantErr: true,
		},
		{
			name:    "数据不足-内容不足",
			data:    []byte{0x30, 0x82, 0x00, 0x05, 0x01}, // 长度5，仅1字节内容
			wantLen: 0,
			wantPos: 0,
			wantErr: true,
		},
		{
			name:    "空数据",
			data:    []byte{},
			wantLen: 0,
			wantPos: 0,
			wantErr: true,
		},
		{
			name: "边界值：长度字段8字节（最大）",
			data: func() []byte {
				// 基础部分：类型(0x30) + 长度标识(0x88=136 → 136-128=8字节长度字段) + 8字节长度值(0x00000000000000FF=255)
				base := []byte{0x30, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF}
				// 补充255字节内容（满足datapos+datalen ≤ len(data)）
				content := make([]byte, 255)
				for i := range content {
					content[i] = 0x01
				}
				return append(base, content...)
			}(),
			wantLen: 255, // 8字节长度值解析结果：0x00000000000000FF=255
			wantPos: 10,  // 2（基础偏移） + 8（长度字段）=10
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLen, gotPos, err := DecodeFiniteLen(tt.data)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantLen, gotLen)
			assert.Equal(t, tt.wantPos, gotPos)
		})
	}
}

// TestDecodeInfiniteLen 测试无限长度解码（0x0000结束符）
func TestDecodeInfiniteLen(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantLen uint64
		wantPos uint64
		wantErr bool
	}{
		{
			name:    "正常结束符",
			data:    []byte{0x30, 0x80, 0x01, 0x02, 0x00, 0x00},
			wantLen: 4,
			wantPos: 2,
			wantErr: false,
		},
		{
			name:    "无结束符",
			data:    []byte{0x30, 0x80, 0x01, 0x02},
			wantLen: 4,
			wantPos: 2,
			wantErr: true,
		},
		{
			name:    "空数据",
			data:    []byte{},
			wantLen: 0,
			wantPos: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLen, gotPos, err := DecodeInfiniteLen(tt.data)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantLen, gotLen)
			assert.Equal(t, tt.wantPos, gotPos)
		})
	}
}

// TestDecodeFiniteAndInfiniteLen 测试混合长度解码
func TestDecodeFiniteAndInfiniteLen(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantLen uint64
		wantPos uint64
		wantErr bool
	}{
		{
			name:    "有限长度",
			data:    []byte{0x30, 0x03, 0x01, 0x02, 0x03},
			wantLen: 3,
			wantPos: 2,
			wantErr: false,
		},
		{
			name:    "无限长度",
			data:    []byte{0x30, 0x80, 0x01, 0x02, 0x00, 0x00},
			wantLen: 4,
			wantPos: 2,
			wantErr: false,
		},
		{
			name:    "空数据",
			data:    []byte{},
			wantLen: 0,
			wantPos: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLen, gotPos, err := DecodeFiniteAndInfiniteLen(tt.data)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantLen, gotLen)
			assert.Equal(t, tt.wantPos, gotPos)
		})
	}
}

// TestIndexEndOfBytes 测试0x0000结束符查找（正常+边界+异常）
func TestIndexEndOfBytes(t *testing.T) {
	tests := []struct {
		name              string
		oldb              []byte
		tagType           uint8
		hierarchyFor00    int
		topHierarchyFor00 int
		want              int
		wantErr           bool
	}{
		{
			name:              "正常结束符",
			oldb:              []byte{0x30, 0x80, 0x01, 0x02, 0x00, 0x00},
			tagType:           32,
			hierarchyFor00:    0,
			topHierarchyFor00: 1,
			want:              4,
			wantErr:           false,
		},
		{
			name:              "多个连续结束符",
			oldb:              []byte{0x30, 0x80, 0x00, 0x00, 0x00, 0x00},
			tagType:           32,
			hierarchyFor00:    0,
			topHierarchyFor00: 1,
			want:              4,
			wantErr:           false,
		},
		{
			name:              "无结束符",
			oldb:              []byte{0x30, 0x80, 0x01, 0x02},
			tagType:           32,
			hierarchyFor00:    0,
			topHierarchyFor00: 1,
			want:              4,
			wantErr:           true,
		},
		{
			name:              "短字节（长度≤2）",
			oldb:              []byte{0x30},
			tagType:           32,
			hierarchyFor00:    0,
			topHierarchyFor00: 1,
			want:              -1,
			wantErr:           true,
		},
		{
			name:              "TypeLastIndex场景（0xA0）",
			oldb:              []byte{0xA0, 0x80, 0x01, 0x02, 0x00, 0x00},
			tagType:           32,
			hierarchyFor00:    0,
			topHierarchyFor00: 1,
			want:              4,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IndexEndOfBytes(tt.oldb, tt.tagType, tt.hierarchyFor00, tt.topHierarchyFor00)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGetTopHierarchyFor00 测试层级计算（正常+边界）
func TestGetTopHierarchyFor00(t *testing.T) {
	tests := []struct {
		name string
		oldb []byte
		want int
	}{
		{
			name: "单层结束符",
			oldb: []byte{0x30, 0x80, 0x01, 0x02, 0x00, 0x00},
			want: 1,
		},
		{
			name: "多层结束符",
			oldb: []byte{0x30, 0x80, 0x01, 0x00, 0x00, 0x02, 0x00, 0x00},
			want: 2,
		},
		{
			name: "空字节",
			oldb: []byte{},
			want: 0,
		},
		{
			name: "无结束符",
			oldb: []byte{0x30, 0x80, 0x01, 0x02},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetTopHierarchyFor00(tt.oldb))
		})
	}
}

// TestTrimPrefixSuffix00 测试前后缀0x00修剪（单个/连续/混合）
func TestTrimPrefixSuffix00(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantTrim []byte
	}{
		{
			name:     "前缀连续0x0000",
			input:    []byte{0x00, 0x00, 0x01, 0x02},
			wantTrim: []byte{0x01, 0x02},
		},
		{
			name:     "后缀连续0x0000",
			input:    []byte{0x01, 0x02, 0x00, 0x00},
			wantTrim: []byte{0x01, 0x02},
		},
		{
			name:     "前后缀单个0x00",
			input:    []byte{0x00, 0x01, 0x02, 0x00},
			wantTrim: []byte{0x01, 0x02},
		},
		{
			name:     "全0x00",
			input:    []byte{0x00, 0x00, 0x00},
			wantTrim: []byte{},
		},
		{
			name:     "空字节",
			input:    []byte{},
			wantTrim: []byte{},
		},
		{
			name:     "混合连续+单个0x00",
			input:    []byte{0x00, 0x00, 0x01, 0x02, 0x00},
			wantTrim: []byte{0x01, 0x02},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 组合修剪：先前缀后后缀
			trimmed := TrimPrefix00(tt.input)
			trimmed, _ = TrimSuffix00(trimmed, len(trimmed))
			assert.Equal(t, tt.wantTrim, trimmed)
		})
	}
}

// TestExtKeyUsagesToInts 测试扩展密钥用法转换（标准枚举值）
func TestExtKeyUsagesToInts(t *testing.T) {
	tests := []struct {
		name string
		exts []x509.ExtKeyUsage
		want []int
	}{
		{
			name: "正常转换（ServerAuth+ClientAuth）",
			exts: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			want: []int{1, 2}, // 标准枚举值
		},
		{
			name: "空切片",
			exts: []x509.ExtKeyUsage{},
			want: []int{},
		},
		{
			name: "ExtKeyUsageAny",
			exts: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
			want: []int{0},
		},
		{
			name: "多值转换",
			exts: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning, x509.ExtKeyUsageEmailProtection},
			want: []int{3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExtKeyUsagesToInts(tt.exts))
		})
	}
}

// ======================== 性能测试 ========================

// BenchmarkGetDNFromRDNSeq DN解析性能测试
func BenchmarkGetDNFromRDNSeq(b *testing.B) {
	// 构造复杂测试数据
	rdns := pkix.RDNSequence{
		{
			{Type: asn1.ObjectIdentifier{2, 5, 4, 3}, Value: "test.cn"},
			{Type: asn1.ObjectIdentifier{2, 5, 4, 6}, Value: "CN"},
			{Type: asn1.ObjectIdentifier{2, 5, 4, 10}, Value: "test-org"},
			{Type: asn1.ObjectIdentifier{2, 5, 4, 11}, Value: "test-ou"},
			{Type: asn1.ObjectIdentifier{2, 5, 4, 5}, Value: 123456},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetDNFromRDNSeq(rdns, ",")
	}
}

// BenchmarkDecodeOid OID解码性能测试
func BenchmarkDecodeOid(b *testing.B) {
	// 长OID：1.2.840.113549.1.9.1
	data := []byte{0x2A, 0x86, 0x48, 0x86, 0xF7, 0x0D, 0x01, 0x09, 0x01}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DecodeOid(data)
	}
}

// BenchmarkDecodeBytes 字节解码性能测试（1KB数据）
func BenchmarkDecodeBytes(b *testing.B) {
	data := make([]byte, 1024)
	for i := 0; i < 1024; i++ {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DecodeBytes(data)
	}
}

// BenchmarkEncodeDecodeInteger 整数编解码性能测试
func BenchmarkEncodeDecodeInteger(b *testing.B) {
	val := uint64(1234567890)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoded := EncodeInteger(val)
		_ = DecodeInteger(encoded)
	}
}

// BenchmarkDecodeFiniteLen 长度解析性能测试
func BenchmarkDecodeFiniteLen(b *testing.B) {
	data := []byte{0x30, 0x82, 0x00, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = DecodeFiniteLen(data)
	}
}

// BenchmarkIndexEndOfBytes 结束符查找性能测试
func BenchmarkIndexEndOfBytes(b *testing.B) {
	data := []byte{0x30, 0x80}
	data = append(data, make([]byte, 1024)...)
	data = append(data, 0x00, 0x00)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = IndexEndOfBytes(data, 32, 0, 1)
	}
}

// BenchmarkTrimPrefixSuffix00 字节修剪性能测试
func BenchmarkTrimPrefixSuffix00(b *testing.B) {
	data := []byte{0x00, 0x00, 0x01, 0x02, 0x03, 0x00, 0x00}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trimmed := TrimPrefix00(data)
		_, _ = TrimSuffix00(trimmed, len(trimmed))
	}
}

// ======================== 集成测试（模拟真实场景） ========================

// TestIntegration_ParseCertDN 集成测试：模拟解析证书DN
func TestIntegration_ParseCertDN(t *testing.T) {
	// 模拟证书的RDNSequence
	rdns := pkix.RDNSequence{
		{
			{Type: asn1.ObjectIdentifier{2, 5, 4, 3}, Value: "example.com"},
			{Type: asn1.ObjectIdentifier{2, 5, 4, 6}, Value: "US"},
			{Type: asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}, Value: "admin@example.com"},
			{Type: asn1.ObjectIdentifier{2, 5, 4, 5}, Value: uint64(987654321)},
		},
	}

	// 步骤1：解析DN
	dn, err := GetDNFromRDNSeq(rdns, ",")
	assert.NoError(t, err)
	assert.Equal(t, "CN=example.com,C=US,emailAddress=admin@example.com,serialNumber=987654321", dn)

	// 步骤2：验证OID映射
	oidStr := DecodeOid([]byte{0x2A, 0x86, 0x48, 0x86, 0xF7, 0x0D, 0x01, 0x09, 0x01})
	assert.Equal(t, "1.2.840.113549.1.9.1", oidStr)

	// 步骤3：验证字节修剪
	rawBytes := []byte{0x00, 0x00, 0x01, 0x02, 0x00}
	trimmed := TrimPrefix00(rawBytes)
	trimmed, _ = TrimSuffix00(trimmed, len(trimmed))
	assert.Equal(t, []byte{0x01, 0x02}, trimmed)
}
