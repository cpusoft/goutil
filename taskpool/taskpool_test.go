package taskpool

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskPool_InvalidSize(t *testing.T) {
	pool, err := NewTaskPool(0)
	assert.Nil(t, pool)
	assert.Error(t, err)
}

func TestNewTaskPool_ValidSize(t *testing.T) {
	pool, err := NewTaskPool(5)
	assert.NoError(t, err)
	assert.NotNil(t, pool)
	defer pool.Release()
}

func TestSubmit_TaskSuccess(t *testing.T) {
	pool, err := NewTaskPool(1)
	assert.NoError(t, err)
	defer pool.Release()

	var wg sync.WaitGroup
	wg.Add(1)

	err = pool.Submit(func() error {
		wg.Done()
		return nil
	}, 0)

	wg.Wait()
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // 等待统计更新

	stats := pool.GetStats()
	assert.Equal(t, int64(1), stats.SubmittedTasks)
	assert.Equal(t, int64(1), stats.CompletedTasks)
	assert.Equal(t, int64(0), stats.FailedTasks)
	assert.Equal(t, int64(0), stats.RunningTasks)
}

// TestSubmit_TaskFailureWithRetrySuccess tests task failure but success on retry using SubmitSync
func TestSubmit_TaskFailureWithRetrySuccess(t *testing.T) {
	pool, err := NewTaskPool(1)
	assert.NoError(t, err)
	defer pool.Release()

	callCount := 0
	var mu sync.Mutex

	config := TaskConfig{
		MaxRetries: 2,
		Timeout:    5 * time.Second,
		Priority:   PriorityNormal,
	}

	result, err := pool.SubmitSync(func() error {
		mu.Lock()
		callCount++
		currentCall := callCount
		mu.Unlock()

		if currentCall == 2 {
			return nil //
		}
		return errors.New("simulated failure")
	}, config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, 2, result.Attempts)
	assert.NoError(t, result.Error)

	time.Sleep(100 * time.Millisecond)

	stats := pool.GetStats()
	assert.Equal(t, int64(1), stats.SubmittedTasks)
	assert.Equal(t, int64(1), stats.CompletedTasks)
	assert.Equal(t, int64(0), stats.FailedTasks)
	assert.Equal(t, int64(0), stats.RunningTasks)
	assert.Equal(t, int64(1), stats.TotalRetries) //
}

// TestSubmit_TaskFailureAfterAllRetries tests all retries fail using SubmitSync
func TestSubmit_TaskFailureAfterAllRetries(t *testing.T) {
	pool, err := NewTaskPool(1)
	assert.NoError(t, err)
	defer pool.Release()

	config := TaskConfig{
		MaxRetries: 2,
		Timeout:    5 * time.Second,
		Priority:   PriorityNormal,
	}

	result, err := pool.SubmitSync(func() error {
		return errors.New("always fails")
	}, config)

	assert.NoError(t, err) //
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, 3, result.Attempts) // 1次初始尝试 + 2次重试
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "always fails")

	time.Sleep(100 * time.Millisecond)

	stats := pool.GetStats()
	assert.Equal(t, int64(1), stats.SubmittedTasks)
	assert.Equal(t, int64(1), stats.CompletedTasks)
	assert.Equal(t, int64(1), stats.FailedTasks) //
	assert.Equal(t, int64(0), stats.RunningTasks)
	assert.Equal(t, int64(2), stats.TotalRetries) //
}

func TestSubmit_PoolClosedBeforeSubmit(t *testing.T) {
	pool, err := NewTaskPool(1)
	assert.NoError(t, err)
	pool.Release()

	err = pool.Submit(func() error {
		return nil
	}, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task pool is shutting down")
}

func TestSubmitWithConfig(t *testing.T) {
	pool, err := NewTaskPool(2)
	assert.NoError(t, err)
	defer pool.Release()

	var wg sync.WaitGroup
	wg.Add(1)

	config := TaskConfig{
		MaxRetries: 1,
		Timeout:    2 * time.Second,
		Priority:   PriorityHigh,
	}

	err = pool.SubmitWithConfig(func() error {
		wg.Done()
		return nil
	}, config)

	wg.Wait()
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	stats := pool.GetStats()
	assert.Equal(t, int64(1), stats.SubmittedTasks)
	assert.Equal(t, int64(1), stats.CompletedTasks)
	assert.Equal(t, int64(1), stats.HighPriorityTasks)
}

func TestSubmitSync(t *testing.T) {
	pool, err := NewTaskPool(1)
	assert.NoError(t, err)
	defer pool.Release()

	config := TaskConfig{
		MaxRetries: 0,
		Timeout:    2 * time.Second,
		Priority:   PriorityNormal,
	}

	result, err := pool.SubmitSync(func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}, config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, 1, result.Attempts)
	assert.NoError(t, result.Error)
	assert.True(t, result.Duration > 0)
}

func TestTaskTimeout(t *testing.T) {
	pool, err := NewTaskPool(1)
	assert.NoError(t, err)
	defer pool.Release()

	config := TaskConfig{
		MaxRetries: 0,
		Timeout:    100 * time.Millisecond,
		Priority:   PriorityNormal,
	}

	result, err := pool.SubmitSync(func() error {
		time.Sleep(200 * time.Millisecond)
		return nil
	}, config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, 1, result.Attempts)
	assert.Error(t, result.Error)

	time.Sleep(100 * time.Millisecond)

	stats := pool.GetStats()
	assert.Equal(t, int64(1), stats.TimeoutTasks)
}

func TestResize(t *testing.T) {
	pool, err := NewTaskPool(1)
	assert.NoError(t, err)
	defer pool.Release()

	err = pool.Resize(5)
	assert.NoError(t, err)
	assert.Equal(t, 5, pool.pool.Cap())
}

func TestGetStats_ConcurrentAccess(t *testing.T) {
	pool, err := NewTaskPool(2)
	assert.NoError(t, err)
	defer pool.Release()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.Submit(func() error {
				time.Sleep(10 * time.Millisecond)
				return nil
			}, 0)
		}()
	}
	wg.Wait()

	time.Sleep(200 * time.Millisecond)

	stats := pool.GetStats()
	assert.Equal(t, int64(10), stats.SubmittedTasks)
	assert.Equal(t, int64(10), stats.CompletedTasks)
	assert.Equal(t, int64(0), stats.FailedTasks)
	assert.Equal(t, int64(0), stats.RunningTasks)
}

func TestPriority(t *testing.T) {
	pool, err := NewTaskPool(1)
	assert.NoError(t, err)
	defer pool.Release()

	var results []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 低优先级
	wg.Add(1)
	pool.SubmitWithConfig(func() error {
		mu.Lock()
		results = append(results, "low")
		mu.Unlock()
		wg.Done()
		return nil
	}, TaskConfig{Priority: PriorityLow})

	// 高优先级
	wg.Add(1)
	pool.SubmitWithConfig(func() error {
		mu.Lock()
		results = append(results, "high")
		mu.Unlock()
		wg.Done()
		return nil
	}, TaskConfig{Priority: PriorityHigh})

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	stats := pool.GetStats()
	assert.Equal(t, int64(1), stats.HighPriorityTasks)
	assert.Equal(t, int64(1), stats.LowPriorityTasks)
}

func TestRelease(t *testing.T) {
	pool, err := NewTaskPool(1)
	assert.NoError(t, err)

	pool.Release()

	time.Sleep(100 * time.Millisecond)

	err = pool.Submit(func() error {
		return nil
	}, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task pool is shutting down")
}

// TestShutdownWithTimeout tests graceful shutdown with timeout
func TestShutdownWithTimeout(t *testing.T) {
	pool, err := NewTaskPool(2)
	assert.NoError(t, err)

	var completed int32
	for i := 0; i < 3; i++ {
		pool.Submit(func() error {
			atomic.AddInt32(&completed, 1)
			return nil
		}, 0)
	}

	time.Sleep(100 * time.Millisecond)

	start := time.Now()
	pool.ShutdownWithTimeout(1 * time.Second)
	duration := time.Since(start)

	for i := 0; i < 100; i++ {
		if atomic.LoadInt32(&completed) == 3 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	assert.True(t, duration < 1*time.Second)
	assert.Equal(t, int32(3), atomic.LoadInt32(&completed))
}
