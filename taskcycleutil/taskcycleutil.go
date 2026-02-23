package taskcycleutil

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

// ========== 核心枚举定义 ==========
type TaskState string

const (
	TaskStatePending   TaskState = "pending"
	TaskStateRunning   TaskState = "running"
	TaskStateCompleted TaskState = "completed"
)

type TaskResult string

const (
	TaskResultOK   TaskResult = "ok"
	TaskResultFail TaskResult = "fail"
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
	MaxTimeout    time.Duration // 最长执行超时（默认70分钟）
	MaxConcurrent int           // 最大并发执行任务数（默认100）
}

func NewConfig(mode AddTaskMode) *Config {
	return &Config{
		Mode:          mode,
		CycleInterval: 30 * time.Minute,
		CheckInterval: 10 * time.Minute,
		MaxTimeout:    70 * time.Minute,
		MaxConcurrent: 100,
	}
}

// ========== 任务结构体定义 ==========
type TaskData struct {
	Content string
	Params  map[string]interface{}
}

type Task struct {
	Key          string
	Data         TaskData
	State        TaskState
	Result       TaskResult
	SuccessTime  *time.Time
	FailTime     *time.Time
	FailReason   string
	SuccessCount uint64
	FailCount    uint64
	StartTime    *time.Time
	executeFunc  func(ctx context.Context, task *Task) (bool, error)
}

type GenerateTasksFunc func(completedTask *Task) []*Task

// ========== 框架核心结构体 ==========
type TaskFramework struct {
	tasksMu        sync.RWMutex
	tasks          map[string]*Task
	pendingTasks   map[string]struct{}
	completedTasks map[string]struct{}

	forbiddenMu   sync.RWMutex
	forbiddenKeys map[string]struct{}

	configMu          sync.RWMutex
	config            *Config
	generateMu        sync.RWMutex
	generateTasksFunc GenerateTasksFunc

	cycleMu           sync.RWMutex
	currentCycleStart time.Time
	executorCtx       context.Context
	executorCancel    context.CancelFunc
	wg                sync.WaitGroup
	semaphore         chan struct{}
}

// ========== NewTaskFramework ==========
func NewTaskFramework(ctx context.Context, config *Config) *TaskFramework {
	if config == nil {
		panic("config cannot be nil")
	}
	if config.Mode != AddTaskModeRecursive && config.Mode != AddTaskModeExternal {
		panic("invalid add task mode")
	}
	if config.CycleInterval <= 0 {
		config.CycleInterval = 30 * time.Minute
	}
	if config.CheckInterval <= 0 {
		config.CheckInterval = 10 * time.Minute
	}
	if config.MaxTimeout <= 0 {
		config.MaxTimeout = 70 * time.Minute
	}
	if config.MaxConcurrent <= 0 {
		config.MaxConcurrent = 100
	}
	if config.CheckInterval >= config.CycleInterval {
		panic("check interval must be less than cycle interval")
	}

	executorCtx, executorCancel := context.WithCancel(context.Background())

	return &TaskFramework{
		tasks:          make(map[string]*Task),
		pendingTasks:   make(map[string]struct{}),
		completedTasks: make(map[string]struct{}),
		forbiddenKeys:  make(map[string]struct{}),
		config:         config,
		executorCtx:    executorCtx,
		executorCancel: executorCancel,
		semaphore:      make(chan struct{}, config.MaxConcurrent),
	}
}

// ========== 公开API方法 ==========
func (f *TaskFramework) SetGenerateTasksFunc(ctx context.Context, fn GenerateTasksFunc) {
	f.generateMu.Lock()
	defer f.generateMu.Unlock()
	f.generateTasksFunc = fn
}

// ========== 修改：AddTasks 返回失败的任务列表 ==========
// AddTasks 批量添加任务
// ctx: 用于系统记录
// tasks: 待添加的任务列表
// executeFunc: 任务执行函数
// 返回:
//
//	successCount: 成功添加的任务数量
//	failedTasks: 失败的任务列表（每个任务的 FailReason 字段会填充失败原因）
func (f *TaskFramework) AddTasks(ctx context.Context, tasks []*Task, executeFunc func(ctx context.Context, task *Task) (bool, error)) (successCount int, failedTasks []*Task) {
	// 🔧 修改：不再返回 error，而是返回失败的任务列表

	if executeFunc == nil {
		// 如果执行函数为nil，所有任务都失败
		for _, task := range tasks {
			task.FailReason = "executeFunc cannot be nil"
			failedTasks = append(failedTasks, task)
		}
		return 0, failedTasks
	}

	if len(tasks) == 0 {
		return 0, nil
	}

	f.configMu.RLock()
	mode := f.config.Mode
	f.configMu.RUnlock()

	f.tasksMu.Lock()
	defer f.tasksMu.Unlock()

	f.forbiddenMu.RLock()
	defer f.forbiddenMu.RUnlock()

	// 🔧 修改：收集失败的任务而不是错误字符串
	for _, task := range tasks {
		if task.Key == "" {
			task.FailReason = "task key cannot be empty"
			failedTasks = append(failedTasks, task)
			continue
		}
		if _, forbidden := f.forbiddenKeys[task.Key]; forbidden {
			task.FailReason = fmt.Sprintf("task %s is forbidden", task.Key)
			failedTasks = append(failedTasks, task)
			continue
		}
		if _, exists := f.tasks[task.Key]; exists {
			task.FailReason = fmt.Sprintf("task %s already exists", task.Key)
			failedTasks = append(failedTasks, task)
			continue
		}

		// 成功添加的任务
		task.executeFunc = executeFunc
		task.SuccessCount = 0
		task.FailCount = 0
		task.State = TaskStatePending
		task.Result = ""
		task.SuccessTime = nil
		task.FailTime = nil
		task.FailReason = ""
		task.StartTime = nil

		switch mode {
		case AddTaskModeRecursive:
			now := time.Now()
			task.State = TaskStateRunning
			task.StartTime = &now
			f.tasks[task.Key] = task

			successCount++
			f.wg.Add(1)
			go f.executeTask(ctx, task)

		case AddTaskModeExternal:
			f.tasks[task.Key] = task
			f.pendingTasks[task.Key] = struct{}{}
			successCount++

		default:
			task.FailReason = fmt.Sprintf("invalid mode for task %s", task.Key)
			failedTasks = append(failedTasks, task)
		}
	}

	return successCount, failedTasks
}

func (f *TaskFramework) AddForbiddenKeys(ctx context.Context, keys ...string) {
	f.forbiddenMu.Lock()
	defer f.forbiddenMu.Unlock()
	for _, key := range keys {
		if key == "" {
			continue
		}
		f.forbiddenKeys[key] = struct{}{}
	}
}

func (f *TaskFramework) RemoveForbiddenKeys(ctx context.Context, keys ...string) {
	f.forbiddenMu.Lock()
	defer f.forbiddenMu.Unlock()
	for _, key := range keys {
		delete(f.forbiddenKeys, key)
	}
}

// ========== 内部方法（保持不变） ==========
func (f *TaskFramework) updateTaskState(ctx context.Context, task *Task, newState TaskState) {
	if task == nil {
		return
	}
	oldState := task.State
	if oldState == newState {
		return
	}

	switch oldState {
	case TaskStatePending:
		delete(f.pendingTasks, task.Key)
	case TaskStateCompleted:
		delete(f.completedTasks, task.Key)
	}

	task.State = newState

	switch newState {
	case TaskStatePending:
		f.pendingTasks[task.Key] = struct{}{}
	case TaskStateCompleted:
		f.completedTasks[task.Key] = struct{}{}
	}
}

func (f *TaskFramework) executeTask(ctx context.Context, task *Task) {
	defer f.wg.Done()

	defer func() {
		if r := recover(); r != nil {
			f.tasksMu.Lock()
			defer f.tasksMu.Unlock()

			now := time.Now()
			f.updateTaskState(ctx, task, TaskStateCompleted)
			task.Result = TaskResultFail
			task.FailTime = &now
			task.FailReason = fmt.Sprintf("execute panic: %v\nstack: %s", r, debug.Stack())
			task.FailCount++
		}
	}()

	f.configMu.RLock()
	maxTimeout := f.config.MaxTimeout
	mode := f.config.Mode
	checkInterval := f.config.CheckInterval
	f.configMu.RUnlock()

	select {
	case f.semaphore <- struct{}{}:
		defer func() { <-f.semaphore }()
	case <-f.executorCtx.Done():
		return
	}

	taskCtx, cancel := context.WithTimeout(f.executorCtx, maxTimeout)
	defer cancel()

	resultChan := make(chan struct {
		success bool
		err     error
	}, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- struct {
					success bool
					err     error
				}{
					success: false,
					err:     fmt.Errorf("executeFunc panic: %v\nstack: %s", r, debug.Stack()),
				}
			}
		}()
		success, err := task.executeFunc(taskCtx, task)
		resultChan <- struct {
			success bool
			err     error
		}{success, err}
	}()

	select {
	case <-f.executorCtx.Done():
		f.tasksMu.Lock()
		now := time.Now()
		f.updateTaskState(ctx, task, TaskStateCompleted)
		task.Result = TaskResultFail
		task.FailTime = &now
		task.FailReason = "framework stopped"
		task.FailCount++
		f.tasksMu.Unlock()
		return

	case <-taskCtx.Done():
		f.tasksMu.Lock()
		now := time.Now()
		f.updateTaskState(ctx, task, TaskStateCompleted)
		task.Result = TaskResultFail
		task.FailTime = &now
		task.FailReason = fmt.Sprintf("timeout after %v: %v", maxTimeout, taskCtx.Err())
		task.FailCount++
		f.tasksMu.Unlock()
		return

	case res := <-resultChan:
		f.tasksMu.Lock()
		now := time.Now()

		if res.success {
			task.Result = TaskResultOK
			task.SuccessTime = &now
			task.SuccessCount++

			var needGenerate bool
			var generateFunc GenerateTasksFunc
			var cycleStart time.Time
			var taskCopyForGenerate *Task

			if mode == AddTaskModeRecursive {
				f.updateTaskState(ctx, task, TaskStatePending)
				f.generateMu.RLock()
				generateFunc = f.generateTasksFunc
				f.generateMu.RUnlock()

				f.cycleMu.RLock()
				cycleStart = f.currentCycleStart
				f.cycleMu.RUnlock()

				if !cycleStart.IsZero() && generateFunc != nil {
					needGenerate = now.After(cycleStart) && now.Before(cycleStart.Add(checkInterval))
				}
				if needGenerate {
					taskCopyForGenerate = &Task{
						Key:         task.Key,
						Data:        task.Data,
						Result:      task.Result,
						SuccessTime: task.SuccessTime,
						executeFunc: task.executeFunc,
					}
				}
			} else {
				f.updateTaskState(ctx, task, TaskStateCompleted)
			}

			f.tasksMu.Unlock()

			if needGenerate && mode == AddTaskModeRecursive && generateFunc != nil && taskCopyForGenerate != nil {
				go func() {
					defer func() {
						_ = recover()
					}()
					newTasks := generateFunc(taskCopyForGenerate)
					_, _ = f.AddTasks(ctx, newTasks, taskCopyForGenerate.executeFunc)
				}()
			}
			return
		} else {
			f.updateTaskState(ctx, task, TaskStateCompleted)
			task.Result = TaskResultFail
			task.FailTime = &now
			task.FailReason = res.err.Error()
			task.FailCount++
		}

		f.tasksMu.Unlock()
	}
}

func (f *TaskFramework) cycleExecutor(ctx context.Context) {
	f.triggerCycle(ctx)

	f.configMu.RLock()
	cycleInterval := f.config.CycleInterval
	f.configMu.RUnlock()

	cycleTicker := time.NewTicker(cycleInterval)
	defer cycleTicker.Stop()

	for {
		select {
		case <-f.executorCtx.Done():
			return
		case <-cycleTicker.C:
			f.triggerCycle(ctx)
		}
	}
}

func (f *TaskFramework) triggerCycle(ctx context.Context) {
	now := time.Now()
	f.cycleMu.Lock()
	f.currentCycleStart = now
	f.cycleMu.Unlock()

	f.configMu.RLock()
	checkInterval := f.config.CheckInterval
	f.configMu.RUnlock()

	f.tasksMu.RLock()
	taskKeysToRun := make([]string, 0, len(f.pendingTasks)+len(f.completedTasks))
	for key := range f.pendingTasks {
		taskKeysToRun = append(taskKeysToRun, key)
	}
	for key := range f.completedTasks {
		taskKeysToRun = append(taskKeysToRun, key)
	}
	f.tasksMu.RUnlock()

	for _, key := range taskKeysToRun {
		f.forbiddenMu.RLock()
		_, forbidden := f.forbiddenKeys[key]
		f.forbiddenMu.RUnlock()
		if forbidden {
			continue
		}

		f.tasksMu.Lock()
		task, exists := f.tasks[key]
		if !exists || (task.State != TaskStatePending && task.State != TaskStateCompleted) {
			f.tasksMu.Unlock()
			continue
		}
		f.updateTaskState(ctx, task, TaskStateRunning)
		task.StartTime = &now
		f.tasksMu.Unlock()

		f.wg.Add(1)
		go f.executeTask(ctx, task)
	}

	go func() {
		time.Sleep(checkInterval)
		f.triggerCheck(ctx)
	}()
}

func (f *TaskFramework) triggerCheck(ctx context.Context) {
	f.configMu.RLock()
	mode := f.config.Mode
	checkInterval := f.config.CheckInterval
	f.configMu.RUnlock()

	f.generateMu.RLock()
	generateFunc := f.generateTasksFunc
	f.generateMu.RUnlock()

	if mode != AddTaskModeRecursive || generateFunc == nil {
		return
	}

	f.cycleMu.RLock()
	cycleStart := f.currentCycleStart
	f.cycleMu.RUnlock()
	if cycleStart.IsZero() {
		return
	}

	f.tasksMu.RLock()
	tasksToCheck := make([]*Task, 0)
	for key := range f.completedTasks {
		task := f.tasks[key]
		if task.Result == TaskResultOK {
			tasksToCheck = append(tasksToCheck, task)
		}
	}
	f.tasksMu.RUnlock()

	for _, task := range tasksToCheck {
		f.tasksMu.RLock()
		completeTime := task.SuccessTime
		currentTask, exists := f.tasks[task.Key]
		executeFunc := task.executeFunc
		f.tasksMu.RUnlock()

		if !exists || completeTime == nil {
			continue
		}
		if !completeTime.After(cycleStart) || !completeTime.Before(cycleStart.Add(checkInterval)) {
			continue
		}

		go func(t *Task) {
			defer func() {
				_ = recover()
			}()
			newTasks := generateFunc(t)
			_, _ = f.AddTasks(ctx, newTasks, executeFunc)
		}(currentTask)
	}
}

func (f *TaskFramework) Start(ctx context.Context) {
	f.configMu.RLock()
	config := *f.config
	f.configMu.RUnlock()

	go f.cycleExecutor(ctx)
	fmt.Printf("task framework started, config: mode=%d, cycle=%v, check=%v, timeout=%v, max_concurrent=%d\n",
		config.Mode, config.CycleInterval, config.CheckInterval, config.MaxTimeout, config.MaxConcurrent)
}

func (f *TaskFramework) Stop(ctx context.Context) {
	f.executorCancel()
	f.wg.Wait()
	fmt.Println("task framework stopped gracefully")
}
