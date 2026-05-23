package crontabutil

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// CronTab 封装 crontab 操作的结构体
type CronTab struct {
	Tasks []string // 存储所有定时任务（每行一个）
}

// Load 读取当前用户的 crontab 任务列表
func (c *CronTab) Load() error {
	// 执行 crontab -l 命令获取任务
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 处理 "no crontab for user" 错误（无任务时的正常情况）
		if strings.Contains(string(output), "no crontab for") {
			c.Tasks = []string{}
			return nil
		}
		return fmt.Errorf("读取 crontab 失败: %v, 输出: %s", err, output)
	}

	// 将输出按行分割为任务列表（过滤空行）
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			c.Tasks = append(c.Tasks, line)
		}
	}
	return scanner.Err()
}

// AddTask 添加新的定时任务（自动去重）
func (c *CronTab) AddTask(task string) error {
	// 校验任务语法（简单校验：必须包含 5 个空格分隔的时间字段 + 命令）
	task = strings.TrimSpace(task)
	if task == "" {
		return errors.New("任务内容不能为空")
	}
	parts := strings.Fields(task)
	if len(parts) < 6 {
		return errors.New("任务格式错误，需符合 crontab 语法：分 时 日 月 周 命令")
	}

	// 检查是否已存在相同任务，避免重复
	for _, t := range c.Tasks {
		if strings.TrimSpace(t) == task {
			return fmt.Errorf("任务已存在：%s", task)
		}
	}

	c.Tasks = append(c.Tasks, task)
	return nil
}

// RemoveTask 删除指定的定时任务（按内容匹配）
func (c *CronTab) RemoveTask(task string) error {
	task = strings.TrimSpace(task)
	newTasks := []string{}
	found := false

	for _, t := range c.Tasks {
		if strings.TrimSpace(t) != task {
			newTasks = append(newTasks, t)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("未找到要删除的任务：%s", task)
	}

	c.Tasks = newTasks
	return nil
}

// Save 将修改后的任务列表写入 crontab
func (c *CronTab) Save() error {
	// 将任务列表拼接为字符串（每行一个任务）
	taskStr := strings.Join(c.Tasks, "\n") + "\n"

	// 执行 crontab - 命令，从标准输入写入配置
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(taskStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("写入 crontab 失败: %v, 输出: %s", err, output)
	}
	return nil
}
