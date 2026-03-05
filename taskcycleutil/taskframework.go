package taskcycleutil

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

// ========== 核心枚举定义 ==========

const (
	AddTaskModeRecursive = "recursive" // 递归模式：从成功任务生成新任务并立即执行, 用于从tal递归读取
	AddTaskModeExternal  = "external"  // 外部模式：外部注入任务，待下周期执行，用于precept动态感知一次性导入

	
	DefaultCycleInterval = 1800 * time.Second
	DefaultCheckInterval = 600 * time.Second
	DefaultMaxTimeout    = 4199 * time.Second // 2*1800+600-1（避免正好70分钟超时）
)

// ========== 配置结构体定义 ==========
type TaskFrameworkConfig struct {
	AddTaskMode   string        // 任务添加模式（必填）
	CycleInterval time.Duration // 周期间隔（默认30分钟）
	CheckInterval time.Duration // 周期内检查间隔（默认10分钟）
	MaxTimeout    time.Duration // 最长执行超时（默认70分钟=2*30+10）
}

// addTaskMode: recursive or external
// when CycleInterval or CheckInterval or MaxTimeout is zero or negative, use default value
func NewTaskFrameworkConfig(addTaskMode string,
	cycleInterval time.Duration, checkInterval time.Duration,
	maxTimeout time.Duration) (*TaskFrameworkConfig, error) {

	// 校验配置
	if addTaskMode != AddTaskModeRecursive && addTaskMode != AddTaskModeExternal {
		return nil, errors.New("invalid add task mode (must be 1 or 2)")
	}
	// 配置默认值
	if cycleInterval <= 0 {
		cycleInterval = DefaultCycleInterval
	}
	if checkInterval <= 0 || checkInterval >= cycleInterval {
		checkInterval = DefaultCheckInterval
	}
	if maxTimeout <= 0 {
		maxTimeout = DefaultMaxTimeout
	}

	return &TaskFrameworkConfig{
		AddTaskMode:   addTaskMode,
		CycleInterval: cycleInterval,
		CheckInterval: checkInterval,
		MaxTimeout:    maxTimeout,
	}, nil
}

// 从成功任务生成新任务的函数类型
type GenerateTasksFunc func(completedTask *Task) []*Task

// ========== 框架核心结构体 ==========
type TaskFramework struct {
	mu sync.RWMutex   // 主锁：保护所有共享数据
	wg sync.WaitGroup // 等待组：确保优雅退出

	tasks         map[string]*Task    // 所有任务（按key存储）
	forbiddenKeys map[string]struct{} // 禁止执行的任务key

	taskFrameworkConfig *TaskFrameworkConfig // 框架配置
	generateFunc        GenerateTasksFunc    // 生成新任务的函数
	executorCtx         context.Context      // 框架上下文
	executorCancel      context.CancelFunc   // 框架取消函数

	cycleStartTime time.Time // 当前周期开始时间
	cycleCount     uint64    // 当前周期执行次数（用于测试）
}

// ========== 框架初始化 ==========
func NewTaskFramework(ctx context.Context, taskFrameworkConfig *TaskFrameworkConfig) (*TaskFramework, error) {
	if taskFrameworkConfig == nil {
		return nil, errors.New("taskFrameworkConfig cannot be nil")
	}
	belogs.Debug("NewTaskFramework(): input taskFrameworkConfig: ", jsonutil.MarshalJson(taskFrameworkConfig))

	// 初始化框架
	executorCtx, cancel := context.WithCancel(ctx)
	return &TaskFramework{
		tasks:               make(map[string]*Task),
		forbiddenKeys:       make(map[string]struct{}),
		taskFrameworkConfig: taskFrameworkConfig,
		executorCtx:         executorCtx,
		executorCancel:      cancel,
	}, nil
}

// ========== 框架配置方法 ==========
func (f *TaskFramework) SetGenerateTasksFunc(fn GenerateTasksFunc) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.generateFunc = fn
}

// 添加禁止执行的任务key
func (f *TaskFramework) AddForbiddenKeys(keys ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, key := range keys {
		if key != "" {
			f.forbiddenKeys[key] = struct{}{}
		}
	}
	belogs.Debug("AddForbiddenKeys(): added keys: ", keys)
}

// 移除禁止执行的任务key
func (f *TaskFramework) RemoveForbiddenKeys(keys ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, key := range keys {
		delete(f.forbiddenKeys, key)
	}
	belogs.Debug("RemoveForbiddenKeys(): removed keys: ", keys)
}

// ========== 任务添加方法 ==========
// AddTasks 批量添加任务
// 返回：成功添加数、失败任务列表（包含失败原因）
func (f *TaskFramework) AddTasks(tasks []*Task, executeFunc func(ctx context.Context, task *Task) TaskExecutionResult) (int, []*Task, error) {
	belogs.Debug("AddTasks(): input tasks count: ", len(tasks))

	// 校验执行函数
	if executeFunc == nil {
		belogs.Error("AddTasks(): executeFunc cannot be nil")
		return 0, nil, errors.New("executeFunc cannot be nil")
	}
	if len(tasks) == 0 {
		belogs.Debug("AddTasks(): no tasks to add")
		return 0, nil, nil
	}

	successCount := 0
	failedTasks := make([]*Task, 0)

	f.mu.Lock()
	defer f.mu.Unlock()

	for _, task := range tasks {
		// 基础校验
		if task.Key == "" {
			task.TaskResult.Result = TaskResultFail
			task.TaskResult.ResultReason = "task key cannot be empty"
			task.TaskResult.ResultTime = time.Now()
			task.TaskResult.ResultCount++
			failedTasks = append(failedTasks, task)
			continue
		}

		// 排重校验：禁止key + 已存在任务
		if _, forbidden := f.forbiddenKeys[task.Key]; forbidden {
			task.TaskResult.Result = TaskResultFail
			task.TaskResult.ResultReason = fmt.Sprintf("task %s is in forbidden list", task.Key)
			task.TaskResult.ResultTime = time.Now()
			task.TaskResult.ResultCount++
			failedTasks = append(failedTasks, task)
			continue
		}
		if _, exists := f.tasks[task.Key]; exists {
			task.TaskResult.Result = TaskResultFail
			task.TaskResult.ResultReason = fmt.Sprintf("task %s already exists (any state)", task.Key)
			task.TaskResult.ResultTime = time.Now()
			task.TaskResult.ResultCount++
			failedTasks = append(failedTasks, task)
			continue
		}

		// 初始化任务
		task.executeFunc = executeFunc

		// 根据模式设置初始状态
		switch f.taskFrameworkConfig.AddTaskMode {
		case AddTaskModeRecursive:
			// 递归模式：立即执行（状态设为running）
			task.TaskState = TaskStateRunning
			task.StartTime = time.Now()
			f.tasks[task.Key] = task
			successCount++

			// 异步执行任务，立即执行
			f.wg.Add(1)
			go f.executeTask(task)

		case AddTaskModeExternal:
			// 外部模式：待执行（状态设为pending）
			task.TaskState = TaskStatePending
			f.tasks[task.Key] = task
			successCount++
		}

		belogs.Debug("AddTasks(): added task successfully: ", task.Key)
	}

	return successCount, failedTasks, nil
}

// ========== 任务执行核心逻辑 ==========
func (f *TaskFramework) executeTask(task *Task) {
	defer f.wg.Done()

	// 超时控制（70分钟）
	ctx, cancel := context.WithTimeout(f.executorCtx, f.taskFrameworkConfig.MaxTimeout)
	defer cancel()

	// 执行任务
	resultChan := make(chan TaskExecutionResult, 1)
	go func() {
		taskExecutionResult := task.executeFunc(ctx, task)
		resultChan <- taskExecutionResult
	}()

	// 等待执行结果/超时/框架停止
	select {
	case <-f.executorCtx.Done():
		// 框架停止
		f.mu.Lock()
		now := time.Now()
		task.TaskState = TaskStateCompleted
		task.TaskResult.Result = TaskResultFail
		task.TaskResult.ResultTime = now
		task.TaskResult.ResultReason = "framework stopped"
		task.TaskResult.ResultCount++
		f.mu.Unlock()
		belogs.Info("executeTask(): framework stopped for task ", task.Key)

	case <-ctx.Done():
		// 任务超时（需求3）
		f.mu.Lock()
		now := time.Now()
		task.TaskState = TaskStateCompleted
		task.TaskResult.Result = TaskResultFail
		task.TaskResult.ResultTime = now
		task.TaskResult.ResultReason = fmt.Sprintf("timeout (max %v): %v", f.taskFrameworkConfig.MaxTimeout, ctx.Err())
		task.TaskResult.ResultCount++
		f.mu.Unlock()
		belogs.Info("executeTask(): timeout for task ", task.Key, ": ", task.TaskResult.ResultReason)

	case taskExecutionResult := <-resultChan:
		// 任务执行完成（需求2）
		f.mu.Lock()
		now := time.Now()
		if taskExecutionResult.Result == TaskResultOK {
			// 执行成功，原任务进入completed状态，记录成功时间和次数，下30分钟继续执行
			task.TaskState = TaskStateCompleted
			task.TaskResult.Result = TaskResultOK
			task.TaskResult.ResultTime = now
			task.TaskResult.ResultCount++
			belogs.Info("executeTask(): success for task ", task.Key)

			// 递归模式：生成新任务（需求4.2）
			if f.taskFrameworkConfig.AddTaskMode == AddTaskModeRecursive && f.generateFunc != nil {

				// 拷贝任务数据（避免锁持有过久）
				taskCopy := *task
				go func() {
					newTasks := f.generateFunc(&taskCopy)
					belogs.Debug("executeTask(): generated ", len(newTasks), " new tasks from ", task.Key)
					f.AddTasks(newTasks, task.executeFunc)
				}()
			}
		} else {
			// 执行失败
			task.TaskState = TaskStateCompleted
			task.TaskResult.Result = TaskResultFail
			task.TaskResult.ResultTime = now
			task.TaskResult.ResultReason = taskExecutionResult.Err
			task.TaskResult.ResultCount++
			belogs.Info("executeTask(): fail for task ", task.Key, ": ", taskExecutionResult.Err)
		}
		f.mu.Unlock()
	}
}

// ========== 周期调度逻辑 ==========
// Start 启动框架
func (f *TaskFramework) Start() {
	// 启动周期执行器
	f.cycleCount = 0
	go f.cycleExecutor()
	belogs.Info("TaskFramework started: ", jsonutil.MarshalJson(f.taskFrameworkConfig))
}

// Stop 停止框架（优雅退出）
func (f *TaskFramework) Stop() {
	f.executorCancel()
	f.wg.Wait()
	belogs.Info("TaskFramework stopped gracefully")
}

// cycleExecutor 周期执行器（每30分钟触发一次）
func (f *TaskFramework) cycleExecutor() {
	// 立即执行第一个周期
	f.runCycle()

	// 启动周期定时器
	ticker := time.NewTicker(f.taskFrameworkConfig.CycleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-f.executorCtx.Done():
			belogs.Info("cycleExecutor stopped")
			return
		case <-ticker.C:
			f.runCycle()
		}
	}
}

// runCycle 执行单个周期（严格符合需求1.3.1）
// 1. 收集所有状态为 pending（待执行）、completed（执行成功/失败）的任务
// 2. 排除禁止执行的任务
// 3. 将这些任务改为 running 状态并异步执行
// 4. 10分钟后触发检查逻辑
func (f *TaskFramework) runCycle() {
	f.mu.Lock()
	f.cycleCount++
	f.cycleStartTime = time.Now()
	f.mu.Unlock()

	belogs.Info("runCycle(): start new cycle,cycleCount:", f.cycleCount, " cycleStartTime:", f.cycleStartTime)

	// 1. 收集所有待处理的任务：pending + completed（排除禁止key和running）
	f.mu.RLock()
	tasksToRun := make([]*Task, 0)
	for _, task := range f.tasks {
		// 排除禁止执行的任务
		if _, forbidden := f.forbiddenKeys[task.Key]; forbidden {
			continue
		}
		// 仅处理 pending 或 completed 状态的任务
		if task.TaskState == TaskStatePending || task.TaskState == TaskStateCompleted {
			belogs.Debug("runCycle(): task will add to run, key:", task.Key, " state:", task.TaskState)
			tasksToRun = append(tasksToRun, task)
		}
	}
	f.mu.RUnlock()
	belogs.Info("runCycle(): found len(tasksToRun):", len(tasksToRun), " tasks to run (pending/completed)")

	// 2. 将任务改为 running 并异步执行
	for _, task := range tasksToRun {
		f.mu.Lock()
		// 再次校验状态（避免并发修改）
		if task.TaskState == TaskStatePending || task.TaskState == TaskStateCompleted {
			task.TaskState = TaskStateRunning
			task.StartTime = time.Now()
		}
		f.mu.Unlock()

		// 异步执行任务
		f.wg.Add(1)
		go f.executeTask(task)
	}

	// 3. 10分钟后检查任务状态（需求1.3.2）
	go func() {
		time.Sleep(f.taskFrameworkConfig.CheckInterval)
		f.checkCycleTasks(f.cycleStartTime)
	}()
}

// checkCycleTasks 周期内检查任务状态（需求1.3.2）
func (f *TaskFramework) checkCycleTasks(cycleStart time.Time) {
	belogs.Info("checkCycleTasks(): check tasks for cycle at ", cycleStart)

	// 1. 检查运行中任务：未完成则无需处理（下周期自动执行）
	f.mu.RLock()
	for _, task := range f.tasks {
		if task.TaskState == TaskStateRunning {
			belogs.Info("checkCycleTasks(): task still running: ", task.Key)
		}
	}
	f.mu.RUnlock()

	// 2. 递归模式：检查本周期内完成的成功任务，生成新任务（需求2.3）
	if f.taskFrameworkConfig.AddTaskMode == AddTaskModeRecursive && f.generateFunc != nil {
		f.mu.RLock()
		completedSuccessTasks := make([]*Task, 0)
		for _, task := range f.tasks {
			if task.TaskState == TaskStateCompleted && task.TaskResult.Result == TaskResultOK {
				belogs.Debug("checkCycleTasks(): task completed successfully in this cycle: ", task.Key)
				completedSuccessTasks = append(completedSuccessTasks, task)
			}
		}
		f.mu.RUnlock()

		// 生成新任务
		for _, task := range completedSuccessTasks {
			taskCopy := *task
			go func() {
				newTasks := f.generateFunc(&taskCopy)
				belogs.Debug("checkCycleTasks(): generated ", len(newTasks), " new tasks from ", task.Key)
				f.AddTasks(newTasks, task.executeFunc)
			}()
		}
	}
}
