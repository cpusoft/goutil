package randutil

import (
	"fmt"
	"math"
	"sync"
	"testing"
)

// -------------------------- 功能测试 & 临界值测试 --------------------------
// TestIntn 测试Intn函数的功能正确性和临界值
func TestIntn(t *testing.T) {
	// 定义测试用例：子测试名称 -> (输入n, 预期是否panic, 预期结果范围)
	testCases := []struct {
		name      string
		n         uint
		wantPanic bool
		min       int // 结果最小值（包含）
		max       int // 结果最大值（不包含）
	}{
		{
			name:      "n=0（临界值，预期panic）",
			n:         0,
			wantPanic: true,
			min:       0,
			max:       0,
		},
		{
			name:      "n=1（临界值，只能返回0）",
			n:         1,
			wantPanic: false,
			min:       0,
			max:       1,
		},
		{
			name:      "n=100（正常范围）",
			n:         100,
			wantPanic: false,
			min:       0,
			max:       100,
		},
		{
			name:      "n=最大int值（大数临界值）",
			n:         uint(math.MaxInt),
			wantPanic: false,
			min:       0,
			max:       math.MaxInt,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 处理panic场景
			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("调用Intn(%d)预期panic，但未发生", tc.n)
					}
				}()
				Intn(tc.n)
				return
			}

			// 非panic场景：多次调用验证结果在范围内
			const callTimes = 1000
			for i := 0; i < callTimes; i++ {
				res := Intn(tc.n)
				if res < tc.min || res >= tc.max {
					t.Errorf("第%d次调用Intn(%d)返回%d，超出范围[%d, %d)", i, tc.n, res, tc.min, tc.max)
				}
			}
		})
	}
}

// TestIntRange 测试IntRange函数的功能正确性和临界值
func TestIntRange(t *testing.T) {
	// 定义测试用例：子测试名称 -> (min, n, 预期是否panic, 预期结果范围)
	testCases := []struct {
		name      string
		min       uint
		n         uint
		wantPanic bool
		minRes    int // 结果最小值（包含）
		maxRes    int // 结果最大值（不包含）
	}{
		{
			name:      "min=0, n=0（临界值，预期panic）",
			min:       0,
			n:         0,
			wantPanic: true,
			minRes:    0,
			maxRes:    0,
		},
		{
			name:      "min=0, n=1（临界值，只能返回0）",
			min:       0,
			n:         1,
			wantPanic: false,
			minRes:    0,
			maxRes:    1,
		},
		{
			name:      "min=10, n=1（临界值，只能返回10）",
			min:       10,
			n:         1,
			wantPanic: false,
			minRes:    10,
			maxRes:    11,
		},
		{
			name:      "min=5, n=10（正常范围）",
			min:       5,
			n:         10,
			wantPanic: false,
			minRes:    5,
			maxRes:    15,
		},
		{
			name:      "大数临界值（min=MaxInt-5, n=5）",
			min:       uint(math.MaxInt - 5),
			n:         5,
			wantPanic: false,
			minRes:    math.MaxInt - 5,
			maxRes:    math.MaxInt,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 处理panic场景
			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("调用IntRange(%d, %d)预期panic，但未发生", tc.min, tc.n)
					}
				}()
				IntRange(tc.min, tc.n)
				return
			}

			// 非panic场景：多次调用验证结果在范围内
			const callTimes = 1000
			for i := 0; i < callTimes; i++ {
				res := IntRange(tc.min, tc.n)
				if res < tc.minRes || res >= tc.maxRes {
					t.Errorf("第%d次调用IntRange(%d, %d)返回%d，超出范围[%d, %d)", i, tc.min, tc.n, res, tc.minRes, tc.maxRes)
				}
			}
		})
	}
}

// TestIntnConcurrent 测试Intn并发安全性（无数据竞争、无panic、结果合法）
func TestIntnConcurrent(t *testing.T) {
	const (
		goroutineNum     = 100  // 并发协程数
		callPerGoroutine = 1000 // 每个协程调用次数
		n                = 1000 // 测试用的随机数上限
	)

	var wg sync.WaitGroup
	wg.Add(goroutineNum)

	// 启动多协程调用Intn
	for i := 0; i < goroutineNum; i++ {
		go func(gid int) {
			defer wg.Done()
			for j := 0; j < callPerGoroutine; j++ {
				res := Intn(n)
				if res < 0 || res >= int(n) {
					t.Errorf("协程%d第%d次调用Intn(%d)返回%d，超出范围[0, %d)", gid, j, n, res, n)
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestIntRangeConcurrent 测试IntRange并发安全性
func TestIntRangeConcurrent(t *testing.T) {
	const (
		goroutineNum     = 100
		callPerGoroutine = 1000
		min              = 100
		n                = 200
	)

	var wg sync.WaitGroup
	wg.Add(goroutineNum)

	for i := 0; i < goroutineNum; i++ {
		go func(gid int) {
			defer wg.Done()
			for j := 0; j < callPerGoroutine; j++ {
				res := IntRange(min, n)
				if res < int(min) || res >= int(min)+int(n) {
					t.Errorf("协程%d第%d次调用IntRange(%d, %d)返回%d，超出范围[%d, %d)", gid, j, min, n, res, min, min+n)
				}
			}
		}(i)
	}

	wg.Wait()
}

// -------------------------- 性能测试 --------------------------
// BenchmarkIntn 单协程下Intn的性能测试
func BenchmarkIntn(b *testing.B) {
	// 重置计时器（排除初始化耗时）
	b.ResetTimer()
	// 执行b.N次（测试框架自动调整b.N以获得稳定结果）
	for i := 0; i < b.N; i++ {
		Intn(1000)
	}
}

// BenchmarkIntn_Concurrent 多协程下Intn的性能测试
func BenchmarkIntn_Concurrent(b *testing.B) {
	// 设置并发数为GOMAXPROCS（CPU核心数）
	b.SetParallelism(10)
	// 启用并发测试
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Intn(1000)
		}
	})
}

// BenchmarkIntRange 单协程下IntRange的性能测试
func BenchmarkIntRange(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IntRange(100, 200)
	}
}

// BenchmarkIntRange_Concurrent 多协程下IntRange的性能测试
func BenchmarkIntRange_Concurrent(b *testing.B) {
	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			IntRange(100, 200)
		}
	})
}

func TestIntn1(t *testing.T) {
	a := Intn(2)
	b := fmt.Sprintf("%s?%d", "http://www.aaa.com/", Intn(1))
	fmt.Println(a, b)
}
