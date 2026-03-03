package taskcycleutil

import (
	"context"
	"time"
)

const (
	TaskResultOK   = "ok"   // 执行成功
	TaskResultFail = "fail" // 执行失败

	TaskStatePending   = "pending"   // 待执行
	TaskStateRunning   = "running"   // 执行中
	TaskStateCompleted = "completed" // 执行完毕（成功/失败）
)

// ========== 任务结构体定义 ==========
type TaskParam struct {
	Content string                 `json:"content"`
	Params  map[string]interface{} `json:"params"`
}
type TaskResult struct {
	Result       string                 `json:"result"`      // 执行结果：ok/fail
	ResultTime   time.Time              `json:"successTime"` // 成功/失败时间
	ResultReason string                 `json:"reason"`      // 成功/失败原因
	ResultCount  uint64                 `json:"resultCount"` // 执行结果次数（成功或失败）
	Results      map[string]interface{} `json:"results"`
}

type Task struct {
	Key       string    `json:"key"`       // 任务唯一标识（排重）
	StartTime time.Time `json:"startTime"` // 开始执行时间

	TaskParam  TaskParam  `json:"param"`  // 自定义任务数据
	TaskState  string     `json:"state"`  // 任务状态
	TaskResult TaskResult `json:"result"` // 执行结果

	executeFunc func(ctx context.Context, task *Task) TaskExecutionResult `json:"-"` // 执行结果：ok/fail 任务执行函数

}

func NewTask(key string, param TaskParam, executeFunc func(ctx context.Context, task *Task) TaskExecutionResult) *Task {
	return &Task{
		Key:         key,
		StartTime:   time.Now(),
		TaskParam:   param,
		TaskState:   TaskStatePending,
		executeFunc: executeFunc,
	}
}

type TaskExecutionResult struct {
	Result string `json:"result"` // 执行结果：ok/fail
	Err    string `json:"err"`    // err.Error()
}
