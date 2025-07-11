package taskpool

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

// 封装了几个方法
// 1.Submit    提交任务
// 2.SubmitWithConfig    提交任务带配置
// 3.SubmitSync        提交任务后同步等待结果

//  --------------------   1. Submit    -------------
//pool, _ := NewTaskPool(5)
//err := pool.Submit(func() error {
//	return nil
//}, 0)

//  --------------------   2. SubmitWithConfig    -------------
//config := taskpool.TaskConfig{
//MaxRetries: 3,
//Timeout:    time.Second * 5,
//Priority:   taskpool.PriorityHigh,
//}
//
//err := pool.SubmitWithConfig(func() error {
//	time.Sleep(time.Second * 3)
//	return nil
//}, config)

//  --------------------   3. SubmitWithConfig    -------------
//result, err := pool.SubmitSync(func() error {
//	time.Sleep(time.Millisecond * 500)
//	return errors.New("something went wrong")
//}, taskpool.DefaultTaskConfig)
//
//if result != nil {
//fmt.Printf("Attempts: %d, Success: %v, Error: %v\n", result.Attempts, result.Success, result.Error)
//}

// TaskPriority
type TaskPriority int

const (
	PriorityLow    TaskPriority = 0
	PriorityNormal TaskPriority = 1
	PriorityHigh   TaskPriority = 2
)

// TaskConfig
type TaskConfig struct {
	MaxRetries int //
	Timeout    time.Duration
	Priority   TaskPriority
}

// DefaultTaskConfig 默认任务配置
var DefaultTaskConfig = TaskConfig{
	MaxRetries: 0,
	Timeout:    0, // 0表示无超时
	Priority:   PriorityNormal,
}

// TaskResult
type TaskResult struct {
	Success   bool
	Error     error
	Attempts  int           //
	Duration  time.Duration //
	StartTime time.Time
	EndTime   time.Time
}

// Task
type Task struct {
	ID       string
	Func     func() error
	Config   TaskConfig
	ResultCh chan *TaskResult
}

// TaskPool 任务池
type TaskPool struct {
	pool                *ants.Pool
	ctx                 context.Context
	cancelFunc          context.CancelFunc
	mu                  sync.RWMutex
	taskStats           Stats
	taskCounter         int64
	highPriorityTasks   chan *Task
	normalPriorityTasks chan *Task
	lowPriorityTasks    chan *Task
	workerWg            sync.WaitGroup
	once                sync.Once
}

// Stats
type Stats struct {
	SubmittedTasks      int64
	CompletedTasks      int64
	FailedTasks         int64
	RunningTasks        int64
	TotalRetries        int64
	TimeoutTasks        int64
	HighPriorityTasks   int64
	NormalPriorityTasks int64
	LowPriorityTasks    int64
}

// NewTaskPool
func NewTaskPool(size int) (*TaskPool, error) {
	if size <= 0 {
		return nil, errors.New("pool size must be greater than 0")
	}
	p, err := ants.NewPool(size)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	tp := &TaskPool{
		pool:                p,
		ctx:                 ctx,
		cancelFunc:          cancel,
		highPriorityTasks:   make(chan *Task, size*2),
		normalPriorityTasks: make(chan *Task, size*4),
		lowPriorityTasks:    make(chan *Task, size*2),
	}

	tp.startTaskDispatcher()

	return tp, nil
}

// startTaskDispatcher
func (tp *TaskPool) startTaskDispatcher() {
	tp.workerWg.Add(1)
	go func() {
		defer tp.workerWg.Done()
		for {
			select {
			case <-tp.ctx.Done():
				return
			case task := <-tp.highPriorityTasks:
				tp.executeTask(task)
			case task := <-tp.normalPriorityTasks:
				tp.executeTask(task)
			case task := <-tp.lowPriorityTasks:
				tp.executeTask(task)
			}
		}
	}()
}

// executeTask
func (tp *TaskPool) executeTask(task *Task) {
	tp.pool.Submit(func() {
		result := tp.runTaskWithRetry(task)
		if task.ResultCh != nil {
			task.ResultCh <- result
		}
	})
}

// runTaskWithRetry
func (tp *TaskPool) runTaskWithRetry(task *Task) *TaskResult {
	result := &TaskResult{
		StartTime: time.Now(),
	}

	// 更新运行中任务数
	tp.mu.Lock()
	tp.taskStats.RunningTasks++
	tp.mu.Unlock()

	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)

		tp.mu.Lock()
		tp.taskStats.RunningTasks--
		tp.taskStats.CompletedTasks++
		if !result.Success {
			tp.taskStats.FailedTasks++
		}
		tp.taskStats.TotalRetries += int64(result.Attempts - 1)
		tp.mu.Unlock()
	}()

	maxAttempts := task.Config.MaxRetries + 1
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result.Attempts = attempt

		select {
		case <-tp.ctx.Done():
			result.Error = errors.New("task pool is shutting down")
			return result
		default:
		}

		taskCtx := tp.ctx
		if task.Config.Timeout > 0 {
			var cancel context.CancelFunc
			taskCtx, cancel = context.WithTimeout(tp.ctx, task.Config.Timeout)
			defer cancel()
		}

		err := tp.executeWithTimeout(taskCtx, task.Func)

		if err == nil {
			result.Success = true
			result.Error = nil //
			return result
		}

		if errors.Is(err, context.DeadlineExceeded) {
			tp.mu.Lock()
			tp.taskStats.TimeoutTasks++
			tp.mu.Unlock()
		}

		result.Error = err

		if attempt < maxAttempts {
			select {
			case <-tp.ctx.Done():
				result.Error = errors.New("task pool is shutting down")
				return result
			case <-time.After(time.Duration(attempt) * 100 * time.Millisecond):
				// 退避重试
			}
		}
	}

	return result
}

// executeWithTimeout
func (tp *TaskPool) executeWithTimeout(ctx context.Context, taskFunc func() error) error {
	if ctx == tp.ctx {

		return taskFunc()
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- taskFunc()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Submit
func (tp *TaskPool) Submit(task func() error, retryCount int) error {
	config := DefaultTaskConfig
	config.MaxRetries = retryCount
	return tp.SubmitWithConfig(task, config)
}

// SubmitWithConfig
func (tp *TaskPool) SubmitWithConfig(taskFunc func() error, config TaskConfig) error {
	select {
	case <-tp.ctx.Done():
		return errors.New("task pool is shutting down")
	default:
	}

	tp.mu.Lock()
	tp.taskCounter++
	taskID := tp.taskCounter
	tp.taskStats.SubmittedTasks++

	// 更新优先级统计
	switch config.Priority {
	case PriorityHigh:
		tp.taskStats.HighPriorityTasks++
	case PriorityNormal:
		tp.taskStats.NormalPriorityTasks++
	case PriorityLow:
		tp.taskStats.LowPriorityTasks++
	}
	tp.mu.Unlock()

	task := &Task{
		ID:     string(rune(taskID)),
		Func:   taskFunc,
		Config: config,
	}

	var taskChan chan *Task
	switch config.Priority {
	case PriorityHigh:
		taskChan = tp.highPriorityTasks
	case PriorityNormal:
		taskChan = tp.normalPriorityTasks
	case PriorityLow:
		taskChan = tp.lowPriorityTasks
	}

	select {
	case taskChan <- task:
		return nil
	case <-tp.ctx.Done():
		return errors.New("task pool is shutting down")
	}
}

// SubmitSync
func (tp *TaskPool) SubmitSync(taskFunc func() error, config TaskConfig) (*TaskResult, error) {
	select {
	case <-tp.ctx.Done():
		return nil, errors.New("task pool is shutting down")
	default:
	}

	tp.mu.Lock()
	tp.taskCounter++
	taskID := tp.taskCounter
	tp.taskStats.SubmittedTasks++

	switch config.Priority {
	case PriorityHigh:
		tp.taskStats.HighPriorityTasks++
	case PriorityNormal:
		tp.taskStats.NormalPriorityTasks++
	case PriorityLow:
		tp.taskStats.LowPriorityTasks++
	}
	tp.mu.Unlock()

	task := &Task{
		ID:       string(rune(taskID)),
		Func:     taskFunc,
		Config:   config,
		ResultCh: make(chan *TaskResult, 1),
	}

	var taskChan chan *Task
	switch config.Priority {
	case PriorityHigh:
		taskChan = tp.highPriorityTasks
	case PriorityNormal:
		taskChan = tp.normalPriorityTasks
	case PriorityLow:
		taskChan = tp.lowPriorityTasks
	}

	select {
	case taskChan <- task:

		select {
		case result := <-task.ResultCh:
			return result, nil
		case <-tp.ctx.Done():
			return nil, errors.New("task pool is shutting down")
		}
	case <-tp.ctx.Done():
		return nil, errors.New("task pool is shutting down")
	}
}

// Resize
func (tp *TaskPool) Resize(newSize int) error {
	tp.pool.Tune(newSize)
	return nil
}

// GetStats
func (tp *TaskPool) GetStats() Stats {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.taskStats
}

// ShutdownWithTimeout
func (tp *TaskPool) ShutdownWithTimeout(timeout time.Duration) {
	tp.once.Do(func() {

		tp.cancelFunc()

		done := make(chan struct{})
		go func() {
			tp.workerWg.Wait()
			close(done)
		}()

		if timeout > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			select {
			case <-done:

			case <-ctx.Done():

			}
		} else {
			<-done
		}

		// 关闭任务队列
		close(tp.highPriorityTasks)
		close(tp.normalPriorityTasks)
		close(tp.lowPriorityTasks)

		tp.pool.Release()
	})
}

// Release 关闭任务池并释放所有资源
func (tp *TaskPool) Release() {
	tp.ShutdownWithTimeout(0)
}
