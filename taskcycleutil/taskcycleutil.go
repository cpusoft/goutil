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
type TaskState string

const (
	TaskStatePending   TaskState = "pending"   // 待执行
	TaskStateRunning   TaskState = "running"   // 执行中
	TaskStateCompleted TaskState = "completed" // 执行完毕（成功/失败）
)

type TaskResult string

const (
	TaskResultOK   TaskResult = "ok"   // 执行成功
	TaskResultFail TaskResult = "fail" // 执行失败
)

type AddTaskMode int

const (
	AddTaskModeRecursive AddTaskMode = 1 // 递归模式：从成功任务生成新任务并立即执行
	AddTaskModeExternal  AddTaskMode = 2 // 外部模式：外部注入任务，待下周期执行
)

// ========== 配置结构体定义 ==========
type Config struct {
	Mode          AddTaskMode   // 任务添加模式（必填）
	CycleInterval time.Duration // 周期间隔（默认30分钟）
	CheckInterval time.Duration // 周期内检查间隔（默认10分钟）
	MaxTimeout    time.Duration // 最长执行超时（默认70分钟=2*30+10）
}

func NewConfig(mode AddTaskMode) *Config {
	return &Config{
		Mode:          mode,
		CycleInterval: 30 * time.Minute,
		CheckInterval: 10 * time.Minute,
		MaxTimeout:    70 * time.Minute, // 2个周期+10分钟
	}
}

// ========== 任务结构体定义 ==========
type TaskData struct {
	Content string                 `json:"content"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

type Task struct {
	Key          string     `json:"key"`          // 任务唯一标识（排重）
	Data         TaskData   `json:"data"`         // 自定义任务数据
	State        TaskState  `json:"state"`        // 任务状态
	Result       TaskResult `json:"result"`       // 执行结果
	SuccessTime  *time.Time `json:"successTime"`  // 成功时间
	FailTime     *time.Time `json:"failTime"`     // 失败时间
	FailReason   string     `json:"failReason"`   // 失败原因
	SuccessCount uint64     `json:"successCount"` // 成功次数
	FailCount    uint64     `json:"failCount"`    // 失败次数
	StartTime    *time.Time `json:"startTime"`    // 开始执行时间

	executeFunc func(ctx context.Context, task *Task) (bool, error) `json:"-"` // 任务执行函数
}

// 从成功任务生成新任务的函数类型
type GenerateTasksFunc func(completedTask *Task) []*Task

// ========== 框架核心结构体 ==========
type TaskFramework struct {
	mu             sync.RWMutex        // 主锁：保护所有共享数据
	tasks          map[string]*Task    // 所有任务（按key存储）
	forbiddenKeys  map[string]struct{} // 禁止执行的任务key
	config         *Config             // 框架配置
	generateFunc   GenerateTasksFunc   // 生成新任务的函数
	executorCtx    context.Context     // 框架上下文
	executorCancel context.CancelFunc  // 框架取消函数
	wg             sync.WaitGroup      // 等待组：确保优雅退出
	currentCycle   time.Time           // 当前周期开始时间
}

// ========== 框架初始化 ==========
func NewTaskFramework(config *Config) (*TaskFramework, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	belogs.Debug("NewTaskFramework(): input config: ", jsonutil.MarshalJson(config))

	// 校验配置
	if config.Mode != AddTaskModeRecursive && config.Mode != AddTaskModeExternal {
		return nil, errors.New("invalid add task mode (must be 1 or 2)")
	}
	// 配置默认值
	if config.CycleInterval <= 0 {
		config.CycleInterval = 30 * time.Minute
	}
	if config.CheckInterval <= 0 || config.CheckInterval >= config.CycleInterval {
		config.CheckInterval = 10 * time.Minute
	}
	if config.MaxTimeout <= 0 {
		config.MaxTimeout = 70 * time.Minute // 2*30+10
	}

	// 初始化框架
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskFramework{
		tasks:          make(map[string]*Task),
		forbiddenKeys:  make(map[string]struct{}),
		config:         config,
		executorCtx:    ctx,
		executorCancel: cancel,
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
func (f *TaskFramework) AddTasks(tasks []*Task, executeFunc func(ctx context.Context, task *Task) (bool, error)) (int, []*Task) {
	belogs.Debug("AddTasks(): input tasks count: ", len(tasks))

	// 校验执行函数
	if executeFunc == nil {
		failed := make([]*Task, len(tasks))
		for i, t := range tasks {
			t.FailReason = "executeFunc cannot be nil"
			failed[i] = t
		}
		return 0, failed
	}

	successCount := 0
	failedTasks := make([]*Task, 0)

	f.mu.Lock()
	defer f.mu.Unlock()

	for _, task := range tasks {
		// 基础校验
		if task.Key == "" {
			task.FailReason = "task key cannot be empty"
			failedTasks = append(failedTasks, task)
			continue
		}

		// 排重校验：禁止key + 已存在任务
		if _, forbidden := f.forbiddenKeys[task.Key]; forbidden {
			task.FailReason = fmt.Sprintf("task %s is in forbidden list", task.Key)
			failedTasks = append(failedTasks, task)
			continue
		}
		if _, exists := f.tasks[task.Key]; exists {
			task.FailReason = fmt.Sprintf("task %s already exists (any state)", task.Key)
			failedTasks = append(failedTasks, task)
			continue
		}

		// 初始化任务
		task.executeFunc = executeFunc
		task.SuccessCount = 0
		task.FailCount = 0
		task.Result = ""
		task.SuccessTime = nil
		task.FailTime = nil
		task.FailReason = ""
		task.StartTime = nil

		// 根据模式设置初始状态
		switch f.config.Mode {
		case AddTaskModeRecursive:
			// 递归模式：立即执行（状态设为running）
			now := time.Now()
			task.State = TaskStateRunning
			task.StartTime = &now
			f.tasks[task.Key] = task
			successCount++

			// 异步执行任务
			f.wg.Add(1)
			go f.executeTask(task)

		case AddTaskModeExternal:
			// 外部模式：待执行（状态设为pending）
			task.State = TaskStatePending
			f.tasks[task.Key] = task
			successCount++
		}

		belogs.Debug("AddTasks(): added task successfully: ", task.Key)
	}

	return successCount, failedTasks
}

// ========== 任务执行核心逻辑 ==========
func (f *TaskFramework) executeTask(task *Task) {
	defer f.wg.Done()

	// 超时控制（70分钟）
	ctx, cancel := context.WithTimeout(f.executorCtx, f.config.MaxTimeout)
	defer cancel()

	// 执行任务
	resultChan := make(chan struct {
		success bool
		err     error
	}, 1)
	go func() {
		success, err := task.executeFunc(ctx, task)
		resultChan <- struct {
			success bool
			err     error
		}{success, err}
	}()

	// 等待执行结果/超时/框架停止
	select {
	case <-f.executorCtx.Done():
		// 框架停止
		f.mu.Lock()
		now := time.Now()
		task.State = TaskStateCompleted
		task.Result = TaskResultFail
		task.FailTime = &now
		task.FailReason = "framework stopped"
		task.FailCount++
		f.mu.Unlock()
		belogs.Warn("executeTask(): framework stopped for task ", task.Key)

	case <-ctx.Done():
		// 任务超时（需求3）
		f.mu.Lock()
		now := time.Now()
		task.State = TaskStateCompleted
		task.Result = TaskResultFail
		task.FailTime = &now
		task.FailReason = fmt.Sprintf("timeout (max %v): %v", f.config.MaxTimeout, ctx.Err())
		task.FailCount++
		f.mu.Unlock()
		belogs.Warn("executeTask(): timeout for task ", task.Key, ": ", task.FailReason)

	case res := <-resultChan:
		// 任务执行完成（需求2）
		f.mu.Lock()
		now := time.Now()
		if res.success {
			// 执行成功
			task.State = TaskStateCompleted
			task.Result = TaskResultOK
			task.SuccessTime = &now
			task.SuccessCount++
			belogs.Info("executeTask(): success for task ", task.Key)

			// 递归模式：生成新任务（需求4.2）
			if f.config.Mode == AddTaskModeRecursive && f.generateFunc != nil {
				// 1. 原任务改回待执行（下周期处理）
				task.State = TaskStatePending
				task.Result = "" // 清空结果，等待下周期
				task.SuccessTime = nil

				// 2. 生成新任务并立即执行
				// 拷贝任务数据（避免锁持有过久）
				taskCopy := *task
				go func() {
					newTasks := f.generateFunc(&taskCopy)
					belogs.Debug("executeTask(): generated ", len(newTasks), " new tasks from ", task.Key)
					_, _ = f.AddTasks(newTasks, task.executeFunc)
				}()
			}
		} else {
			// 执行失败
			task.State = TaskStateCompleted
			task.Result = TaskResultFail
			task.FailTime = &now
			task.FailReason = res.err.Error()
			task.FailCount++
			belogs.Warn("executeTask(): fail for task ", task.Key, ": ", res.err)
		}
		f.mu.Unlock()
	}
}

// ========== 周期调度逻辑 ==========
// Start 启动框架
func (f *TaskFramework) Start() {
	// 启动周期执行器
	go f.cycleExecutor()
	belogs.Info("TaskFramework started: ", jsonutil.MarshalJson(f.config))
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
	ticker := time.NewTicker(f.config.CycleInterval)
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
	f.currentCycle = time.Now() // 记录当前周期开始时间
	cycleStart := f.currentCycle
	f.mu.Unlock()

	belogs.Info("runCycle(): start new cycle at ", cycleStart)

	// 1. 收集所有待处理的任务：pending + completed（排除禁止key和running）
	f.mu.RLock()
	tasksToRun := make([]*Task, 0)
	for _, task := range f.tasks {
		// 排除禁止执行的任务
		if _, forbidden := f.forbiddenKeys[task.Key]; forbidden {
			continue
		}
		// 仅处理 pending 或 completed 状态的任务
		if task.State == TaskStatePending || task.State == TaskStateCompleted {
			tasksToRun = append(tasksToRun, task)
		}
	}
	f.mu.RUnlock()

	belogs.Info("runCycle(): found ", len(tasksToRun), " tasks to run (pending/completed)")

	// 2. 将任务改为 running 并异步执行
	for _, task := range tasksToRun {
		f.mu.Lock()
		// 再次校验状态（避免并发修改）
		if task.State == TaskStatePending || task.State == TaskStateCompleted {
			task.State = TaskStateRunning
			now := time.Now()
			task.StartTime = &now
		}
		f.mu.Unlock()

		// 异步执行任务（保留70分钟超时）
		f.wg.Add(1)
		go f.executeTask(task)
	}

	// 3. 10分钟后检查任务状态（需求1.3.2）
	go func() {
		time.Sleep(f.config.CheckInterval)
		f.checkCycleTasks(cycleStart)
	}()
}

// checkCycleTasks 周期内检查任务状态（需求1.3.2）
func (f *TaskFramework) checkCycleTasks(cycleStart time.Time) {
	belogs.Info("checkCycleTasks(): check tasks for cycle at ", cycleStart)

	// 1. 检查运行中任务：未完成则无需处理（下周期自动执行）
	f.mu.RLock()
	runningTasks := make([]*Task, 0)
	for _, task := range f.tasks {
		if task.State == TaskStateRunning {
			runningTasks = append(runningTasks, task)
		}
	}
	f.mu.RUnlock()

	belogs.Info("checkCycleTasks(): ", len(runningTasks), " tasks still running (will retry next cycle)")

	// 2. 递归模式：检查本周期内完成的成功任务，生成新任务（需求2.3）
	if f.config.Mode == AddTaskModeRecursive && f.generateFunc != nil {
		f.mu.RLock()
		completedSuccessTasks := make([]*Task, 0)
		for _, task := range f.tasks {
			if task.State == TaskStateCompleted && task.Result == TaskResultOK && task.SuccessTime != nil {
				// 判断是否在本周期前10分钟内完成
				if task.SuccessTime.After(cycleStart) && task.SuccessTime.Before(cycleStart.Add(f.config.CheckInterval)) {
					completedSuccessTasks = append(completedSuccessTasks, task)
				}
			}
		}
		f.mu.RUnlock()

		// 生成新任务
		for _, task := range completedSuccessTasks {
			taskCopy := *task
			go func() {
				newTasks := f.generateFunc(&taskCopy)
				belogs.Debug("checkCycleTasks(): generated ", len(newTasks), " new tasks from ", task.Key)
				_, _ = f.AddTasks(newTasks, task.executeFunc)
			}()
		}
	}
}
