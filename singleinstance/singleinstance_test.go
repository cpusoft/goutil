package singleinstance

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// 测试锁文件路径（Linux 下推荐放在 /tmp 目录）
const testLockFile = "/tmp/my_app_single_instance.lock"

// TestSingleInstance_Normal 正常流程测试：单进程成功加锁，释放锁后可再次加锁
func TestSingleInstance_Normal(t *testing.T) {
	fmt.Println("=== 测试1：正常单进程加锁/解锁 ===")

	// 1. 第一次获取锁：成功
	lock, err := AcquireSingleInstance(testLockFile)
	if err != nil {
		t.Fatalf("首次获取锁失败: %v", err)
	}
	fmt.Println("✅ 进程1：成功获取单实例锁")

	// 2. 同一进程再次获取锁：会失败（文件已被锁定）
	_, err2 := AcquireSingleInstance(testLockFile)
	if err2 == nil {
		t.Fatal("同一进程重复获取锁，预期失败，实际成功")
	}
	fmt.Println("✅ 同一进程重复获取锁：预期失败")

	// 3. 释放锁
	err = lock.Close()
	if err != nil {
		t.Fatalf("释放锁失败: %v", err)
	}
	fmt.Println("✅ 成功释放单实例锁")

	// 4. 释放后再次获取锁：成功
	lock2, err3 := AcquireSingleInstance(testLockFile)
	if err3 != nil {
		t.Fatalf("释放锁后重新获取失败: %v", err3)
	}
	defer lock2.Close()
	fmt.Println("✅ 释放锁后重新获取成功")
	fmt.Println()
}

// TestSingleInstance_MultiProcess 多进程测试：第二个进程无法加锁（核心功能）
func TestSingleInstance_MultiProcess(t *testing.T) {
	fmt.Println("=== 测试2：多进程禁止重复启动 ===")

	// 1. 主进程先获取锁
	lock, err := AcquireSingleInstance(testLockFile)
	if err != nil {
		t.Fatalf("主进程获取锁失败: %v", err)
	}
	defer lock.Close()
	fmt.Println("✅ 主进程：成功获取锁")

	// 2. 启动子进程尝试获取锁（Linux 下执行当前测试二进制文件）
	cmd := exec.Command(os.Args[0], "-test.run=TestSingleInstance_ChildProcess")
	// 把子进程标准输出/错误打印到控制台
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// 运行子进程
	err = cmd.Run()
	if err == nil {
		t.Fatal("子进程预期获取锁失败，实际成功，单实例功能失效！")
	}

	// 断言子进程退出码不为0（代表获取锁失败）
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		fmt.Printf("✅ 子进程获取锁失败，退出码: %d（符合预期）\n", exitErr.ExitCode())
	} else {
		t.Fatalf("子进程异常错误: %v", err)
	}
	fmt.Println()
}

// TestSingleInstance_ChildProcess 子进程测试函数（被多进程测试调用）
func TestSingleInstance_ChildProcess(t *testing.T) {
	// 子进程尝试获取锁，预期失败
	_, err := AcquireSingleInstance(testLockFile)
	if err != nil {
		fmt.Printf("❌ 子进程：获取锁失败，原因: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// TestSingleInstance_AutoRelease 测试进程异常退出：锁自动释放（Linux flock 特性）
func TestSingleInstance_AutoRelease(t *testing.T) {
	fmt.Println("=== 测试3：进程异常退出，锁自动释放 ===")

	// 1. 启动一个子进程，获取锁后直接崩溃（不释放锁）
	cmd := exec.Command(os.Args[0], "-test.run=TestSingleInstance_CrashProcess")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()

	// 等待内核释放文件锁（flock 进程退出自动释放）
	time.Sleep(100 * time.Millisecond)

	// 2. 主进程重新获取锁：成功
	lock, err := AcquireSingleInstance(testLockFile)
	if err != nil {
		t.Fatalf("异常进程退出后，获取锁失败: %v", err)
	}
	defer lock.Close()
	fmt.Println("✅ 进程异常退出，锁已自动释放，重新获取成功")
	fmt.Println()
}

// TestSingleInstance_CrashProcess 崩溃进程：获取锁后直接退出（不手动释放）
func TestSingleInstance_CrashProcess(t *testing.T) {
	lock, err := AcquireSingleInstance(testLockFile)
	if err != nil {
		fmt.Printf("崩溃进程获取锁失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("⚠️ 崩溃进程：已获取锁，即将异常退出（不释放锁）", lock)
	// 直接退出，不执行 lock.Close()
	os.Exit(0)
}

// TestSingleInstance_Clean 测试完成后清理锁文件
func TestSingleInstance_Clean(t *testing.T) {
	fmt.Println("=== 测试完成：清理锁文件 ===")
	_ = os.Remove(testLockFile)
	fmt.Println("✅ 锁文件已清理")
}

// TestMain 测试入口
func TestMain(m *testing.M) {
	fmt.Println("======== Linux 单实例锁 测试开始 ========")
	// 执行所有测试用例
	code := m.Run()
	fmt.Println("======== Linux 单实例锁 测试结束 ========")
	os.Exit(code)
}
