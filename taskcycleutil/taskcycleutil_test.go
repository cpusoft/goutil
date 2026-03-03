package taskcycleutil

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// ===================== 辅助函数：模拟执行函数 =====================
// 模拟执行成功的函数
func mockSuccessExecuteFunc(ctx context.Context, task *Task) TaskExecutionResult {
	return TaskExecutionResult{
		Result: TaskResultOK,
		Err:    "",
	}
}

// 模拟执行失败的函数
func mockFailExecuteFunc(ctx context.Context, task *Task) TaskExecutionResult {
	return TaskExecutionResult{
		Result: TaskResultFail,
		Err:    "mock execute fail",
	}
}

// 模拟超时的执行函数（阻塞直到ctx超时）
func mockTimeoutExecuteFunc(ctx context.Context, task *Task) TaskExecutionResult {
	<-ctx.Done()
	return TaskExecutionResult{
		Result: TaskResultFail,
		Err:    ctx.Err().Error(),
	}
}

// 模拟生成新任务的函数（从成功任务生成1个新任务）
func mockGenerateTasksFunc(completedTask *Task) []*Task {
	newTask := NewTask(
		completedTask.Key+"_new",
		completedTask.TaskParam,
		completedTask.executeFunc,
	)
	return []*Task{newTask}
}

// ===================== 配置初始化测试 =====================
func TestNewTaskFrameworkConfig(t *testing.T) {
	tests := []struct {
		name          string
		addTaskMode   string
		cycleInterval time.Duration
		checkInterval time.Duration
		maxTimeout    time.Duration
		wantErr       bool
		wantConfig    *TaskFrameworkConfig // 预期的核心配置值
	}{
		{
			name:          "合法递归模式+默认值",
			addTaskMode:   AddTaskModeRecursive,
			cycleInterval: -1, // 触发默认值
			checkInterval: -1, // 触发默认值
			maxTimeout:    -1, // 触发默认值
			wantErr:       false,
			wantConfig: &TaskFrameworkConfig{
				AddTaskMode:   AddTaskModeRecursive,
				CycleInterval: DefaultCycleInterval,
				CheckInterval: DefaultCheckInterval,
				MaxTimeout:    DefaultMaxTimeout,
			},
		},
		{
			name:          "合法外部模式+自定义值",
			addTaskMode:   AddTaskModeExternal,
			cycleInterval: 10 * time.Second,
			checkInterval: 3 * time.Second,
			maxTimeout:    20 * time.Second,
			wantErr:       false,
			wantConfig: &TaskFrameworkConfig{
				AddTaskMode:   AddTaskModeExternal,
				CycleInterval: 10 * time.Second,
				CheckInterval: 3 * time.Second,
				MaxTimeout:    20 * time.Second,
			},
		},
		{
			name:          "非法模式",
			addTaskMode:   "invalid",
			cycleInterval: 10 * time.Second,
			checkInterval: 3 * time.Second,
			maxTimeout:    20 * time.Second,
			wantErr:       true,
			wantConfig:    nil,
		},
		{
			name:          "checkInterval>=cycleInterval（触发默认）",
			addTaskMode:   AddTaskModeRecursive,
			cycleInterval: 5 * time.Second,
			checkInterval: 6 * time.Second, // 大于周期，触发默认
			maxTimeout:    20 * time.Second,
			wantErr:       false,
			wantConfig: &TaskFrameworkConfig{
				AddTaskMode:   AddTaskModeRecursive,
				CycleInterval: 5 * time.Second,
				CheckInterval: DefaultCheckInterval, // 触发默认
				MaxTimeout:    20 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTaskFrameworkConfig(tt.addTaskMode, tt.cycleInterval, tt.checkInterval, tt.maxTimeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTaskFrameworkConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.AddTaskMode != tt.wantConfig.AddTaskMode {
					t.Errorf("AddTaskMode = %v, want %v", got.AddTaskMode, tt.wantConfig.AddTaskMode)
				}
				if got.CycleInterval != tt.wantConfig.CycleInterval {
					t.Errorf("CycleInterval = %v, want %v", got.CycleInterval, tt.wantConfig.CycleInterval)
				}
				if got.CheckInterval != tt.wantConfig.CheckInterval {
					t.Errorf("CheckInterval = %v, want %v", got.CheckInterval, tt.wantConfig.CheckInterval)
				}
				if got.MaxTimeout != tt.wantConfig.MaxTimeout {
					t.Errorf("MaxTimeout = %v, want %v", got.MaxTimeout, tt.wantConfig.MaxTimeout)
				}
			}
		})
	}
}

// ===================== 框架初始化测试 =====================
func TestNewTaskFramework(t *testing.T) {
	// 测试1：空配置入参
	_, err := NewTaskFramework(context.Background(), nil)
	if err == nil || err.Error() != "taskFrameworkConfig cannot be nil" {
		t.Errorf("NewTaskFramework() error = %v, want 'taskFrameworkConfig cannot be nil'", err)
	}

	// 测试2：合法配置入参
	validConfig, _ := NewTaskFrameworkConfig(AddTaskModeExternal, 10*time.Second, 3*time.Second, 20*time.Second)
	fw, err := NewTaskFramework(context.Background(), validConfig)
	if err != nil {
		t.Fatalf("NewTaskFramework() unexpected error: %v", err)
	}
	if fw.tasks == nil || fw.forbiddenKeys == nil {
		t.Error("NewTaskFramework() tasks/forbiddenKeys not initialized")
	}
	if fw.taskFrameworkConfig.AddTaskMode != AddTaskModeExternal {
		t.Errorf("AddTaskMode = %v, want %v", fw.taskFrameworkConfig.AddTaskMode, AddTaskModeExternal)
	}
}

// ===================== 禁止Key测试 =====================
func TestForbiddenKeys(t *testing.T) {
	// 初始化框架
	config, _ := NewTaskFrameworkConfig(AddTaskModeExternal, 10*time.Second, 3*time.Second, 20*time.Second)
	fw, _ := NewTaskFramework(context.Background(), config)

	// 测试1：添加禁止Key
	fw.AddForbiddenKeys("key1", "key2", "") // 空key忽略
	t.Log("after add forbiddenkeys")
	fw.mu.RLock()
	t.Log("after RLock")
	defer fw.mu.RUnlock()
	t.Log("after defer RUnlock")
	if _, exists := fw.forbiddenKeys["key1"]; !exists {
		t.Error("AddForbiddenKeys() key1 not added")
	}
	t.Log("after key1")

	if _, exists := fw.forbiddenKeys["key2"]; !exists {
		t.Error("AddForbiddenKeys() key2 not added")
	}
	t.Log("after key2")

	if _, exists := fw.forbiddenKeys[""]; exists {
		t.Error("AddForbiddenKeys() empty key should be ignored")
	}
	t.Log("after key empty")

	// 测试2：移除禁止Key
	fw.RemoveForbiddenKeys("key1", "key3") // key3不存在
	t.Log("after rm forbiddenkeys")
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	if _, exists := fw.forbiddenKeys["key1"]; exists {
		t.Error("RemoveForbiddenKeys() key1 not removed")
	}
	t.Log("after key1 again")
	if _, exists := fw.forbiddenKeys["key2"]; !exists {
		t.Error("RemoveForbiddenKeys() key2 should remain")
	}
	t.Log("after key2 again")
}

// ===================== 任务添加测试（递归模式） =====================
func TestAddTasks_RecursiveMode(t *testing.T) {
	// 初始化框架（递归模式，缩短时间便于测试）
	config, _ := NewTaskFrameworkConfig(AddTaskModeRecursive, 5*time.Second, 2*time.Second, 10*time.Second)
	fw, _ := NewTaskFramework(context.Background(), config)
	defer fw.Stop()

	// 准备测试任务
	param := TaskParam{Content: "test", Params: map[string]interface{}{"id": 1}}
	task1 := NewTask("key1", param, mockSuccessExecuteFunc)
	task2 := NewTask("", param, mockSuccessExecuteFunc)     // 空key
	task3 := NewTask("key3", param, mockSuccessExecuteFunc) // 禁止key
	task4 := NewTask("key1", param, mockSuccessExecuteFunc) // 重复key
	tasks := []*Task{task1, task2, task3, task4}

	// 添加禁止key
	fw.AddForbiddenKeys("key3")

	// 执行添加
	successCount, failedTasks, err := fw.AddTasks(tasks, mockSuccessExecuteFunc)
	if err != nil {
		t.Fatalf("AddTasks() unexpected error: %v", err)
	}

	// 验证结果
	if successCount != 1 { // 仅task1成功
		t.Errorf("successCount = %v, want 1", successCount)
	}
	if len(failedTasks) != 3 { // task2/task3/task4失败
		t.Errorf("failedTasks count = %v, want 3", len(failedTasks))
	}

	// 验证失败原因
	failReasons := make(map[string]string)
	for _, ft := range failedTasks {
		failReasons[ft.Key] = ft.TaskResult.ResultReason
	}
	if failReasons[""] != "task key cannot be empty" {
		t.Errorf("empty key fail reason = %v, want 'task key cannot be empty'", failReasons[""])
	}
	if failReasons["key3"] != "task key3 is in forbidden list" {
		t.Errorf("forbidden key fail reason = %v, want 'task key3 is in forbidden list'", failReasons["key3"])
	}
	if failReasons["key1"] != "task key1 already exists (any state)" {
		t.Errorf("duplicate key fail reason = %v, want 'task key1 already exists (any state)'", failReasons["key1"])
	}

	// 验证task1状态（递归模式应立即设为running）
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	if task1.TaskState != TaskStateRunning {
		t.Errorf("task1 state = %v, want %v", task1.TaskState, TaskStateRunning)
	}
}

// ===================== 任务添加测试（外部模式） =====================
func TestAddTasks_ExternalMode(t *testing.T) {
	// 初始化框架（外部模式，缩短时间便于测试）
	config, _ := NewTaskFrameworkConfig(AddTaskModeExternal, 5*time.Second, 2*time.Second, 10*time.Second)
	fw, _ := NewTaskFramework(context.Background(), config)
	defer fw.Stop()

	// 准备测试任务
	param := TaskParam{Content: "test", Params: map[string]interface{}{"id": 1}}
	task1 := NewTask("key1", param, mockSuccessExecuteFunc)

	// 执行添加
	successCount, failedTasks, err := fw.AddTasks([]*Task{task1}, mockSuccessExecuteFunc)
	if err != nil {
		t.Fatalf("AddTasks() unexpected error: %v", err)
	}

	// 验证结果
	if successCount != 1 || len(failedTasks) != 0 {
		t.Errorf("successCount = %v, failedTasks = %v, want (1, 0)", successCount, len(failedTasks))
	}

	// 验证task1状态（外部模式应设为pending）
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	if task1.TaskState != TaskStatePending {
		t.Errorf("task1 state = %v, want %v", task1.TaskState, TaskStatePending)
	}
}

// ===================== 任务执行测试：成功 =====================
func TestExecuteTask_Success(t *testing.T) {
	// 初始化框架（递归模式，缩短时间）
	config, _ := NewTaskFrameworkConfig(AddTaskModeRecursive, 5*time.Second, 2*time.Second, 10*time.Second)
	fw, _ := NewTaskFramework(context.Background(), config)
	fw.SetGenerateTasksFunc(mockGenerateTasksFunc) // 设置生成新任务函数
	defer fw.Stop()

	// 添加测试任务
	param := TaskParam{Content: "test", Params: map[string]interface{}{"id": 1}}
	task := NewTask("success_task", param, mockSuccessExecuteFunc)
	_, _, err := fw.AddTasks([]*Task{task}, mockSuccessExecuteFunc)
	if err != nil {
		t.Fatalf("AddTasks() error: %v", err)
	}

	// 等待任务执行完成（异步执行，等待1秒）
	time.Sleep(1 * time.Second)

	// 验证任务结果
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	if task.TaskState != TaskStateCompleted {
		t.Errorf("task state = %v, want %v", task.TaskState, TaskStateCompleted)
	}
	if task.TaskResult.Result != TaskResultOK {
		t.Errorf("task result = %v, want %v", task.TaskResult.Result, TaskResultOK)
	}
	if task.TaskResult.ResultCount != 1 {
		t.Errorf("result count = %v, want 1", task.TaskResult.ResultCount)
	}

	// 验证生成的新任务（success_task_new）
	newTask, exists := fw.tasks["success_task_new"]
	if !exists {
		t.Error("generated new task not found")
	} else {
		if newTask.TaskState != TaskStateRunning {
			t.Errorf("new task state = %v, want %v", newTask.TaskState, TaskStateRunning)
		}
	}
}

// ===================== 任务执行测试：失败 =====================
func TestExecuteTask_Fail(t *testing.T) {
	// 初始化框架
	config, _ := NewTaskFrameworkConfig(AddTaskModeRecursive, 5*time.Second, 2*time.Second, 10*time.Second)
	fw, _ := NewTaskFramework(context.Background(), config)
	defer fw.Stop()

	// 添加测试任务
	param := TaskParam{Content: "test", Params: map[string]interface{}{"id": 1}}
	task := NewTask("fail_task", param, mockFailExecuteFunc)
	_, _, err := fw.AddTasks([]*Task{task}, mockFailExecuteFunc)
	if err != nil {
		t.Fatalf("AddTasks() error: %v", err)
	}

	// 等待任务执行完成
	time.Sleep(1 * time.Second)

	// 验证结果
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	if task.TaskState != TaskStateCompleted {
		t.Errorf("task state = %v, want %v", task.TaskState, TaskStateCompleted)
	}
	if task.TaskResult.Result != TaskResultFail {
		t.Errorf("task result = %v, want %v", task.TaskResult.Result, TaskResultFail)
	}
	if task.TaskResult.ResultReason != "mock execute fail" {
		t.Errorf("fail reason = %v, want 'mock execute fail'", task.TaskResult.ResultReason)
	}
	if task.TaskResult.ResultCount != 1 {
		t.Errorf("result count = %v, want 1", task.TaskResult.ResultCount)
	}
}

// ===================== 任务执行测试：超时 =====================
func TestExecuteTask_Timeout(t *testing.T) {
	// 初始化框架（超时设为1秒，便于测试）
	config, _ := NewTaskFrameworkConfig(AddTaskModeRecursive, 5*time.Second, 2*time.Second, 1*time.Second)
	fw, _ := NewTaskFramework(context.Background(), config)
	defer fw.Stop()

	// 添加测试任务（执行函数会阻塞直到超时）
	param := TaskParam{Content: "test", Params: map[string]interface{}{"id": 1}}
	task := NewTask("timeout_task", param, mockTimeoutExecuteFunc)
	_, _, err := fw.AddTasks([]*Task{task}, mockTimeoutExecuteFunc)
	if err != nil {
		t.Fatalf("AddTasks() error: %v", err)
	}

	// 等待超时触发（2秒）
	time.Sleep(2 * time.Second)

	// 验证结果
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	if task.TaskState != TaskStateCompleted {
		t.Errorf("task state = %v, want %v", task.TaskState, TaskStateCompleted)
	}
	if task.TaskResult.Result != TaskResultFail {
		t.Errorf("task result = %v, want %v", task.TaskResult.Result, TaskResultFail)
	}
	if task.TaskResult.ResultReason == "" || task.TaskResult.ResultCount != 1 {
		t.Errorf("timeout task result invalid: reason=%v, count=%v", task.TaskResult.ResultReason, task.TaskResult.ResultCount)
	}
}

// ===================== 任务执行测试：框架停止 =====================
func TestExecuteTask_FrameworkStop(t *testing.T) {
	// 初始化框架
	config, _ := NewTaskFrameworkConfig(AddTaskModeRecursive, 5*time.Second, 2*time.Second, 10*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	fw, _ := NewTaskFramework(ctx, config)

	// 添加测试任务（执行函数阻塞）
	param := TaskParam{Content: "test", Params: map[string]interface{}{"id": 1}}
	blockFunc := func(ctx context.Context, task *Task) TaskExecutionResult {
		<-ctx.Done()
		return TaskExecutionResult{}
	}
	task := NewTask("stop_task", param, blockFunc)
	_, _, err := fw.AddTasks([]*Task{task}, blockFunc)
	if err != nil {
		t.Fatalf("AddTasks() error: %v", err)
	}

	// 立即停止框架
	cancel()
	fw.Stop()

	// 验证结果
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	if task.TaskState != TaskStateCompleted {
		t.Errorf("task state = %v, want %v", task.TaskState, TaskStateCompleted)
	}
	if task.TaskResult.Result != TaskResultFail {
		t.Errorf("task result = %v, want %v", task.TaskResult.Result, TaskResultFail)
	}
	if task.TaskResult.ResultReason != "framework stopped" {
		t.Errorf("fail reason = %v, want 'framework stopped'", task.TaskResult.ResultReason)
	}
}

// ===================== 周期调度测试 =====================
func TestCycleExecutor(t *testing.T) {
	// 初始化框架（外部模式，周期2秒，检查间隔1秒，超时5秒）
	config, _ := NewTaskFrameworkConfig(AddTaskModeExternal, 2*time.Second, 1*time.Second, 5*time.Second)
	fw, _ := NewTaskFramework(context.Background(), config)
	fw.Start() // 启动周期调度
	defer fw.Stop()

	// 添加pending状态任务
	param := TaskParam{Content: "cycle_task", Params: map[string]interface{}{"id": 1}}
	task := NewTask("cycle_task", param, mockSuccessExecuteFunc)
	_, _, err := fw.AddTasks([]*Task{task}, mockSuccessExecuteFunc)
	if err != nil {
		t.Fatalf("AddTasks() error: %v", err)
	}

	// 等待第一个周期执行（3秒）
	time.Sleep(3 * time.Second)

	// 验证任务状态（应从pending转为running→completed）
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	if task.TaskState != TaskStateCompleted {
		t.Errorf("cycle task state = %v, want %v", task.TaskState, TaskStateCompleted)
	}
	if fw.cycleCount < 1 {
		t.Errorf("cycle count = %v, want >=1", fw.cycleCount)
	}
}

// ===================== 临界值测试 =====================
func TestCriticalValues(t *testing.T) {
	// 测试1：超时临界值（刚好等于MaxTimeout）
	t.Run("timeout_critical", func(t *testing.T) {
		timeout := 2 * time.Second
		config, _ := NewTaskFrameworkConfig(AddTaskModeRecursive, 5*time.Second, 2*time.Second, timeout)
		fw, _ := NewTaskFramework(context.Background(), config)
		defer fw.Stop()

		// 任务执行函数阻塞timeout+100ms
		blockFunc := func(ctx context.Context, task *Task) TaskExecutionResult {
			time.Sleep(timeout + 100*time.Millisecond)
			return TaskExecutionResult{}
		}
		task := NewTask("critical_timeout", TaskParam{}, blockFunc)
		_, _, err := fw.AddTasks([]*Task{task}, blockFunc)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(timeout + 200*time.Millisecond)
		fw.mu.RLock()
		defer fw.mu.RUnlock()
		if task.TaskResult.Result != TaskResultFail || task.TaskResult.ResultReason == "" {
			t.Error("critical timeout task not marked as fail")
		}
	})

	// 测试2：周期检查间隔临界（等于周期）
	t.Run("check_interval_critical", func(t *testing.T) {
		cycle := 5 * time.Second
		check := cycle // 等于周期，触发默认值
		config, err := NewTaskFrameworkConfig(AddTaskModeExternal, cycle, check, 10*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if config.CheckInterval != DefaultCheckInterval {
			t.Errorf("check interval = %v, want %v", config.CheckInterval, DefaultCheckInterval)
		}
	})

	// 测试3：任务完成时间在周期前10分钟/后20分钟（测试时缩短为前1秒/后1秒）
	t.Run("complete_time_critical", func(t *testing.T) {
		cycle := 3 * time.Second
		check := 1 * time.Second // 模拟前10分钟
		config, _ := NewTaskFrameworkConfig(AddTaskModeRecursive, cycle, check, 10*time.Second)
		fw, _ := NewTaskFramework(context.Background(), config)
		fw.SetGenerateTasksFunc(mockGenerateTasksFunc)
		fw.Start()
		defer fw.Stop()

		// 添加任务，使其在周期前1秒完成
		task := NewTask("time_critical", TaskParam{}, mockSuccessExecuteFunc)
		_, _, err := fw.AddTasks([]*Task{task}, mockSuccessExecuteFunc)
		if err != nil {
			t.Fatal(err)
		}

		// 等待检查间隔触发（2秒）
		time.Sleep(2 * time.Second)
		fw.mu.RLock()
		defer fw.mu.RUnlock()
		_, exists := fw.tasks["time_critical_new"]
		if !exists {
			t.Error("new task not generated in check phase (pre 10min)")
		}
	})
}

// ===================== 性能测试 =====================
// 批量添加任务性能测试
func BenchmarkAddTasks(b *testing.B) {
	// 初始化框架
	config, _ := NewTaskFrameworkConfig(AddTaskModeExternal, 10*time.Second, 3*time.Second, 20*time.Second)
	fw, _ := NewTaskFramework(context.Background(), config)
	defer fw.Stop()

	// 准备批量任务
	param := TaskParam{Content: "benchmark", Params: map[string]interface{}{"id": 0}}
	tasks := make([]*Task, 100) // 每次添加100个任务
	for i := 0; i < 100; i++ {
		tasks[i] = NewTask(fmt.Sprintf("bench_key_%d", i), param, mockSuccessExecuteFunc)
	}

	// 重置计时器，执行测试
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每次测试前清空任务（避免重复）
		fw.mu.Lock()
		fw.tasks = make(map[string]*Task)
		fw.mu.Unlock()

		_, _, err := fw.AddTasks(tasks, mockSuccessExecuteFunc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 批量执行任务性能测试
func BenchmarkExecuteTasks(b *testing.B) {
	// 初始化框架（递归模式，缩短超时）
	config, _ := NewTaskFrameworkConfig(AddTaskModeRecursive, 10*time.Second, 3*time.Second, 5*time.Second)
	fw, _ := NewTaskFramework(context.Background(), config)
	defer fw.Stop()

	// 准备批量任务
	param := TaskParam{Content: "benchmark", Params: map[string]interface{}{"id": 0}}
	tasks := make([]*Task, 50) // 每次执行50个任务
	for i := 0; i < 50; i++ {
		tasks[i] = NewTask(fmt.Sprintf("bench_exec_key_%d", i), param, mockSuccessExecuteFunc)
	}

	// 重置计时器
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 清空任务
		fw.mu.Lock()
		fw.tasks = make(map[string]*Task)
		fw.mu.Unlock()

		// 添加并执行任务
		_, _, err := fw.AddTasks(tasks, mockSuccessExecuteFunc)
		if err != nil {
			b.Fatal(err)
		}

		// 等待执行完成
		time.Sleep(1 * time.Second)
	}
}
