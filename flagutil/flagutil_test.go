package flagutil

import (
	"flag"
	"fmt"
	"reflect" // 新增：用于深度比较切片
	"testing"
)

// ---------------- 单元测试（全覆盖 + 临界值） ----------------

// TestArrayFlags_String 测试 String() 方法的所有场景（临界值：nil指针、空切片、有元素）
func TestArrayFlags_String(t *testing.T) {
	// 定义测试用例
	tests := []struct {
		name     string
		input    *ArrayFlags // 传入的ArrayFlags指针（模拟不同状态）
		expected string      // 期望的输出结果
	}{
		{
			name:     "nil指针场景",
			input:    nil,
			expected: "",
		},
		{
			name:     "空切片场景",
			input:    &ArrayFlags{}, // 非nil但空切片
			expected: "[]",          // fmt.Sprint([]string{}) 的默认输出
		},
		{
			name:     "单个元素场景",
			input:    &ArrayFlags{"192.168.0.55"},
			expected: "[192.168.0.55]",
		},
		{
			name:     "多个元素场景（临界值：2个元素，模拟命令行传参）",
			input:    &ArrayFlags{"192.168.0.55", "192.168.0.56"},
			expected: "[192.168.0.55 192.168.0.56]",
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("String() 输出错误：期望 %q，实际 %q", tt.expected, result)
			}
		})
	}
}

// TestArrayFlags_Set 测试 Set() 方法的所有场景（临界值：nil指针、正常添加1个/多个元素）
func TestArrayFlags_Set(t *testing.T) {
	// 定义测试用例
	tests := []struct {
		name      string
		inputPtr  *ArrayFlags // 传入的ArrayFlags指针
		value     string      // Set方法的入参
		wantErr   bool        // 是否期望返回错误
		wantSlice ArrayFlags  // 期望的切片结果
	}{
		{
			name:      "nil指针场景（临界值：必返回错误）",
			inputPtr:  nil,
			value:     "192.168.0.55",
			wantErr:   true,
			wantSlice: nil,
		},
		{
			name:      "正常添加单个元素",
			inputPtr:  &ArrayFlags{},
			value:     "192.168.0.55",
			wantErr:   false,
			wantSlice: ArrayFlags{"192.168.0.55"},
		},
		{
			name:      "正常添加多个元素（模拟命令行多次传参）",
			inputPtr:  &ArrayFlags{},
			value:     "192.168.0.56", // 先加56
			wantErr:   false,
			wantSlice: ArrayFlags{"192.168.0.56"}, // 第一步结果
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inputPtr.Set(tt.value)
			// 校验错误是否符合预期
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() 错误状态错误：期望错误=%v，实际错误=%v（错误信息：%v）", tt.wantErr, err != nil, err)
				return
			}
			// 无错误时，校验切片内容（修复：改用reflect.DeepEqual比较切片）
			if !tt.wantErr {
				if len(*tt.inputPtr) != len(tt.wantSlice) {
					t.Errorf("Set() 切片长度错误：期望 %d，实际 %d", len(tt.wantSlice), len(*tt.inputPtr))
					return
				}
				// 修复点1：用reflect.DeepEqual比较切片内容（替代直接!=）
				if !reflect.DeepEqual(*tt.inputPtr, tt.wantSlice) {
					t.Errorf("Set() 切片元素错误：期望 %v，实际 %v", tt.wantSlice, *tt.inputPtr)
				}
			}
		})
	}

	// 额外测试：多次调用Set（模拟命令行传多个--addr）
	t.Run("多次调用Set添加多个元素", func(t *testing.T) {
		af := &ArrayFlags{}
		// 模拟命令行传2个addr
		err1 := af.Set("192.168.0.55")
		err2 := af.Set("192.168.0.56")
		if err1 != nil || err2 != nil {
			t.Fatalf("多次Set失败：err1=%v, err2=%v", err1, err2)
		}
		// 校验最终切片（修复点2：替换直接!=）
		expected := ArrayFlags{"192.168.0.55", "192.168.0.56"}
		if !reflect.DeepEqual(*af, expected) {
			t.Errorf("多次Set结果错误：期望 %v，实际 %v", expected, *af)
		}
	})
}

// TestArrayFlags_Integration 集成测试：模拟真实命令行传参场景（最接近实际使用的临界场景）
func TestArrayFlags_Integration(t *testing.T) {
	// 重置flag包，避免其他测试干扰
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// 模拟命令行传参：--addr 192.168.0.55 --addr 192.168.0.56
	var addrs ArrayFlags
	flag.Var(&addrs, "addr", "Addresses")
	// 模拟解析命令行参数
	err := flag.CommandLine.Parse([]string{"--addr", "192.168.0.55", "--addr", "192.168.0.56"})
	if err != nil {
		t.Fatalf("解析命令行参数失败：%v", err)
	}

	// 校验结果（修复点3：替换直接!=）
	expected := ArrayFlags{"192.168.0.55", "192.168.0.56"}
	if !reflect.DeepEqual(addrs, expected) {
		t.Errorf("集成测试结果错误：期望 %v，实际 %v", expected, addrs)
	}
}

// ---------------- 性能测试（基准测试） ----------------

// BenchmarkArrayFlags_Set 基准测试：正常场景下Set()方法的性能（高频调用场景）
func BenchmarkArrayFlags_Set(b *testing.B) {
	// 初始化：创建非nil的ArrayFlags（模拟正常使用场景）
	var af ArrayFlags
	// 重置计时器：排除初始化耗时
	b.ResetTimer()

	// 执行b.N次Set调用（b.N由Go测试框架自动调整，保证测试准确性）
	for i := 0; i < b.N; i++ {
		// 每次传入不同的IP，模拟真实场景（避免编译器优化）
		_ = af.Set(fmt.Sprintf("192.168.0.%d", i%255))
	}
}

// BenchmarkArrayFlags_String_Empty 基准测试：空切片场景下String()方法的性能
func BenchmarkArrayFlags_String_Empty(b *testing.B) {
	af := &ArrayFlags{} // 空切片（非nil）
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = af.String()
	}
}

// BenchmarkArrayFlags_String_WithElements 基准测试：有元素场景下String()方法的性能（临界值：2个元素）
func BenchmarkArrayFlags_String_WithElements(b *testing.B) {
	// 初始化：填充2个元素（模拟实际使用场景）
	af := &ArrayFlags{"192.168.0.55", "192.168.0.56"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = af.String()
	}
}

// BenchmarkArrayFlags_Set_Nil 基准测试：nil指针场景下Set()方法的性能（临界场景）
func BenchmarkArrayFlags_Set_Nil(b *testing.B) {
	var af *ArrayFlags = nil // 显式置nil
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = af.Set("192.168.0.55")
	}
}
