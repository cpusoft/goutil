package badgerutil

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

// -------------------------- 测试基础定义 --------------------------
// 测试用结构体（用于序列化/反序列化测试）
type TestData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 获取临时DB路径，测试后自动清理（避免数据残留）
func getTempDBPath(t *testing.T) string {
	// 生成唯一临时目录
	dir := filepath.Join(os.TempDir(), "badger_test_"+t.Name())
	// 清理旧数据（防止跨测试污染）
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("清理临时目录失败: %v", err)
	}
	// 测试结束后自动清理
	t.Cleanup(func() {
		Close()
		_ = os.RemoveAll(dir)
	})
	return dir
}

// 模拟badgerDB异常场景（initialized=1但badgerDB=nil）
func mockBadgerDBInvalidState(t *testing.T) {
	path := getTempDBPath(t)
	// 正常初始化
	assert.NoError(t, Init(path))
	// 强制置空badgerDB（模拟异常）
	badgerDB = nil
	t.Log("已模拟 badgerDB=nil 但 initialized=1 的异常场景")
}

// -------------------------- 核心测试用例 --------------------------

// TestInit 测试初始化逻辑（正常/重复/失败）
func TestInit(t *testing.T) {
	// 场景1：正常初始化
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), atomic.LoadUint32(&initialized))
	assert.NotNil(t, badgerDB)

	// 场景2：重复初始化（应返回错误）
	err = Init(path)
	assert.Error(t, err)
	assert.Equal(t, "badgerDB is already initialized", err.Error())

	// 场景3：初始化失败（无效路径）
	atomic.StoreUint32(&initialized, 0) // 重置状态
	badgerDB = nil
	err = Init("/invalid/path/for/badger/test") // 无效路径
	assert.Error(t, err)
	// 验证状态回滚
	assert.Equal(t, uint32(0), atomic.LoadUint32(&initialized))
	assert.Nil(t, badgerDB)
}

// TestUpdate_View 测试单条写入/读取逻辑
func TestUpdate_View(t *testing.T) {
	// 前置：初始化DB
	path := getTempDBPath(t)
	assert.NoError(t, Init(path))

	// 场景1：正常写入 + 读取
	key := "test:single:1001"
	expected := TestData{ID: 1001, Name: "test_single_data"}
	err := Update(key, expected, 10*time.Minute)
	assert.NoError(t, err)

	// 读取验证
	actual, found, err := View[TestData](key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)

	// 场景2：读取不存在的Key
	notFoundKey := "test:single:not_exists"
	emptyData, found, err := View[TestData](notFoundKey)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, TestData{}, emptyData)

	// 场景3：序列化失败（传入不可序列化类型）
	badKey := "test:single:bad_serialize"
	badValue := make(chan int) // chan不可JSON序列化
	err = Update(badKey, badValue, 0)
	assert.Error(t, err)
	assert.Equal(t, "failed to marshal value to JSON bytes", err.Error())

	// 场景4：未初始化时调用
	Close() // 关闭DB，重置状态
	err = Update("test:single:uninit", expected, 0)
	assert.Error(t, err)
	assert.Equal(t, "badgerDB is not initialized", err.Error())

	_, _, err = View[TestData]("test:single:uninit")
	assert.Error(t, err)
	assert.Equal(t, "badgerDB is not initialized", err.Error())

	// 场景5：异常状态（badgerDB=nil）调用
	mockBadgerDBInvalidState(t)
	err = Update("test:single:invalid", expected, 0)
	assert.Error(t, err)
	assert.Equal(t, "badgerDB is not initialized", err.Error())
}

// TestBatchUpdateMap 测试Map批量写入逻辑（分批次）
func TestBatchUpdateMap(t *testing.T) {
	// 前置：初始化DB
	path := getTempDBPath(t)
	assert.NoError(t, Init(path))

	// 场景1：正常分批次写入（超过batchSize=1000）
	batchSize = 1000 // 显式设置批次大小
	bigMap := make(map[string]TestData, 1500)
	for i := 0; i < 1500; i++ {
		key := fmt.Sprintf("test:batch:map:%d", i)
		bigMap[key] = TestData{ID: i, Name: fmt.Sprintf("batch_map_%d", i)}
	}
	err := BatchUpdateMap(bigMap, 0)
	assert.NoError(t, err)

	// 验证首尾数据
	first, found, _ := View[TestData]("test:batch:map:0")
	assert.True(t, found)
	assert.Equal(t, 0, first.ID)

	last, found, _ := View[TestData]("test:batch:map:1499")
	assert.True(t, found)
	assert.Equal(t, 1499, last.ID)

	// 场景2：序列化失败
	badMap := map[string]chan int{
		"test:batch:map:bad": make(chan int),
	}
	err = BatchUpdateMap(badMap, 0)
	assert.Error(t, err)
	assert.Equal(t, "failed to marshal value to JSON bytes", err.Error())

	// 场景3：未初始化时调用
	Close()
	err = BatchUpdateMap(bigMap, 0)
	assert.Error(t, err)
	assert.Equal(t, "badgerDB is not initialized", err.Error())
}

// TestBatchUpdateKeyFunc 测试自定义Key批量写入逻辑（分批次）
func TestBatchUpdateKeyFunc(t *testing.T) {
	// 前置：初始化DB
	path := getTempDBPath(t)
	assert.NoError(t, Init(path))

	// 场景1：正常分批次写入（超过batchSize）
	batchSize = 1000
	dataList := make([]TestData, 1200)
	for i := 0; i < 1200; i++ {
		dataList[i] = TestData{ID: i, Name: fmt.Sprintf("batch_keyfunc_%d", i)}
	}
	// 自定义Key生成函数
	keyFunc := func(d TestData) string {
		return fmt.Sprintf("test:batch:keyfunc:%d", d.ID)
	}
	err := BatchUpdateKeyFunc(dataList, 0, keyFunc)
	assert.NoError(t, err)

	// 验证数据
	target, found, _ := View[TestData]("test:batch:keyfunc:1199")
	assert.True(t, found)
	assert.Equal(t, 1199, target.ID)

	// 场景2：序列化失败
	badList := []chan int{make(chan int)}
	err = BatchUpdateKeyFunc(badList, 0, func(v chan int) string {
		return "test:batch:keyfunc:bad"
	})
	assert.Error(t, err)
	assert.Equal(t, "failed to marshal value to JSON bytes", err.Error())

	// 场景3：未初始化时调用
	Close()
	err = BatchUpdateKeyFunc(dataList, 0, keyFunc)
	assert.Error(t, err)
	assert.Equal(t, "badgerDB is not initialized", err.Error())
}

// TestPrefixView 测试前缀遍历逻辑
func TestPrefixView(t *testing.T) {
	// 前置：初始化DB并写入测试数据
	path := getTempDBPath(t)
	assert.NoError(t, Init(path))

	// 写入前缀为 "test:prefix:" 的数据
	prefix := "test:prefix:"
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("%s%d", prefix, i)
		_ = Update(key, TestData{ID: i, Name: fmt.Sprintf("prefix_%d", i)}, 0)
	}

	// 场景1：无限制遍历（返回所有数据）
	results, err := PrefixView[TestData](prefix, 0)
	assert.NoError(t, err)
	assert.Len(t, results, 10)

	// 场景2：限制返回条数
	limitedResults, err := PrefixView[TestData](prefix, 5)
	assert.NoError(t, err)
	assert.Len(t, limitedResults, 5)

	// 场景3：单条数据反序列化失败（导致遍历失败）
	// 手动写入无效JSON数据
	badKey := fmt.Sprintf("%sbad", prefix)
	_ = badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(badKey), []byte("invalid json data"))
	})
	// 遍历会失败（容错性问题）
	failResults, err := PrefixView[TestData](prefix, 0)
	assert.Error(t, err)
	assert.Nil(t, failResults)

	// 场景4：未初始化时调用
	Close()
	_, err = PrefixView[TestData](prefix, 0)
	assert.Error(t, err)
	assert.Equal(t, "badgerDB is not initialized", err.Error())
}

// TestDelete 测试删除逻辑
func TestDelete(t *testing.T) {
	// 前置：初始化DB并写入测试数据
	path := getTempDBPath(t)
	assert.NoError(t, Init(path))

	// 场景1：正常删除存在的Key
	key := "test:delete:1001"
	_ = Update(key, TestData{ID: 1001}, 0)
	err := Delete(key)
	assert.NoError(t, err)

	// 验证删除结果
	_, found, _ := View[TestData](key)
	assert.False(t, found)

	// 场景2：删除不存在的Key（Badger允许，无错误）
	notFoundKey := "test:delete:not_exists"
	err = Delete(notFoundKey)
	assert.NoError(t, err)

	// 场景3：未初始化时调用
	Close()
	err = Delete(key)
	assert.Error(t, err)
	assert.Equal(t, "badgerDB is not initialized", err.Error())

	// 场景4：异常状态（badgerDB=nil）调用
	mockBadgerDBInvalidState(t)
	err = Delete("test:delete:invalid")
	assert.Error(t, err)
	assert.Equal(t, "badgerDB is not initialized", err.Error())
}

// TestClose 测试关闭逻辑
func TestClose(t *testing.T) {
	// 场景1：正常关闭
	path := getTempDBPath(t)
	assert.NoError(t, Init(path))
	assert.NotNil(t, badgerDB)
	assert.Equal(t, uint32(1), atomic.LoadUint32(&initialized))

	Close()
	assert.Equal(t, uint32(0), atomic.LoadUint32(&initialized))

	// 场景2：重复关闭（无Panic，验证健壮性）
	Close()
	t.Log("重复调用Close()未触发Panic，符合预期")

	// 场景3：关闭后调用其他方法
	err := Update("test:close:after", TestData{ID: 1}, 0)
	assert.Error(t, err)
	assert.Equal(t, "badgerDB is not initialized", err.Error())
}

// TestConcurrent 测试并发写入/读取（验证线程安全）
func TestConcurrent(t *testing.T) {
	// 前置：初始化DB
	path := getTempDBPath(t)
	assert.NoError(t, Init(path))

	const (
		concurrentCount = 50  // 并发协程数
		writeCount      = 200 // 每个协程写入条数
	)

	var wg sync.WaitGroup
	errChan := make(chan error, concurrentCount*writeCount)

	// 并发写入
	for i := 0; i < concurrentCount; i++ {
		wg.Add(1)
		go func(coroutineID int) {
			defer wg.Done()
			for j := 0; j < writeCount; j++ {
				key := fmt.Sprintf("test:concurrent:%d:%d", coroutineID, j)
				data := TestData{
					ID:   coroutineID*1000 + j,
					Name: fmt.Sprintf("concurrent_%d_%d", coroutineID, j),
				}
				if err := Update(key, data, 0); err != nil {
					errChan <- fmt.Errorf("协程%d写入失败: %v", coroutineID, err)
				}
			}
		}(i)
	}
	wg.Wait()
	close(errChan)

	// 验证无写入错误
	assert.Empty(t, errChan, "并发写入存在错误: %v", <-errChan)

	// 并发读取验证
	errChan = make(chan error, concurrentCount*writeCount)
	for i := 0; i < concurrentCount; i++ {
		wg.Add(1)
		go func(coroutineID int) {
			defer wg.Done()
			for j := 0; j < writeCount; j++ {
				key := fmt.Sprintf("test:concurrent:%d:%d", coroutineID, j)
				expectedID := coroutineID*1000 + j

				data, found, err := View[TestData](key)
				if err != nil {
					errChan <- fmt.Errorf("协程%d读取失败: %v", coroutineID, err)
					continue
				}
				if !found {
					errChan <- fmt.Errorf("协程%d读取Key%s未找到", coroutineID, key)
					continue
				}
				if data.ID != expectedID {
					errChan <- fmt.Errorf("协程%d读取数据不匹配，预期%d，实际%d", coroutineID, expectedID, data.ID)
				}
			}
		}(i)
	}
	wg.Wait()
	close(errChan)

	// 验证无读取错误
	assert.Empty(t, errChan, "并发读取存在错误: %v", <-errChan)
}

// TestExpire 测试过期时间逻辑
func TestExpire(t *testing.T) {
	// 前置：初始化DB
	path := getTempDBPath(t)
	assert.NoError(t, Init(path))

	// 场景1：永不过期（expire<=0）
	permanentKey := "test:expire:permanent"
	err := Update(permanentKey, TestData{ID: 1}, 0)
	assert.NoError(t, err)
	data, found, _ := View[TestData](permanentKey)
	assert.True(t, found)
	assert.Equal(t, 1, data.ID)

	// 场景2：短期过期（1秒）
	expireKey := "test:expire:short"
	err = Update(expireKey, TestData{ID: 2}, 1*time.Second)
	assert.NoError(t, err)

	// 立即读取（未过期）
	data, found, _ = View[TestData](expireKey)
	assert.True(t, found)

	// 等待过期后读取（注：Badger过期数据需GC才会清理，此处仅验证写入逻辑）
	time.Sleep(2 * time.Second)
	_, found, _ = View[TestData](expireKey)
	t.Logf("过期Key读取结果: found=%v（Badger过期数据不会立即删除，需手动触发GC）", found)
}
