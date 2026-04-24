package badgerutil

import (
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/cpusoft/goutil/belogs"
)

// 测试数据结构体（泛型测试用）
type TestModel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
}

// 测试全局临时目录
var testDir string

// TestMain 测试入口：统一初始化/清理环境
func TestMain(m *testing.M) {
	// 创建临时DB目录
	var err error
	testDir, err = os.MkdirTemp("", "badger-test-*")
	if err != nil {
		belogs.Error("TestMain: create temp dir fail", err)
		os.Exit(1)
	}
	defer func() {
		// 测试结束清理临时文件
		_ = os.RemoveAll(testDir)
	}()

	// 初始化DB
	if err = Init(testDir); err != nil {
		belogs.Error("TestMain: Init badger fail", err)
		os.Exit(1)
	}
	defer Close()

	// 运行所有测试用例
	code := m.Run()
	os.Exit(code)
}

// ------------------------------
// 一、基础方法功能测试 (Base Funcs)
// ------------------------------

// TestUpdateView 测试Update/View/Exists/Delete
func TestUpdateView(t *testing.T) {
	t.Parallel()
	key := "test:base:1"
	data := TestModel{ID: "1", Name: "base-test"}

	// 1. 测试写入（永不过期）
	err := Update(key, data, 0)
	if err != nil {
		t.Fatalf("Update fail: %v", err)
	}

	// 2. 测试存在性
	exists, err := Exists(key)
	if err != nil || !exists {
		t.Fatalf("Exists check fail: %v, exists:%v", err, exists)
	}

	// 3. 测试读取
	res, found, err := View[TestModel](key)
	if err != nil || !found || res.ID != data.ID {
		t.Fatalf("View fail: err=%v, found=%v, res=%+v", err, found, res)
	}

	// 4. 测试删除
	err = Delete(key)
	if err != nil {
		t.Fatalf("Delete fail: %v", err)
	}

	// 5. 验证删除后不存在
	res, found, err = View[TestModel](key)
	if err != nil || found {
		t.Fatalf("View after delete fail: found=%v", found)
	}
	belogs.Info("TestUpdateView success")
}

// TestUpdateExpire 测试过期时间临界值
func TestUpdateExpire(t *testing.T) {
	t.Parallel()
	key := "test:base:expire"
	data := TestModel{ID: "expire"}

	// 写入100ms过期
	err := Update(key, data, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// 立即读取：存在
	_, found, _ := View[TestModel](key)
	if !found {
		t.Fatal("immediate read not found")
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 过期后读取：不存在
	_, found, _ = View[TestModel](key)
	if found {
		t.Fatal("expired key still exists")
	}
	belogs.Info("TestUpdateExpire success")
}

// TestAppend 测试数组追加功能
func TestAppend(t *testing.T) {
	t.Parallel()
	key := "test:base:append"

	// 第一次追加（自动创建数组）
	err := Append(key, 1, 0)
	if err != nil {
		t.Fatal(err)
	}

	// 第二次追加
	err = Append(key, 2, 0)
	if err != nil {
		t.Fatal(err)
	}

	// 读取验证
	res, found, err := View[[]int](key)
	if err != nil || !found || len(res) != 2 || res[1] != 2 {
		t.Fatalf("Append fail: res=%+v", res)
	}
	belogs.Info("TestAppend success")
}

// TestPrefixView 测试前缀查询
func TestPrefixView(t *testing.T) {
	t.Parallel()
	prefix := "test:prefix:"
	_ = Update(prefix+"1", 10, 0)
	_ = Update(prefix+"2", 20, 0)

	// 不限制数量
	res, err := PrefixView[int](prefix, 0)
	if err != nil || len(res) != 2 {
		t.Fatalf("PrefixView fail: len=%d, err=%v", len(res), err)
	}

	// 限制数量1
	res, err = PrefixView[int](prefix, 1)
	if err != nil || len(res) != 1 {
		t.Fatalf("PrefixView limit fail: len=%d", len(res))
	}
	belogs.Info("TestPrefixView success")
}

// ------------------------------
// 二、高级多键方法测试 (MultiKeys)
// 🔴 禁止使用任何基础方法 Update/View 等
// ------------------------------

// 生成测试用的mainKey/outerKey函数
func getMainKey(data TestModel) string {
	return "main:" + data.ID
}
func getOuterKeys(data TestModel) []string {
	return []string{
		"outer:name:" + data.Name,
		"outer:id:" + data.ID,
		"outer:time:" + time.Now().Format("20060102"),
	}
}

// TestBatchUpdateByMultiKeys 测试批量多键写入（临界值全覆盖）
func TestBatchUpdateByMultiKeys(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		datas     []TestModel
		batchSize int
		expire    time.Duration
		wantErr   bool
	}{
		{"空数据", []TestModel{}, 10, 0, false},
		{"批次大小1", genTestData(5), 1, 0, false},
		{"正常批次", genTestData(10), 5, 0, false},
		{"带过期时间", genTestData(3), 5, 1 * time.Second, false},
		{"非法批次大小", genTestData(2), 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := BatchUpdateByMultiKeys(tt.datas, tt.expire, tt.batchSize, getMainKey, getOuterKeys)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	belogs.Info("TestBatchUpdateByMultiKeys success")
}

// TestViewByMultiKeys 测试多键查询（通过任意outerKey查value）
func TestViewByMultiKeys(t *testing.T) {
	t.Parallel()
	// 1. 写入测试数据
	datas := genTestData(1)
	_ = BatchUpdateByMultiKeys(datas, 0, 10, getMainKey, getOuterKeys)
	testData := datas[0]
	outerKey := "outer:name:" + testData.Name

	// 2. 通过outerKey查询
	res, err := ViewByMultiKeys[TestModel](outerKey)
	if err != nil || res.ID != testData.ID {
		t.Fatalf("ViewByMultiKeys fail: err=%v, res=%+v", err, res)
	}

	// 3. 查询不存在的outerKey
	_, err = ViewByMultiKeys[TestModel]("outer:not-exist")
	if err == nil {
		t.Fatal("query not exist outerKey should return error")
	}
	belogs.Info("TestViewByMultiKeys success")
}

// TestDeleteByMultiKeys 测试多键删除（删除所有关联数据）
func TestDeleteByMultiKeys(t *testing.T) {
	t.Parallel()
	// 1. 写入数据
	datas := genTestData(1)
	_ = BatchUpdateByMultiKeys(datas, 0, 10, getMainKey, getOuterKeys)
	testData := datas[0]
	outerKey := "outer:id:" + testData.ID

	// 2. 删除
	err := DeleteByMultiKeys(outerKey)
	if err != nil {
		t.Fatal(err)
	}

	// 3. 验证：所有关联key都被删除（通过ViewByMultiKeys验证）
	_, err = ViewByMultiKeys[TestModel](outerKey)
	if err == nil {
		t.Fatal("delete fail, key still exists")
	}
	belogs.Info("TestDeleteByMultiKeys success")
}

// ------------------------------
// 三、性能压力测试 (Performance)
// ------------------------------

// TestBatchPerformance 批量写入压力测试（10万条数据）
func TestBatchPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}
	t.Parallel()

	// 生成10万条测试数据
	const count = 100000
	datas := genTestData(count)
	start := time.Now()

	// 批量写入
	err := BatchUpdateByMultiKeys(datas, 0, 1000, getMainKey, getOuterKeys)
	if err != nil {
		t.Fatal(err)
	}

	cost := time.Since(start)
	t.Logf("批量写入 %d 条数据，耗时: %v, 平均: %.2f op/s",
		count, cost, float64(count)/cost.Seconds())
	belogs.Info("TestBatchPerformance success")
}

// TestConcurrentPerformance 并发读写压力测试
func TestConcurrentPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}
	t.Parallel()

	const (
		writeCount    = 1000
		readGoroutine = 100
		readCount     = 10000
	)
	// 先写入基础数据
	datas := genTestData(writeCount)
	_ = BatchUpdateByMultiKeys(datas, 0, 100, getMainKey, getOuterKeys)
	testOuterKey := "outer:name:test-0"

	// 并发读取
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(readGoroutine)
	errCh := make(chan error, readGoroutine)

	for i := 0; i < readGoroutine; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < readCount/readGoroutine; j++ {
				_, err := ViewByMultiKeys[TestModel](testOuterKey)
				if err != nil {
					errCh <- err
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}

	cost := time.Since(start)
	totalRead := readGoroutine * (readCount / readGoroutine)
	t.Logf("并发读取 %d 次，%d协程，耗时: %v, 平均: %.2f op/s",
		totalRead, readGoroutine, cost, float64(totalRead)/cost.Seconds())
	belogs.Info("TestConcurrentPerformance success")
}

// ------------------------------
// 工具函数
// ------------------------------

// genTestData 生成指定数量的测试数据
func genTestData(n int) []TestModel {
	datas := make([]TestModel, 0, n)
	for i := 0; i < n; i++ {
		str := strconv.Itoa(i)
		datas = append(datas, TestModel{
			ID:        str,
			Name:      "test-" + str,
			Timestamp: time.Now().Unix(),
		})
	}
	return datas
}
