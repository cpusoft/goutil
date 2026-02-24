package executil

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

// only use in linux, not support windows
// ExecCommandCombinedOutput 基础版：执行命令并返回组合输出（stdout+stderr）
// 调用方需自己保证命令的安全性和正确性，避免不必要的路径解析和权限问题
func ExecCommandCombinedOutput(commandName string, params []string) (out string, err error) {
	// 直接创建命令（不再强制解析绝对路径）
	result := exec.Command(commandName, params...)
	b, err := result.CombinedOutput()
	out = string(b)

	// 仅保留基础日志和错误返回
	if err != nil {
		belogs.Error("ExecCommandCombinedOutput: CombinedOutput fail, commandName:", commandName,
			"params:", jsonutil.MarshalJson(params), "out:", out, err)
		return out, err
	}

	belogs.Debug("ExecCommandCombinedOutput: success, commandName:", commandName,
		"params:", jsonutil.MarshalJson(params), "out:", out)
	return out, nil
}

// only use in linux, not support windows
// ExecCommandStdoutPipe 基础版：执行命令并逐行读取stdout
// 调用方需自己保证命令的安全性和正确性，避免不必要的路径解析和权限问题
func ExecCommandStdoutPipe(commandName string, params []string, fmtShow bool) (contentArray []string, err error) {
	contentArray = make([]string, 0)
	cmd := exec.Command(commandName, params...)

	// 显示运行的命令（保留基础打印）
	if fmtShow {
		fmt.Printf("exec:%s\n", strings.Join(cmd.Args[:], " "))
	}

	// 创建stdout管道（仅基础错误处理）
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		if fmtShow {
			fmt.Fprintln(os.Stderr, "error=>", err.Error())
		}
		return contentArray, err
	}
	defer stdout.Close() // 仅保留基础资源释放

	// 启动命令
	if err := cmd.Start(); err != nil {
		if fmtShow {
			fmt.Fprintln(os.Stderr, "error=> start command fail:", err.Error())
		}
		return contentArray, err
	}

	// 逐行读取stdout（简化逻辑）
	reader := bufio.NewReader(stdout)
	for {
		tmp, _, err2 := reader.ReadLine()
		line := string(tmp)
		if err2 != nil || io.EOF == err2 {
			break
		}
		if fmtShow {
			fmt.Println(line)
		}
		contentArray = append(contentArray, line)
	}

	// 等待命令完成（不校验退出码）
	cmd.Wait()
	return contentArray, nil
}

// Deprecated: using ExecCommandStdoutPipe
func ExecCommand(commandName string, params []string, fmtShow bool) (contentArray []string, err error) {
	return ExecCommandStdoutPipe(commandName, params, fmtShow)
}
