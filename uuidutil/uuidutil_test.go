package uuidutil

import (
	"fmt"
	"regexp"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 验证UUID V4格式的正则（标准格式：8-4-4-4-12，第13位固定为4）
var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// TestGetUuid_Normal 正常场景测试：验证真实生成的UUID非空、格式符合V4规范
// 无mock，直接调用真实的uuid.NewV4
func TestGetUuid_Normal(t *testing.T) {
	// 增加调用次数（200次），提升测试可信度
	testCount := 200
	for i := 0; i < testCount; i++ {
		uuidStr := GetUuid()
		// 断言UUID非空（正常环境下uuid.NewV4几乎不会失败）
		assert.NotEmpty(t, uuidStr, "第%d次生成的UUID为空（正常环境下不应出现）", i)
		// 断言UUID严格符合V4格式规范
		assert.True(t, uuidV4Pattern.MatchString(uuidStr),
			"第%d次生成的UUID格式不符合V4规范：%s", i, uuidStr)
	}
}

// TestGetUuid_Concurrent 临界值测试：高并发调用（2000个goroutine）
// 无mock，验证真实场景下并发安全、UUID无重复、无空值
func TestGetUuid_Concurrent(t *testing.T) {
	const (
		goroutineCount   = 2000                              // 并发协程数（提升临界值压力）
		callPerGoroutine = 100                               // 每个协程调用次数
		totalCallCount   = goroutineCount * callPerGoroutine // 总调用次数
	)

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex                                  // 保护共享数据的并发读写
		uuidSet    = make(map[string]struct{}, totalCallCount) // 存储生成的UUID，验证唯一性
		emptyCount int                                         // 统计空UUID数量（正常应为0）
	)

	// 启动所有协程
	wg.Add(goroutineCount)
	for gid := 0; gid < goroutineCount; gid++ {
		go func(gid int) {
			defer wg.Done()
			// 每个协程多次调用
			for callIdx := 0; callIdx < callPerGoroutine; callIdx++ {
				uuidStr := GetUuid()
				mu.Lock() // 加锁操作共享数据
				if uuidStr == "" {
					emptyCount++
					t.Logf("协程%d第%d次调用生成空UUID（异常，正常环境下不应出现）", gid, callIdx)
				} else {
					// 断言UUID未重复（V4 UUID重复概率约为1/(2^122)，理论上不会触发）
					if _, exists := uuidSet[uuidStr]; exists {
						t.Errorf("协程%d第%d次调用生成重复UUID：%s", gid, callIdx, uuidStr)
					}
					uuidSet[uuidStr] = struct{}{}
				}
				mu.Unlock()
			}
		}(gid)
	}

	// 等待所有协程完成
	wg.Wait()

	// 断言无空UUID（验证函数在高并发下仍稳定）
	assert.Zero(t, emptyCount, "高并发场景下生成了%d个空UUID（异常）", emptyCount)
	// 断言生成的UUID总数与预期一致（无重复、无丢失）
	assert.Equal(t, totalCallCount, len(uuidSet),
		"生成的UUID总数不符：预期%d，实际%d", totalCallCount, len(uuidSet))
}

// BenchmarkGetUuid 单协程性能测试：基于真实UUID生成逻辑评估性能
func BenchmarkGetUuid(b *testing.B) {
	// 重置计时器，排除初始化耗时
	b.ResetTimer()
	// b.N由测试框架自动调整（保证测试结果稳定）
	for i := 0; i < b.N; i++ {
		GetUuid()
	}
}

// BenchmarkGetUuid_Concurrent 并发性能测试：多协程下真实场景的性能表现
func BenchmarkGetUuid_Concurrent(b *testing.B) {
	// 按当前CPU核心数并发执行，模拟真实高并发场景
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GetUuid()
		}
	})
}

// /////////////////////////////////////////////
func TestGetUuid(t *testing.T) {
	u := GetUuid()
	fmt.Println(u)
}
