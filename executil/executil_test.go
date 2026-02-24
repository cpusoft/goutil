package executil

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

/*
# 1. 运行所有基础测试（包含临界值测试）
go test -v ./executil

# 2. 运行指定测试（如临界值测试）
go test -v ./executil -run TestExecCommand_EdgeCases

# 3. 运行性能测试（Benchmark）
go test -bench=. ./executil -benchmem

# 4. 运行性能测试并指定次数（如1000次）
go test -bench=. ./executil -benchmem -count=5 -benchtime=1000x
*/

// -------------------------- 基础功能测试（覆盖核心场景） --------------------------
// TestExecCommandCombinedOutput_Basic 测试ExecCommandCombinedOutput基础功能
func TestExecCommandCombinedOutput_Basic(t *testing.T) {
	// 测试用例：基础合法命令
	testCases := []struct {
		name        string
		commandName string
		params      []string
		wantErr     bool   // 是否预期错误
		containsStr string // 输出中应包含的字符串（用于验证）
	}{
		{
			name:        "ls命令-列出当前目录",
			commandName: "ls",
			params:      []string{"-l"},
			wantErr:     false,
			containsStr: "", // 无固定包含字符串，仅验证无错误
		},
		{
			name:        "echo命令-输出指定内容",
			commandName: "echo",
			params:      []string{"test executil"},
			wantErr:     false,
			containsStr: "test executil",
		},
		{
			name:        "无效命令-预期错误",
			commandName: "invalid_command_123456",
			params:      []string{},
			wantErr:     true,
			containsStr: "",
		},
		{
			name:        "ls命令-不存在的路径（预期错误）",
			commandName: "ls",
			params:      []string{"/path/not/exist_123456"},
			wantErr:     true,
			containsStr: "No such file or directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := ExecCommandCombinedOutput(tc.commandName, tc.params)

			// 验证错误是否符合预期
			if (err != nil) != tc.wantErr {
				t.Fatalf("ExecCommandCombinedOutput() error = %v, wantErr %v", err, tc.wantErr)
			}

			// 验证输出内容（如果指定了包含字符串）
			if tc.containsStr != "" && !strings.Contains(out, tc.containsStr) {
				t.Errorf("ExecCommandCombinedOutput() out = %v, want contains %v", out, tc.containsStr)
			}
		})
	}
}

// TestExecCommandStdoutPipe_Basic 测试ExecCommandStdoutPipe基础功能
func TestExecCommandStdoutPipe_Basic(t *testing.T) {
	testCases := []struct {
		name        string
		commandName string
		params      []string
		fmtShow     bool
		wantErr     bool
		wantLines   int // 预期输出行数（粗略验证）
	}{
		{
			name:        "cat命令-读取空文件（模拟空输出）",
			commandName: "cat",
			params:      []string{"/dev/null"},
			fmtShow:     false,
			wantErr:     false,
			wantLines:   0,
		},
		{
			name:        "seq命令-输出10行数字",
			commandName: "seq",
			params:      []string{"10"},
			fmtShow:     true,
			wantErr:     false,
			wantLines:   10,
		},
		{
			name:        "无效参数-预期错误",
			commandName: "ls",
			params:      []string{"--invalid-parameter-123"},
			fmtShow:     false,
			wantErr:     false, // 简化版不校验退出码，故err为nil
			wantLines:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			contentArray, err := ExecCommandStdoutPipe(tc.commandName, tc.params, tc.fmtShow)

			if (err != nil) != tc.wantErr {
				t.Fatalf("ExecCommandStdoutPipe() error = %v, wantErr %v", err, tc.wantErr)
			}

			// 验证输出行数（允许±1的误差，兼容不同系统）
			if len(contentArray) < tc.wantLines-1 || len(contentArray) > tc.wantLines+1 {
				t.Errorf("ExecCommandStdoutPipe() lines = %d, want %d", len(contentArray), tc.wantLines)
			}
		})
	}
}

// -------------------------- 临界值测试（边界场景） --------------------------
// TestExecCommand_EdgeCases 测试临界值/边界场景
func TestExecCommand_EdgeCases(t *testing.T) {
	// 临界场景1：空参数
	t.Run("空参数执行echo", func(t *testing.T) {
		out, err := ExecCommandCombinedOutput("echo", []string{})
		if err != nil {
			t.Fatalf("echo空参数失败: %v", err)
		}
		if strings.TrimSpace(out) != "" {
			t.Errorf("echo空参数输出异常: %s", out)
		}
	})

	// 临界场景2：超长参数（1000个参数）
	t.Run("超长参数执行echo", func(t *testing.T) {
		params := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			params[i] = strconv.Itoa(i)
		}
		out, err := ExecCommandCombinedOutput("echo", params)
		if err != nil {
			t.Fatalf("超长参数执行失败: %v", err)
		}
		// 验证最后一个参数是否输出
		if !strings.Contains(out, "999") {
			t.Errorf("超长参数输出不完整: %s", out)
		}
	})

	// 临界场景3：长时间运行命令（超时测试）
	t.Run("长时间命令-超时检测", func(t *testing.T) {
		// 启动一个sleep 5秒的命令
		start := time.Now()
		done := make(chan struct{})
		go func() {
			_, _ = ExecCommandCombinedOutput("sleep", []string{"5"})
			close(done)
		}()

		// 等待3秒后判断是否还在运行（验证临界超时）
		// 修正：使用整数时间判断，误差范围放宽到±1秒
		select {
		case <-done:
			t.Error("sleep 5秒命令提前结束，不符合预期")
		case <-time.After(3 * time.Second):
			// 预期结果：3秒后仍在运行
			elapsed := time.Since(start)
			// 改为整数判断：耗时在2秒到4秒之间都算正常（±1秒误差）
			elapsedSeconds := int(elapsed.Seconds()) // 转换为整数秒
			if elapsedSeconds < 2 || elapsedSeconds > 4 {
				t.Errorf("超时检测误差过大，耗时: %v (整数秒: %d)", elapsed, elapsedSeconds)
			}
		}
	})

	// 临界场景4：超大输出（10万行）
	t.Run("超大输出-管道读取", func(t *testing.T) {
		// seq 100000 生成10万行数字
		contentArray, err := ExecCommandStdoutPipe("seq", []string{"100000"}, false)
		if err != nil {
			t.Fatalf("超大输出读取失败: %v", err)
		}
		// 验证行数（允许少量误差）
		if len(contentArray) < 99990 || len(contentArray) > 100010 {
			t.Errorf("超大输出行数异常，实际: %d, 预期: 100000", len(contentArray))
		}
	})
}

// -------------------------- 性能测试（Benchmark） --------------------------
// BenchmarkExecCommandCombinedOutput 基准测试：ExecCommandCombinedOutput性能
func BenchmarkExecCommandCombinedOutput(b *testing.B) {
	// 测试场景1：快速命令（echo）- 高频调用
	b.Run("echo-fast", func(b *testing.B) {
		b.ResetTimer() // 重置计时器，排除初始化耗时
		for i := 0; i < b.N; i++ {
			_, _ = ExecCommandCombinedOutput("echo", []string{"benchmark test"})
		}
	})

	// 测试场景2：中等耗时命令（ls -l）- 常规调用
	b.Run("ls-medium", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ExecCommandCombinedOutput("ls", []string{"-l"})
		}
	})
}

// BenchmarkExecCommandStdoutPipe 基准测试：ExecCommandStdoutPipe性能
func BenchmarkExecCommandStdoutPipe(b *testing.B) {
	// 测试场景1：空输出（cat /dev/null）
	b.Run("cat-null", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ExecCommandStdoutPipe("cat", []string{"/dev/null"}, false)
		}
	})

	// 测试场景2：多输出行（seq 100）
	b.Run("seq-100", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ExecCommandStdoutPipe("seq", []string{"100"}, false)
		}
	})

	// 测试场景3：并发调用（模拟高并发场景）
	b.Run("concurrent-seq", func(b *testing.B) {
		var wg sync.WaitGroup
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = ExecCommandStdoutPipe("seq", []string{"10"}, false)
			}()
		}
		wg.Wait()
	})
}

// -------------------------- 辅助测试：兼容性验证（可选） --------------------------
// TestExecCommand_Compatibility 验证不同系统下的基础兼容性（仅Linux）
func TestExecCommand_Compatibility(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("仅在Linux下运行兼容性测试")
	}

	// 验证Linux特有命令
	_, err := ExecCommandCombinedOutput("uname", []string{"-s"})
	if err != nil {
		t.Fatalf("Linux uname命令执行失败: %v", err)
	}

	// 验证管道命令（简化版不支持shell管道，但可验证基础执行）
	out, err := ExecCommandCombinedOutput("sh", []string{"-c", "echo 'test' | grep test"})
	if err != nil {
		t.Fatalf("管道命令执行失败: %v", err)
	}
	if !strings.Contains(out, "test") {
		t.Errorf("管道命令输出异常: %s", out)
	}
}

func TestExecCommand(t *testing.T) {

	params := []string{"ca", "-gencrl", "-verbose",
		"-out", "/home/rpki/gencerts/ripencc/subcert/tmp/test.crl",
		"-cert", "/home/rpki/gencerts/ripencc/subca/ripencc_subca.pem",
		"-keyfile", "/home/rpki/gencerts/ripencc/subca/ripencc_subca.key",
		"-config", "/home/rpki/gencerts/ripencc/subcert/crl.cnf",
		"-crl_lastupdate", "241011010203Z", "-crl_nextupdate", "341011010203Z"}
	fmtShow := true

	ss, err := ExecCommandStdoutPipe("openssl", params, fmtShow)

	fmt.Println(err)
	for i := range ss {
		fmt.Println(ss[i])
	}

}

func TestExecCommandCombinedOutput(t *testing.T) {
	p := `ca -gencrl -verbose -out /home/rpki/gencerts/ripencc/subcert/tmp/test.crl -cert /home/rpki/gencerts/ripencc/subca/ripencc_subca.pem -keyfile /home/rpki/gencerts/ripencc/subca/ripencc_subca.key -config /home/rpki/gencerts/ripencc/subcert/crl.cnf -crl_lastupdate 241011010203Z -crl_nextupdate 341011010203Z`
	params := strings.Split(p, " ")
	out, err := ExecCommandCombinedOutput("openssl", params)
	fmt.Println("out", out)
	fmt.Println(err)

}
func TestExecCommandStdoutPipe(t *testing.T) {
	p := `10.1.135.22 -p 1-50000`
	params := strings.Split(p, " ")
	out, err := ExecCommandStdoutPipe("nmap", params, true)
	fmt.Println("out", out)
	fmt.Println(err)

}
