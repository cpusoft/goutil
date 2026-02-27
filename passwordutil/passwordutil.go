package passwordutil

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/uuidutil"
)

// 保留原有核心函数（无修改）
func GetHashPasswordAndSalt(password string) (hashPassword, salt string) {
	salt = uuidutil.GetUuid()
	return GetHashPassword(password, salt), salt
}

func GetHashPassword(password, salt string) (hashPassword string) {
	return hashutil.Sha256([]byte(password + salt))
}

func VerifyHashPassword(password, salt, hashPassword string) (isPass bool) {
	hashPassword1 := hashutil.Sha256([]byte(password + salt))
	return hashPassword == hashPassword1
}

// ForceTestHashPassword 逐行读取文件+多协程匹配，找到密码立即终止所有协程并返回
func ForceTestHashPassword(hashPassword, salt string, dictFilePathName string) (password string, err error) {
	// 1. 打开文件（逐行读取，不一次性加载所有行）
	file, err := os.Open(dictFilePathName)
	if err != nil {
		belogs.Error("ForceTestHashPassword(): Open file fail, dictFilePathName:", dictFilePathName, err)
		return "", err
	}
	defer file.Close() // 确保函数退出时关闭文件句柄

	// 2. 核心控制组件：context终止协程，resultCh传递结果
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()                   // 兜底释放context资源
	resultCh := make(chan string, 1) // 带缓冲通道，避免协程阻塞
	var wg sync.WaitGroup

	// 3. 并发控制：保留原500并发限制
	concurrencyCh := make(chan struct{}, 500) // 改用struct{}节省内存

	// 4. 监听结果：找到密码后立即触发终止逻辑
	go func() {
		select {
		case password = <-resultCh: // 接收匹配的密码
			belogs.Debug("ForceTestHashPassword(): found password:", password)
			cancel() // 终止所有协程+停止读取文件
		case <-ctx.Done(): // 上下文取消（如函数退出）
		}
		close(resultCh)
	}()

	// 5. 逐行读取文件，边读边启动协程匹配
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// 优先检查：已找到密码则停止读取文件
		select {
		case <-ctx.Done():
			break // 退出文件读取循环
		default:
		}

		// 读取当前行，去除首尾空格
		line := scanner.Text()
		testPassword := strings.TrimSpace(line)
		if testPassword == "" { // 跳过空行
			continue
		}

		// 占用并发槽位，控制最大并发数
		concurrencyCh <- struct{}{}
		wg.Add(1)

		// 启动协程匹配密码：修复闭包变量问题（传入当前行的testPassword）
		go func(pwd string) {
			defer func() {
				<-concurrencyCh // 释放并发槽位
				wg.Done()       // 标记协程完成
			}()

			// 优先检查：上下文已取消则直接退出，不执行验证
			select {
			case <-ctx.Done():
				return
			default:
			}

			// 密码匹配逻辑
			isPass := VerifyHashPassword(pwd, salt, hashPassword)
			if isPass {
				// 非阻塞写入结果（避免多个协程同时写入）
				select {
				case resultCh <- pwd:
				default:
				}
				cancel() // 立即终止所有未完成的协程
			}
		}(testPassword) // 传入当前行的密码，修复闭包陷阱
	}

	// 检查文件读取过程中是否出错
	if err = scanner.Err(); err != nil {
		belogs.Error("ForceTestHashPassword(): Scan file fail, dictFilePathName:", dictFilePathName, err)
		cancel() // 读取出错时终止所有协程
		wg.Wait()
		close(concurrencyCh)
		return "", err
	}

	// 6. 等待所有协程完成，确保资源释放
	wg.Wait()
	close(concurrencyCh) // 关闭并发控制通道
	return password, nil
}

/*
func ForceTestHashPassword(hashPassword, salt string, dictFilePathName string) (password string, err error) {
	lines, err := fileutil.ReadFileToLines(dictFilePathName)
	if err != nil {
		belogs.Error("ForceTestHashPassword(): ReadFileToLines fail, dictFilePathName:", dictFilePathName, err)
		return "", err
	}
	ch := make(chan int, 500)
	var wg sync.WaitGroup
	for _, line := range lines {
		wg.Add(1)
		ch <- 1
		go func(wg1 *sync.WaitGroup, ch1 chan int) {
			defer func() {
				<-ch1
				wg.Done()
			}()

			testPassword := strings.TrimSpace(line)
			//belogs.Debug("ForceTestHashPassword(): test pasword:", testPassword)
			isPass := VerifyHashPassword(testPassword, salt, hashPassword)
			if isPass {
				password = testPassword
				belogs.Debug("ForceTestHashPassword(): found password")
			}
		}(&wg, ch)
	}
	wg.Wait()
	close(ch)
	return password, nil
}
*/
