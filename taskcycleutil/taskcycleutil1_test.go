package taskcycleutil

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ========== 测试工具函数 ==========
// 生成测试任务
func generateTestTask(key string) *Task {
	return &Task{
		Key: key,
		TaskParam: TaskParam{
			Content: "test content",
			Params:  map[string]interface{}{"test": "param"},
		},
	}
}

// 简单执行函数：立即成功
func successExecuteFunc(ctx context.Context, task *Task) TaskExecutionResult {
	return TaskExecutionResult{
		Result: TaskResultOK,
		Err:    "",
	}
}

// 简单执行函数：立即失败
func failExecuteFunc(ctx context.Context, task *Task) TaskExecutionResult {
	return TaskExecutionResult{
		Result: TaskResultFail,
		Err:    "test fail",
	}
}

// 等待任务达到指定状态
func waitForTaskState(framework *TaskFramework, key string, targetState string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		framework.mu.RLock()
		task, exists := framework.tasks[key]
		framework.mu.RUnlock()

		if exists && task.TaskState == targetState {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// 等待任务结果计数达到指定值
func waitForTaskResultCount(framework *TaskFramework, key string, targetCount uint64, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		framework.mu.RLock()
		task, exists := framework.tasks[key]
		framework.mu.RUnlock()

		if exists && task.TaskResult.ResultCount == targetCount {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// 修复超时执行函数（确保只触发一次超时）
func timeoutExecuteFunc(ctx context.Context, task *Task) TaskExecutionResult {
	// 无限等待直到上下文取消（确保超时，且只执行一次）
	<-ctx.Done()
	return TaskExecutionResult{
		Result: TaskResultFail,
		Err:    ctx.Err().Error(),
	}
}

// 上下文取消执行函数：检测框架停止
func ctxCancelExecuteFunc(ctx context.Context, task *Task) TaskExecutionResult {
	select {
	case <-ctx.Done():
		return TaskExecutionResult{
			Result: TaskResultFail,
			Err:    ctx.Err().Error(),
		}
	case <-time.After(100 * time.Millisecond):
		return TaskExecutionResult{
			Result: TaskResultOK,
			Err:    "",
		}
	}
}

// 递归生成任务函数：从成功任务生成新任务
func testGenerateTasksFunc(completedTask *Task) []*Task {
	newKey := completedTask.Key + "_generated"
	return []*Task{generateTestTask(newKey)}
}

// 等待框架中完成的任务总数达到指定值
func waitForTaskResultCountTotal(framework *TaskFramework, targetCount int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		framework.mu.RLock()
		count := 0
		for _, task := range framework.tasks {
			if task.TaskResult.ResultCount > 0 {
				count++
			}
		}
		framework.mu.RUnlock()

		if count >= targetCount {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// ========== 配置初始化测试 ==========
func TestNewTaskFrameworkConfig1(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		// 测试合法配置
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, 30*time.Second, 10*time.Second, 60*time.Second)
		assert.NoError(t, err)
		assert.Equal(t, AddTaskModeExternal, config.AddTaskMode)
		assert.Equal(t, 30*time.Second, config.CycleInterval)
		assert.Equal(t, 10*time.Second, config.CheckInterval)
		assert.Equal(t, 60*time.Second, config.MaxTimeout)
	})

	t.Run("InvalidMode", func(t *testing.T) {
		// 测试非法模式
		config, err := NewTaskFrameworkConfig("invalid_mode", 0, 0, 0)
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "invalid add task mode")
	})

	t.Run("DefaultValues", func(t *testing.T) {
		// 测试默认值填充
		config, err := NewTaskFrameworkConfig(AddTaskModeRecursive, -1, -1, -1)
		assert.NoError(t, err)
		assert.Equal(t, DefaultCycleInterval, config.CycleInterval)
		assert.Equal(t, DefaultCheckInterval, config.CheckInterval)
		assert.Equal(t, DefaultMaxTimeout, config.MaxTimeout)
	})

	t.Run("CheckIntervalExceedCycle", func(t *testing.T) {
		// 测试检查间隔大于周期间隔的情况
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, 10*time.Second, 20*time.Second, 0)
		assert.NoError(t, err)
		assert.Equal(t, DefaultCheckInterval, config.CheckInterval) // 自动修正为默认值
	})
}

// ========== 基础功能测试 ==========
func TestTaskFramework_Basic(t *testing.T) {
	// 初始化框架（外部模式）
	td := time.Duration(0)
	config, err := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
	assert.NoError(t, err)
	config.CycleInterval = 1 * time.Second        // 缩短周期便于测试
	config.CheckInterval = 500 * time.Millisecond // 缩短检查间隔
	config.MaxTimeout = 2 * time.Second           // 缩短超时时间
	framework, err := NewTaskFramework(context.Background(), config)
	assert.NoError(t, err)
	defer framework.Stop()

	// 启动框架
	framework.Start()

	t.Run("AddForbiddenKeys", func(t *testing.T) {
		// 添加禁止key
		framework.AddForbiddenKeys("forbidden_key")

		// 尝试添加禁止key的任务
		task := generateTestTask("forbidden_key")
		successCount, failedTasks, err := framework.AddTasks([]*Task{task}, successExecuteFunc)
		assert.NoError(t, err)
		assert.Equal(t, 0, successCount)
		assert.Len(t, failedTasks, 1)
		assert.Equal(t, "task forbidden_key is in forbidden list", failedTasks[0].TaskResult.ResultReason)
		assert.Equal(t, TaskResultFail, failedTasks[0].TaskResult.Result)
	})

	t.Run("RemoveForbiddenKeys", func(t *testing.T) {
		// 添加后移除禁止key
		framework.AddForbiddenKeys("temp_forbidden")
		framework.RemoveForbiddenKeys("temp_forbidden")

		// 验证可以添加该key的任务
		task := generateTestTask("temp_forbidden")
		successCount, failedTasks, err := framework.AddTasks([]*Task{task}, successExecuteFunc)
		assert.NoError(t, err)
		assert.Equal(t, 1, successCount)
		assert.Empty(t, failedTasks)
	})

	t.Run("AddTasks_Duplicate", func(t *testing.T) {
		// 添加第一个任务
		task1 := generateTestTask("test_key_1")
		successCount, failedTasks, err := framework.AddTasks([]*Task{task1}, successExecuteFunc)
		assert.NoError(t, err)
		assert.Equal(t, 1, successCount)
		assert.Empty(t, failedTasks)

		// 重复添加相同key的任务
		task1Duplicate := generateTestTask("test_key_1")
		successCount2, failedTasks2, err := framework.AddTasks([]*Task{task1Duplicate}, successExecuteFunc)
		assert.NoError(t, err)
		assert.Equal(t, 0, successCount2)
		assert.Len(t, failedTasks2, 1)
		assert.Contains(t, failedTasks2[0].TaskResult.ResultReason, "already exists")
	})

	t.Run("AddTasks_EmptyKey", func(t *testing.T) {
		// 添加空key任务
		emptyKeyTask := &Task{Key: ""}
		successCount, failedTasks, err := framework.AddTasks([]*Task{emptyKeyTask}, successExecuteFunc)
		assert.Equal(t, 0, successCount)
		assert.NoError(t, err)
		assert.Len(t, failedTasks, 1)
		assert.Equal(t, "task key cannot be empty", failedTasks[0].TaskResult.ResultReason)
	})

	t.Run("AddTasks_NilExecuteFunc", func(t *testing.T) {
		// 执行函数为nil
		task := generateTestTask("test_key_nil_func")
		successCount, failedTasks, err := framework.AddTasks([]*Task{task}, nil)
		assert.Error(t, err) // 原测试错误：此处应该返回错误
		assert.Equal(t, 0, successCount)
		assert.Nil(t, failedTasks)
		assert.Equal(t, "executeFunc cannot be nil", err.Error())
	})

	t.Run("SetGenerateTasksFunc", func(t *testing.T) {
		// 测试设置生成函数
		framework.SetGenerateTasksFunc(testGenerateTasksFunc)
		framework.mu.RLock()
		assert.NotNil(t, framework.generateFunc)
		framework.mu.RUnlock()
	})
}

// ========== 临界场景测试 ==========
func TestTaskFramework_CriticalScenarios(t *testing.T) {
	td := time.Duration(0)

	// 测试超时场景
	t.Run("TaskTimeout", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
		assert.NoError(t, err)
		config.CycleInterval = 1 * time.Second
		config.CheckInterval = 500 * time.Millisecond
		config.MaxTimeout = 50 * time.Millisecond // 短超时
		framework, err := NewTaskFramework(context.Background(), config)
		assert.NoError(t, err)
		defer framework.Stop()

		framework.Start()

		// 添加超时任务
		task := generateTestTask("timeout_task")
		successCount, failedTasks, err := framework.AddTasks([]*Task{task}, timeoutExecuteFunc)
		assert.NoError(t, err)
		assert.Equal(t, 1, successCount)
		assert.Empty(t, failedTasks)

		// 等待任务超时完成（使用工具函数精准等待）
		success := waitForTaskResultCount(framework, "timeout_task", 1, 2*time.Second)
		assert.True(t, success, "任务超时计数未达到1")

		// 检查任务状态：超时失败
		framework.mu.RLock()
		defer framework.mu.RUnlock()
		taskObj, exists := framework.tasks["timeout_task"]
		assert.True(t, exists)
		assert.Equal(t, TaskStateCompleted, taskObj.TaskState)
		assert.Equal(t, TaskResultFail, taskObj.TaskResult.Result)
		assert.Contains(t, taskObj.TaskResult.ResultReason, "timeout")
		assert.Equal(t, uint64(1), taskObj.TaskResult.ResultCount) // 确保计数为1
	})

	// 测试周期触发场景
	t.Run("CycleTrigger", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
		assert.NoError(t, err)
		config.CycleInterval = 2 * time.Second // 延长周期，避免执行过快
		config.CheckInterval = 1 * time.Second
		config.MaxTimeout = 5 * time.Second
		framework, err := NewTaskFramework(context.Background(), config)
		assert.NoError(t, err)
		defer framework.Stop()

		framework.Start()

		// 添加待执行任务
		task := generateTestTask("cycle_task")
		successCount, failedTasks, err := framework.AddTasks([]*Task{task}, successExecuteFunc)
		assert.NoError(t, err)
		assert.Equal(t, 1, successCount)
		assert.Empty(t, failedTasks)

		// 初始状态：pending
		framework.mu.RLock()
		assert.Equal(t, TaskStatePending, framework.tasks["cycle_task"].TaskState)
		framework.mu.RUnlock()

		// 等待第一个周期触发并执行完成
		success := waitForTaskState(framework, "cycle_task", TaskStateCompleted, 3*time.Second)
		assert.True(t, success, "任务未完成第一个周期执行")

		// 检查第一个周期执行结果
		framework.mu.RLock()
		assert.Equal(t, TaskStateCompleted, framework.tasks["cycle_task"].TaskState)
		assert.Equal(t, TaskResultOK, framework.tasks["cycle_task"].TaskResult.Result)
		framework.mu.RUnlock()

		// 等待第二个周期触发（确保任务变为running）
		// 关键修复：等待第二个周期开始后立即检查，而不是等待完整周期
		time.Sleep(config.CycleInterval / 2) // 等待第二个周期开始
		success = waitForTaskState(framework, "cycle_task", TaskStateRunning, 1*time.Second)
		assert.True(t, success, "任务未进入第二个周期的running状态")

		// 最终状态检查
		framework.mu.RLock()
		assert.Equal(t, TaskStateRunning, framework.tasks["cycle_task"].TaskState)
		framework.mu.RUnlock()
	})

	// 测试递归生成任务场景
	t.Run("RecursiveGenerateTasks", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeRecursive, td, td, td)
		assert.NoError(t, err)
		config.CycleInterval = 2 * time.Second
		config.CheckInterval = 1 * time.Second
		config.MaxTimeout = 5 * time.Second
		framework, err := NewTaskFramework(context.Background(), config)
		assert.NoError(t, err)
		defer framework.Stop()

		// 设置生成任务函数
		framework.SetGenerateTasksFunc(testGenerateTasksFunc)
		framework.Start()

		// 添加初始任务
		initialTask := generateTestTask("initial_task")
		successCount, failedTasks, err := framework.AddTasks([]*Task{initialTask}, successExecuteFunc)
		assert.NoError(t, err)
		assert.Equal(t, 1, successCount)
		assert.Empty(t, failedTasks)

		// 等待初始任务执行完成（计数为1）
		success := waitForTaskResultCount(framework, "initial_task", 1, 3*time.Second)
		assert.True(t, success, "初始任务执行计数未达到1")

		// 检查初始任务状态
		framework.mu.RLock()
		initialTaskObj, exists := framework.tasks["initial_task"]
		assert.True(t, exists)
		assert.Equal(t, TaskStateCompleted, initialTaskObj.TaskState)
		assert.Equal(t, uint64(1), initialTaskObj.TaskResult.ResultCount) // 确保计数为1

		// 检查生成的新任务（允许延迟）
		newTaskExists := false
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			_, newTaskExists = framework.tasks["initial_task_generated"]
			if newTaskExists {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		assert.True(t, newTaskExists, "未生成新任务")
		framework.mu.RUnlock()
	})

	// 测试框架停止场景
	t.Run("FrameworkStop", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
		assert.NoError(t, err)
		config.CycleInterval = 1 * time.Second
		config.MaxTimeout = 2 * time.Second
		framework, err := NewTaskFramework(context.Background(), config)
		assert.NoError(t, err)

		framework.Start()

		// 添加任务
		task := generateTestTask("stop_task")
		_, _, err = framework.AddTasks([]*Task{task}, ctxCancelExecuteFunc)
		assert.NoError(t, err)

		// 等待任务开始执行后停止框架
		time.Sleep(100 * time.Millisecond)
		framework.Stop()

		// 检查任务状态：因框架停止失败
		framework.mu.RLock()
		defer framework.mu.RUnlock()
		taskObj, exists := framework.tasks["stop_task"]
		assert.True(t, exists)
		assert.Equal(t, TaskResultFail, taskObj.TaskResult.Result)
		assert.Contains(t, taskObj.TaskResult.ResultReason, "framework stopped")
	})

	// 测试禁止key动态更新
	t.Run("ForbiddenKeyDynamicUpdate", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
		assert.NoError(t, err)
		config.CycleInterval = 1 * time.Second
		framework, err := NewTaskFramework(context.Background(), config)
		assert.NoError(t, err)
		defer framework.Stop()
		framework.Start()

		// 添加任务
		taskKey := "dynamic_forbidden_task"
		task := generateTestTask(taskKey)
		_, _, err = framework.AddTasks([]*Task{task}, successExecuteFunc)
		assert.NoError(t, err)

		// 添加禁止key
		framework.AddForbiddenKeys(taskKey)

		// 等待周期执行
		time.Sleep(1500 * time.Millisecond)

		// 检查任务未被执行（仍为pending）
		framework.mu.RLock()
		assert.Equal(t, TaskStatePending, framework.tasks[taskKey].TaskState)
		framework.mu.RUnlock()

		// 移除禁止key
		framework.RemoveForbiddenKeys(taskKey)

		// 等待下一个周期
		time.Sleep(1000 * time.Millisecond)

		// 检查任务被执行
		framework.mu.RLock()
		assert.Equal(t, TaskStateCompleted, framework.tasks[taskKey].TaskState)
		framework.mu.RUnlock()
	})
}

// ========== 性能测试 ==========
func TestTaskFramework_Performance(t *testing.T) {
	// 跳过短测试模式下的性能测试
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	td := time.Duration(0)

	// 批量添加任务性能
	t.Run("BatchAddTasks", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
		assert.NoError(t, err)
		framework, err := NewTaskFramework(context.Background(), config)
		assert.NoError(t, err)
		defer framework.Stop()

		// 生成10000个测试任务（提升数量更能体现性能）
		taskCount := 10000
		tasks := make([]*Task, 0, taskCount)
		for i := 0; i < taskCount; i++ {
			tasks = append(tasks, generateTestTask(fmt.Sprintf("perf_add_task_%d", i)))
		}

		// 性能测试：记录耗时
		start := time.Now()
		successCount, failedTasks, err := framework.AddTasks(tasks, successExecuteFunc)
		elapsed := time.Since(start)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, taskCount, successCount)
		assert.Empty(t, failedTasks)

		// 计算性能指标
		ms := float64(elapsed.Milliseconds())
		tps := float64(taskCount) / (ms / 1000) // 每秒处理任务数
		t.Logf("BatchAddTasks: %d tasks added in %v (%.2f ms, %.2f TPS)",
			taskCount, elapsed, ms, tps)

		// 性能要求：10000个任务添加耗时 < 2秒
		assert.Less(t, elapsed, 2*time.Second)
	})

	// 批量执行任务性能
	t.Run("BatchExecuteTasks", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
		assert.NoError(t, err)
		config.CycleInterval = 5 * time.Second // 延长周期避免重复执行
		config.CheckInterval = 2 * time.Second
		config.MaxTimeout = 5 * time.Second
		framework, err := NewTaskFramework(context.Background(), config)
		assert.NoError(t, err)
		defer framework.Stop()

		framework.Start()

		// 计数器：记录成功执行的任务数（确保每个任务只计数一次）
		executedTasks := make(map[string]bool)
		var mu sync.Mutex
		var successCount atomic.Uint64

		executeFunc := func(ctx context.Context, task *Task) TaskExecutionResult {
			mu.Lock()
			defer mu.Unlock()
			// 确保每个任务只计数一次，避免周期重复执行导致计数翻倍
			if !executedTasks[task.Key] {
				executedTasks[task.Key] = true
				successCount.Add(1)
			}
			return TaskExecutionResult{
				Result: TaskResultOK,
				Err:    "",
			}
		}

		// 添加5000个任务
		taskCount := 5000
		tasks := make([]*Task, 0, taskCount)
		for i := 0; i < taskCount; i++ {
			tasks = append(tasks, generateTestTask(fmt.Sprintf("exec_task_%d", i)))
		}
		_, _, err = framework.AddTasks(tasks, executeFunc)
		assert.NoError(t, err)

		// 等待任务执行完成（使用工具函数精准等待）
		start := time.Now()
		success := waitForTaskResultCountTotal(framework, taskCount, 3*time.Second)
		elapsed := time.Since(start)

		// 强制等待剩余任务执行完成
		if !success {
			time.Sleep(1 * time.Second)
		}

		// 验证执行结果（使用map确保每个任务只执行一次）
		mu.Lock()
		actualCount := len(executedTasks)
		mu.Unlock()
		assert.Equal(t, taskCount, actualCount)

		// 计算性能指标
		ms := float64(elapsed.Milliseconds())
		tps := float64(taskCount) / (ms / 1000)
		t.Logf("BatchExecuteTasks: %d tasks executed in %v (%.2f ms, %.2f TPS)",
			taskCount, elapsed, ms, tps)

		// 性能要求：5000个任务执行耗时 < 3秒
		assert.Less(t, elapsed, 3*time.Second)
	})

	// 并发添加任务性能
	t.Run("ConcurrentAddTasks", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
		assert.NoError(t, err)
		framework, err := NewTaskFramework(context.Background(), config)
		assert.NoError(t, err)
		defer framework.Stop()

		// 并发数
		concurrency := 10
		tasksPerGoroutine := 1000
		totalTasks := concurrency * tasksPerGoroutine
		var totalSuccess atomic.Int64
		var wg sync.WaitGroup

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				tasks := make([]*Task, 0, tasksPerGoroutine)
				for j := 0; j < tasksPerGoroutine; j++ {
					key := fmt.Sprintf("concurrent_task_%d_%d", goroutineID, j)
					tasks = append(tasks, generateTestTask(key))
				}
				success, _, err := framework.AddTasks(tasks, successExecuteFunc)
				assert.NoError(t, err)
				totalSuccess.Add(int64(success))
			}(i)
		}
		wg.Wait()
		elapsed := time.Since(start)

		// 验证结果
		assert.Equal(t, int64(totalTasks), totalSuccess.Load())

		// 性能日志
		ms := float64(elapsed.Milliseconds())
		tps := float64(totalTasks) / (ms / 1000)
		t.Logf("ConcurrentAddTasks: %d tasks added by %d goroutines in %v (%.2f TPS)",
			totalTasks, concurrency, elapsed, tps)

		// 性能要求：10000个任务并发添加 < 3秒
		assert.Less(t, elapsed, 3*time.Second)
	})
}
