package funcutil

import (
	"strconv"
	"testing"
)

// -------------------------- 单元测试 - Map --------------------------
func TestMap(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		f     func(int) string
		want  []string
	}{
		{
			name:  "nil输入切片",
			input: nil,
			f:     func(i int) string { return strconv.Itoa(i) },
			want:  []string{}, // 非nil空切片
		},
		{
			name:  "空切片",
			input: []int{},
			f:     func(i int) string { return strconv.Itoa(i) },
			want:  []string{},
		},
		{
			name:  "单元素切片",
			input: []int{10},
			f:     func(i int) string { return strconv.Itoa(i) },
			want:  []string{"10"},
		},
		{
			name:  "多元素切片（int转string）",
			input: []int{1, 2, 3, 4},
			f:     func(i int) string { return strconv.Itoa(i * 2) },
			want:  []string{"2", "4", "6", "8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Map(tt.input, tt.f)
			// 验证返回值非nil（即使输入nil）
			if got == nil {
				t.Errorf("Map() 返回了nil切片，预期非nil空切片")
			}
			// 验证长度和内容
			if len(got) != len(tt.want) {
				t.Errorf("Map() 长度不符: 预期 %d, 实际 %d", len(tt.want), len(got))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Map() 索引%d值不符: 预期 %s, 实际 %s", i, tt.want[i], got[i])
				}
			}
		})
	}
}

// -------------------------- 单元测试 - Reduce --------------------------
func TestReduce(t *testing.T) {
	tests := []struct {
		name        string
		input       []int
		initializer int
		f           func(int, int) int
		want        int
	}{
		{
			name:        "nil输入切片（返回初始值）",
			input:       nil,
			initializer: 0,
			f:           func(acc, v int) int { return acc + v },
			want:        0,
		},
		{
			name:        "空切片（返回初始值）",
			input:       []int{},
			initializer: 100,
			f:           func(acc, v int) int { return acc + v },
			want:        100,
		},
		{
			name:        "单元素切片（求和）",
			input:       []int{5},
			initializer: 0,
			f:           func(acc, v int) int { return acc + v },
			want:        5,
		},
		{
			name:        "多元素切片（累乘）",
			input:       []int{2, 3, 4},
			initializer: 1,
			f:           func(acc, v int) int { return acc * v },
			want:        24,
		},
	}

	// 额外测试字符串拼接场景（验证泛型兼容性）
	t.Run("字符串拼接（nil输入）", func(t *testing.T) {
		got := Reduce(nil, "", func(acc, v string) string { return acc + v })
		if got != "" {
			t.Errorf("Reduce() 预期 '', 实际 %s", got)
		}
	})
	t.Run("字符串拼接（多元素）", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		got := Reduce(input, "", func(acc, v string) string { return acc + v })
		if got != "abc" {
			t.Errorf("Reduce() 预期 'abc', 实际 %s", got)
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reduce(tt.input, tt.initializer, tt.f)
			if got != tt.want {
				t.Errorf("Reduce() 预期 %d, 实际 %d", tt.want, got)
			}
		})
	}
}

// -------------------------- 单元测试 - Filter --------------------------
func TestFilter(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		f     func(int) bool
		want  []int
	}{
		{
			name:  "nil输入切片（返回非nil空切片）",
			input: nil,
			f:     func(i int) bool { return i%2 == 0 },
			want:  []int{},
		},
		{
			name:  "空切片",
			input: []int{},
			f:     func(i int) bool { return i%2 == 0 },
			want:  []int{},
		},
		{
			name:  "单元素（符合条件）",
			input: []int{4},
			f:     func(i int) bool { return i%2 == 0 },
			want:  []int{4},
		},
		{
			name:  "单元素（不符合条件）",
			input: []int{5},
			f:     func(i int) bool { return i%2 == 0 },
			want:  []int{},
		},
		{
			name:  "多元素（部分符合）",
			input: []int{1, 2, 3, 4, 5},
			f:     func(i int) bool { return i%2 == 0 },
			want:  []int{2, 4},
		},
		{
			name:  "多元素（全符合）",
			input: []int{2, 4, 6},
			f:     func(i int) bool { return i%2 == 0 },
			want:  []int{2, 4, 6},
		},
		{
			name:  "多元素（全不符合）",
			input: []int{1, 3, 5},
			f:     func(i int) bool { return i%2 == 0 },
			want:  []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filter(tt.input, tt.f)
			// 核心验证：返回值永远非nil（修复后的关键行为）
			if got == nil {
				t.Errorf("Filter() 返回了nil切片，预期非nil空切片")
			}
			// 验证长度和内容
			if len(got) != len(tt.want) {
				t.Errorf("Filter() 长度不符: 预期 %d, 实际 %d", len(tt.want), len(got))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Filter() 索引%d值不符: 预期 %d, 实际 %d", i, tt.want[i], got[i])
				}
			}
		})
	}
}

// -------------------------- 性能测试 - Benchmark --------------------------
// 生成测试用切片
func generateTestSlice(size int) []int {
	s := make([]int, size)
	for i := 0; i < size; i++ {
		s[i] = i
	}
	return s
}

// Map性能测试：不同切片规模
func BenchmarkMap_Small(b *testing.B)  { benchmarkMap(b, 100) }    // 小切片
func BenchmarkMap_Medium(b *testing.B) { benchmarkMap(b, 10000) }  // 中切片
func BenchmarkMap_Large(b *testing.B)  { benchmarkMap(b, 100000) } // 大切片
func benchmarkMap(b *testing.B, size int) {
	s := generateTestSlice(size)
	f := func(i int) string { return strconv.Itoa(i) }
	b.ResetTimer() // 重置计时器，排除切片生成耗时
	for i := 0; i < b.N; i++ {
		Map(s, f)
	}
}

// Reduce性能测试：不同切片规模
func BenchmarkReduce_Small(b *testing.B)  { benchmarkReduce(b, 100) }    // 小切片
func BenchmarkReduce_Medium(b *testing.B) { benchmarkReduce(b, 10000) }  // 中切片
func BenchmarkReduce_Large(b *testing.B)  { benchmarkReduce(b, 100000) } // 大切片
func benchmarkReduce(b *testing.B, size int) {
	s := generateTestSlice(size)
	f := func(acc, v int) int { return acc + v }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reduce(s, 0, f)
	}
}

// Filter性能测试：不同切片规模（测试"部分符合"场景，更贴近真实业务）
func BenchmarkFilter_Small(b *testing.B)  { benchmarkFilter(b, 100) }    // 小切片
func BenchmarkFilter_Medium(b *testing.B) { benchmarkFilter(b, 10000) }  // 中切片
func BenchmarkFilter_Large(b *testing.B)  { benchmarkFilter(b, 100000) } // 大切片
func benchmarkFilter(b *testing.B, size int) {
	s := generateTestSlice(size)
	f := func(i int) bool { return i%2 == 0 } // 约50%元素符合条件
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Filter(s, f)
	}
}
