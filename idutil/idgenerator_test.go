package idutil

import (
	"fmt"
	"sync"
	"testing"
)

// -------------------------- 基础功能 & 临界值测试 --------------------------

// TestGenerateSnowflakeInt64 测试雪花ID（int64类型）的基础功能和临界值
func TestGenerateSnowflakeInt64(t *testing.T) {
	tests := []struct {
		name    string
		nodeId  int64
		wantErr bool // 是否期望报错
	}{
		{
			name:    "合法nodeId-常规值",
			nodeId:  1,
			wantErr: false,
		},
		{
			name:    "合法nodeId-临界最大值（snowflake nodeId最大为1023，10位）",
			nodeId:  1023,
			wantErr: false,
		},
		{
			name:    "非法nodeId-超过临界值（1024）",
			nodeId:  1024,
			wantErr: true,
		},
		{
			name:    "非法nodeId-负数",
			nodeId:  -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 GenerateSnowflakeInt64（主函数）
			id, err := GenerateSnowflakeInt64(tt.nodeId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSnowflakeInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id == 0 {
				t.Error("GenerateSnowflakeInt64() 生成的ID为0，不符合预期")
			}

			// 测试 Deprecated 的 GenerateInt64（保证兼容）
			id2, err2 := GenerateInt64(tt.nodeId)
			if (err2 != nil) != tt.wantErr {
				t.Errorf("GenerateInt64() error = %v, wantErr %v", err2, tt.wantErr)
				return
			}
			if !tt.wantErr && id2 == 0 {
				t.Error("GenerateInt64() 生成的ID为0，不符合预期")
			}
		})
	}

	// 临界场景：同一nodeId、同一毫秒内高频调用，验证ID重复（原代码的已知风险）
	t.Run("同一毫秒高频调用-验证ID重复", func(t *testing.T) {
		nodeId := int64(1)
		idMap := make(map[int64]struct{}, 100)
		duplicateCount := 0

		// 短时间内生成100个ID，验证重复
		for i := 0; i < 100; i++ {
			id, err := GenerateSnowflakeInt64(nodeId)
			if err != nil {
				t.Fatalf("生成ID失败: %v", err)
			}
			if _, exists := idMap[id]; exists {
				duplicateCount++
			} else {
				idMap[id] = struct{}{}
			}
		}

		// 打印重复数量（原代码逻辑下大概率会有重复）
		t.Logf("同一毫秒内生成100个ID，重复数量: %d", duplicateCount)
		if duplicateCount > 0 {
			t.Log("注意：原代码每次新建snowflake节点，同一毫秒内会生成重复ID（已知风险）")
		}
	})
}

// TestGenerateSnowflakeString 测试雪花ID（string类型）的基础功能和临界值
func TestGenerateSnowflakeString(t *testing.T) {
	tests := []struct {
		name    string
		nodeId  int64
		wantErr bool // 是否期望报错
	}{
		{
			name:    "合法nodeId-常规值",
			nodeId:  2,
			wantErr: false,
		},
		{
			name:    "合法nodeId-临界最大值",
			nodeId:  1023,
			wantErr: false,
		},
		{
			name:    "非法nodeId-超过临界值",
			nodeId:  1024,
			wantErr: true,
		},
		{
			name:    "非法nodeId-负数",
			nodeId:  -2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 GenerateSnowflakeString（主函数）
			id, err := GenerateSnowflakeString(tt.nodeId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSnowflakeString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id == "" {
				t.Error("GenerateSnowflakeString() 生成的ID为空字符串，不符合预期")
			}

			// 测试 Deprecated 的 GenerateString（保证兼容）
			id2, err2 := GenerateString(tt.nodeId)
			if (err2 != nil) != tt.wantErr {
				t.Errorf("GenerateString() error = %v, wantErr %v", err2, tt.wantErr)
				return
			}
			if !tt.wantErr && id2 == "" {
				t.Error("GenerateString() 生成的ID为空字符串，不符合预期")
			}
		})
	}
}

// TestGenerateSequentialUint64UUID 测试Uint64 UUID的基础功能和唯一性
func TestGenerateSequentialUint64UUID(t *testing.T) {
	// 基础功能：生成的UUID非0
	t.Run("基础功能-非0验证", func(t *testing.T) {
		uuid := GenerateSequentialUint64UUID()
		if uuid == 0 {
			t.Error("GenerateSequentialUint64UUID() 生成的UUID为0，不符合预期")
		}
	})

	// 临界场景：高频生成验证唯一性（随机数种子已初始化，应无重复）
	t.Run("高频生成-唯一性验证", func(t *testing.T) {
		uuidMap := make(map[uint64]struct{}, 10000)
		duplicateCount := 0

		for i := 0; i < 10000; i++ {
			uuid := GenerateSequentialUint64UUID()
			if _, exists := uuidMap[uuid]; exists {
				duplicateCount++
			} else {
				uuidMap[uuid] = struct{}{}
			}
		}

		t.Logf("生成10000个Uint64 UUID，重复数量: %d", duplicateCount)
		if duplicateCount > 0 {
			t.Errorf("GenerateSequentialUint64UUID() 生成了重复的UUID，重复数量: %d", duplicateCount)
		}
	})

	// 并发场景：验证多goroutine生成的唯一性
	t.Run("并发生成-唯一性验证", func(t *testing.T) {
		var wg sync.WaitGroup
		uuidMap := sync.Map{}
		duplicateCount := 0
		total := 1000

		wg.Add(total)
		for i := 0; i < total; i++ {
			go func() {
				defer wg.Done()
				uuid := GenerateSequentialUint64UUID()
				if _, exists := uuidMap.LoadOrStore(uuid, struct{}{}); exists {
					duplicateCount++
				}
			}()
		}
		wg.Wait()

		t.Logf("并发生成%d个Uint64 UUID，重复数量: %d", total, duplicateCount)
		if duplicateCount > 0 {
			t.Errorf("并发生成UUID出现重复，重复数量: %d", duplicateCount)
		}
	})
}

// -------------------------- 性能测试（Benchmark） --------------------------

// BenchmarkGenerateSnowflakeInt64 测试雪花ID（int64）的生成性能
func BenchmarkGenerateSnowflakeInt64(b *testing.B) {
	nodeId := int64(1)
	// 重置计时器（忽略初始化耗时）
	b.ResetTimer()

	// 普通单协程性能
	b.Run("single", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := GenerateSnowflakeInt64(nodeId)
			if err != nil {
				b.Fatalf("生成ID失败: %v", err)
			}
		}
	})

	// 并发性能
	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := GenerateSnowflakeInt64(nodeId)
				if err != nil {
					b.Fatalf("生成ID失败: %v", err)
				}
			}
		})
	})
}

// BenchmarkGenerateSnowflakeString 测试雪花ID（string）的生成性能
func BenchmarkGenerateSnowflakeString(b *testing.B) {
	nodeId := int64(2)
	b.ResetTimer()

	// 普通单协程性能
	b.Run("single", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := GenerateSnowflakeString(nodeId)
			if err != nil {
				b.Fatalf("生成ID失败: %v", err)
			}
		}
	})

	// 并发性能
	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := GenerateSnowflakeString(nodeId)
				if err != nil {
					b.Fatalf("生成ID失败: %v", err)
				}
			}
		})
	})
}

// BenchmarkGenerateSequentialUint64UUID 测试Uint64 UUID的生成性能
func BenchmarkGenerateSequentialUint64UUID(b *testing.B) {
	b.ResetTimer()

	// 普通单协程性能
	b.Run("single", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GenerateSequentialUint64UUID()
		}
	})

	// 并发性能
	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = GenerateSequentialUint64UUID()
			}
		})
	})
}

// -------------------------- 辅助测试：打印ID结构（可选） --------------------------
func ExampleGenerateSnowflakeInt64() {
	id, err := GenerateSnowflakeInt64(1)
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
		return
	}
	fmt.Printf("生成的雪花ID（int64）: %d\n", id)
	// 输出示例：生成的雪花ID（int64）: 1234567890123456789
}

func ExampleGenerateSequentialUint64UUID() {
	uuid := GenerateSequentialUint64UUID()
	fmt.Printf("生成的Uint64 UUID: %d\n", uuid)
	// 输出示例：生成的Uint64 UUID: 1711234567890123456
}
