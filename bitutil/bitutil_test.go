package bitutil

import (
	"fmt"
	"testing"
)

/*
# 运行所有测试
go test -v

# 运行指定函数的测试
go test -v -run TestShift0x00LeftFillOne

# 运行基准测试（查看性能）
go test -bench=.
*/
// TestShift0x00LeftFillOne 测试 Shift0x00LeftFillOne 函数的正确性
func TestShift0x00LeftFillOne(t *testing.T) {
	// 定义测试用例：输入bits -> 预期返回值
	testCases := []struct {
		name     string
		input    uint8
		expected byte
	}{
		{
			name:     "bits=0（边界值）",
			input:    0,
			expected: 0,
		},
		{
			name:     "bits=1（最小有效位）",
			input:    1,
			expected: 1, // 0000 0001
		},
		{
			name:     "bits=2（示例场景）",
			input:    2,
			expected: 3, // 0000 0011
		},
		{
			name:     "bits=4（中间值）",
			input:    4,
			expected: 15, // 0000 1111
		},
		{
			name:     "bits=8（最大有效位）",
			input:    8,
			expected: 255, // 1111 1111
		},
		{
			name:     "bits=9（超过8位，按8位处理）",
			input:    9,
			expected: 255,
		},
		{
			name:     "bits=255（极端异常值，按8位处理）",
			input:    255,
			expected: 255,
		},
	}

	// 遍历测试用例执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Shift0x00LeftFillOne(tc.input)
			if result != tc.expected {
				t.Errorf("输入 %d 时，预期返回 0x%02x，实际返回 0x%02x", tc.input, tc.expected, result)
			}
		})
	}
}

// TestShift0xffLeftFillZero 测试 Shift0xffLeftFillZero 函数的正确性
func TestShift0xffLeftFillZero(t *testing.T) {
	// 定义测试用例：输入bits -> 预期返回值
	testCases := []struct {
		name     string
		input    uint8
		expected byte
	}{
		{
			name:     "bits=0（边界值）",
			input:    0,
			expected: 255, // 1111 1111
		},
		{
			name:     "bits=1（最小有效位）",
			input:    1,
			expected: 254, // 1111 1110
		},
		{
			name:     "bits=6（示例场景）",
			input:    6,
			expected: 192, // 1100 0000
		},
		{
			name:     "bits=7（中间值）",
			input:    7,
			expected: 128, // 1000 0000
		},
		{
			name:     "bits=8（最大有效位）",
			input:    8,
			expected: 0, // 0000 0000
		},
		{
			name:     "bits=9（超过8位，返回0）",
			input:    9,
			expected: 0,
		},
		{
			name:     "bits=255（极端异常值，返回0）",
			input:    255,
			expected: 0,
		},
	}

	// 遍历测试用例执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Shift0xffLeftFillZero(tc.input)
			if result != tc.expected {
				t.Errorf("输入 %d 时，预期返回 0x%02x，实际返回 0x%02x", tc.input, tc.expected, result)
			}
		})
	}
}

// BenchmarkShift0x00LeftFillOne 基准测试（可选，用于性能分析）
func BenchmarkShift0x00LeftFillOne(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Shift0x00LeftFillOne(4)
	}
}

// BenchmarkShift0xffLeftFillZero 基准测试（可选，用于性能分析）
func BenchmarkShift0xffLeftFillZero(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Shift0xffLeftFillZero(6)
	}
}

func TestXYZ(t *testing.T) {
	var z uint8 = 0x80
	a := Shift0x00LeftFillOne(3)
	fmt.Printf("%02x,%d,%b\n", a, a, a)
	z = z | a
	fmt.Printf("%02x,%d,%b\n", z, z, z)

	z = 0xff
	a = Shift0xffLeftFillZero(5)
	fmt.Printf("%02x,%d,%b\n", a, a, a)
	z = z & a
	fmt.Printf("%02x,%d,%b\n", z, z, z)

	newB := Shift0x00LeftFillOne(2)
	fmt.Println(newB)

	newB = Shift0xffLeftFillZero(2)
	fmt.Println(newB)
}
