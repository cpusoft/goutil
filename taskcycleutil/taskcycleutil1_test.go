package taskcycleutil

import (
	"context"
	"fmt"
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

// 超时执行函数：阻塞超过指定时间
func timeoutExecuteFunc(ctx context.Context, task *Task) TaskExecutionResult {
	select {
	case <-ctx.Done():
		return TaskExecutionResult{
			Result: TaskResultFail,
			Err:    ctx.Err().Error(),
		}
	case <-time.After(100 * time.Millisecond): // 模拟超时任务
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

// ========== 基础功能测试 ==========
func TestTaskFramework_Basic(t *testing.T) {
	// 初始化框架（外部模式）
	td := time.Duration(0)
	config, _ := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
	config.CycleInterval = 1 * time.Second        // 缩短周期便于测试
	config.CheckInterval = 500 * time.Millisecond // 缩短检查间隔
	config.MaxTimeout = 2 * time.Second           // 缩短超时时间
	framework, err := NewTaskFramework(context.Background(), config)
	assert.NoError(t, err)
	defer framework.Stop()

	// 启动框架
	framework.Start()
	defer framework.Stop()

	t.Run("AddForbiddenKeys", func(t *testing.T) {
		// 添加禁止key
		framework.AddForbiddenKeys("forbidden_key")

		// 尝试添加禁止key的任务
		task := generateTestTask("forbidden_key")
		successCount, failedTasks, err := framework.AddTasks([]*Task{task}, successExecuteFunc)
		assert.NoError(t, err)
		assert.Equal(t, 0, successCount)
		assert.Len(t, failedTasks, 1)
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
	})

	t.Run("AddTasks_EmptyKey", func(t *testing.T) {
		// 添加空key任务
		emptyKeyTask := &Task{Key: ""}
		successCount, failedTasks, err := framework.AddTasks([]*Task{emptyKeyTask}, successExecuteFunc)
		assert.Equal(t, 0, successCount)
		assert.NoError(t, err)
		assert.Len(t, failedTasks, 1)
	})

	t.Run("AddTasks_NilExecuteFunc", func(t *testing.T) {
		// 执行函数为nil
		task := generateTestTask("test_key_nil_func")
		successCount, failedTasks, err := framework.AddTasks([]*Task{task}, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, successCount)
		assert.Len(t, failedTasks, 1)
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

		// 等待周期触发执行
		time.Sleep(2 * time.Second)

		// 检查任务状态：超时失败
		framework.mu.RLock()
		defer framework.mu.RUnlock()
		assert.Equal(t, TaskStateCompleted, framework.tasks["timeout_task"].TaskState)
		assert.Equal(t, TaskResultFail, framework.tasks["timeout_task"].TaskResult.Result)
		assert.Contains(t, framework.tasks["timeout_task"].TaskResult.ResultReason, "timeout")
		assert.Equal(t, uint64(1), framework.tasks["timeout_task"].TaskResult.ResultCount)
	})

	// 测试周期触发场景
	t.Run("CycleTrigger", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
		assert.NoError(t, err)
		config.CycleInterval = 1 * time.Second // 1秒周期
		config.CheckInterval = 500 * time.Millisecond
		config.MaxTimeout = 2 * time.Second
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

		// 等待第一个周期触发（1秒）
		time.Sleep(1500 * time.Millisecond)

		// 检查状态：执行中/已完成
		framework.mu.RLock()
		taskState := framework.tasks["cycle_task"].TaskState
		framework.mu.RUnlock()
		assert.True(t, taskState == TaskStateRunning || taskState == TaskStateCompleted)

		// 等待第二个周期触发
		time.Sleep(1000 * time.Millisecond)

		// 检查状态：再次变为running（completed任务被重新执行）
		framework.mu.RLock()
		assert.Equal(t, TaskStateRunning, framework.tasks["cycle_task"].TaskState)
		framework.mu.RUnlock()
	})

	// 测试递归生成任务场景
	t.Run("RecursiveGenerateTasks", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeRecursive, td, td, td)
		assert.NoError(t, err)
		config.CycleInterval = 1 * time.Second
		config.CheckInterval = 500 * time.Millisecond
		config.MaxTimeout = 2 * time.Second
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

		// 等待任务执行完成并生成新任务
		time.Sleep(1 * time.Second)

		// 检查初始任务状态：改回pending
		framework.mu.RLock()
		assert.Equal(t, TaskStatePending, framework.tasks["initial_task"].TaskState)
		assert.Equal(t, uint64(1), framework.tasks["initial_task"].TaskResult.ResultCount)
		// 检查生成的新任务
		_, newTaskExists := framework.tasks["initial_task_generated"]
		assert.True(t, newTaskExists)
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
		framework.AddTasks([]*Task{task}, successExecuteFunc)

		// 立即停止框架
		framework.Stop()

		// 检查任务状态：因框架停止失败
		framework.mu.RLock()
		defer framework.mu.RUnlock()
		assert.Equal(t, TaskResultFail, framework.tasks["stop_task"].TaskResult.Result)
		assert.Equal(t, "framework stopped", framework.tasks["stop_task"].TaskResult.ResultReason)
	})
}

// ========== 性能测试 ==========
func TestTaskFramework_Performance(t *testing.T) {
	td := time.Duration(0)
	// 批量添加任务性能
	t.Run("BatchAddTasks", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
		assert.NoError(t, err)
		framework, err := NewTaskFramework(context.Background(), config)
		assert.NoError(t, err)
		defer framework.Stop()

		// 生成1000个测试任务
		taskCount := 1000
		tasks := make([]*Task, 0, taskCount)
		for i := 0; i < taskCount; i++ {
			tasks = append(tasks, generateTestTask(fmt.Sprintf("perf_task_%d", i)))
		}

		// 性能测试：记录耗时
		start := time.Now()
		successCount, failedTasks, err := framework.AddTasks(tasks, successExecuteFunc)
		elapsed := time.Since(start)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, taskCount, successCount)
		assert.Empty(t, failedTasks)

		// 修复类型不匹配：将 int64 转换为 float64
		ms := float64(elapsed.Milliseconds())
		t.Logf("BatchAddTasks: %d tasks added in %v (%.2f tasks/ms)",
			taskCount, elapsed, float64(taskCount)/ms)

		// 要求：1000个任务添加耗时 < 1秒（可根据实际调整）
		assert.Less(t, elapsed, 1*time.Second)
	})

	// 批量执行任务性能
	t.Run("BatchExecuteTasks", func(t *testing.T) {
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, td, td, td)
		assert.NoError(t, err)
		config.CycleInterval = 1 * time.Second
		config.MaxTimeout = 5 * time.Second
		framework, err := NewTaskFramework(context.Background(), config)
		assert.NoError(t, err)
		defer framework.Stop()

		framework.Start()

		// 计数器：记录成功执行的任务数
		var successCount atomic.Uint64
		executeFunc := func(ctx context.Context, task *Task) TaskExecutionResult {
			successCount.Add(1)
			return TaskExecutionResult{
				Result: TaskResultOK,
				Err:    "",
			}
		}

		// 添加500个任务
		taskCount := 500
		tasks := make([]*Task, 0, taskCount)
		for i := 0; i < taskCount; i++ {
			tasks = append(tasks, generateTestTask(fmt.Sprintf("exec_task_%d", i)))
		}
		framework.AddTasks(tasks, executeFunc)

		// 等待周期触发并执行完成
		start := time.Now()
		time.Sleep(2 * time.Second) // 等待2个周期
		elapsed := time.Since(start)

		// 验证执行结果
		assert.Equal(t, uint64(taskCount), successCount.Load())

		// 修复类型不匹配：将 int64 转换为 float64
		ms := float64(elapsed.Milliseconds())
		t.Logf("BatchExecuteTasks: %d tasks executed in %v (%.2f tasks/ms)",
			taskCount, elapsed, float64(taskCount)/ms)

		// 要求：500个任务执行耗时 < 2秒（可根据实际调整）
		assert.Less(t, elapsed, 2*time.Second)
	})
}
