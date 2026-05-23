package crontabutil

import (
	"fmt"
	"testing"
)

// 示例：使用封装的 CronTab 操作 crontab
func TestCrontabAll(t *testing.T) {
	// 1. 初始化并加载现有任务
	ct := &CronTab{}
	if err := ct.Load(); err != nil {
		fmt.Printf("加载 crontab 失败：%v\n", err)
		return
	}
	fmt.Println("当前 crontab 任务：")
	for i, task := range ct.Tasks {
		fmt.Printf("%d: %s\n", i+1, task)
	}

	// 2. 添加新任务（示例：每分钟执行 echo 命令，输出到日志）
	newTask := "* * * * * echo 'Go 程序添加的定时任务' >> /tmp/cron_test.log 2>&1"
	if err := ct.AddTask(newTask); err != nil {
		fmt.Printf("添加任务失败：%v\n", err)
	} else {
		fmt.Println("添加任务成功")
	}

	// 3. 保存修改到系统 crontab
	if err := ct.Save(); err != nil {
		fmt.Printf("保存 crontab 失败：%v\n", err)
		return
	}
	fmt.Println("crontab 保存成功！")

	// 4. 验证：重新加载并打印
	if err := ct.Load(); err != nil {
		fmt.Printf("重新加载 crontab 失败：%v\n", err)
		return
	}
	fmt.Println("修改后的 crontab 任务：")
	for i, task := range ct.Tasks {
		fmt.Printf("%d: %s\n", i+1, task)
	}

	// （可选）删除任务示例
	// if err := ct.RemoveTask(newTask); err != nil {
	// 	fmt.Printf("删除任务失败：%v\n", err)
	// } else {
	// 	fmt.Println("删除任务成功")
	// 	ct.Save() // 删除后需保存
	// }
}
