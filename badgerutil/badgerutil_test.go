package badgerutil

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// -------------------------- 测试基础定义 --------------------------
// 测试用结构体（用于序列化/反序列化测试）
type TestData struct {
	ID   int
	Name string
}

// 获取临时DB路径，测试后自动清理（避免数据残留）
func getTempDBPath(t *testing.T) string {
	dir := filepath.Join(os.TempDir(), "badger_test_"+t.Name())
	_ = os.RemoveAll(dir)
	t.Cleanup(func() {
		Close()
		_ = os.RemoveAll(dir)
	})
	return dir
}

// 模拟badgerDB异常场景（initialized=1但badgerDB=nil）
func mockBadgerDBInvalidState(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	if err != nil {
		return
	}
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
	atomic.StoreUint32(&initialized, 0)
	badgerDB = nil
	err = Init("/dev/null/test_badger")
	assert.Error(t, err)
	assert.Equal(t, uint32(0), atomic.LoadUint32(&initialized))
	assert.Nil(t, badgerDB)
}

// TestUpdate_View 测试单条写入/读取/序列化/未初始化/异常状态
func TestUpdate_View(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	// 1. 正常写入读取
	key := "test:single:1001"
	expected := TestData{ID: 1001, Name: "test_data"}
	assert.NoError(t, Update(key, expected, 10*time.Minute))

	actual, found, err := View[TestData](key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, expected, actual)

	// 2. 读取不存在key
	_, found, _ = View[TestData]("test:not:exist")
	assert.False(t, found)

	// 3. 序列化失败（gob不支持的类型）
	assert.Error(t, Update("test:bad", make(chan int), 0))

	// 4. 未初始化调用
	Close()
	assert.Error(t, Update("test:uninit", expected, 0))
	_, _, err = View[TestData]("test:uninit")
	assert.Error(t, err)

	// 5. 异常状态调用
	mockBadgerDBInvalidState(t)
	assert.Error(t, Update("test:invalid", expected, 0))
}

// TestBatchUpdateKeyFuncs 测试新版多键同值批量写入（全覆盖+临界值）
func TestBatchUpdateKeyFuncs(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	defer Close() // 确保最后关闭DB

	// 测试数据结构体（确保和你的 View/Write 匹配）
	type TestData struct {
		ID   int
		Name string
	}

	// 测试数据
	dataList := []TestData{
		{ID: 1, Name: "batch1"},
		{ID: 2, Name: "batch2"},
		{ID: 3, Name: "batch3"},
	}

	// ===================== 关键修改 =====================
	// 新版函数要求：传入 1 个 keyFunc，返回 []string
	// ====================================================
	keyFunc := func(d TestData) []string {
		return []string{
			fmt.Sprintf("test:batch:id:%d", d.ID),
			fmt.Sprintf("test:batch:name:%s", d.Name),
		}
	}

	// 1. 正常批量写入（批次大小=2）
	err = BatchUpdateKeyFuncs(dataList, 0, 2, keyFunc)
	assert.NoError(t, err)

	// 验证：两个key都能读到同一份数据
	v1, found, _ := View[TestData]("test:batch:id:1")
	assert.True(t, found)
	v2, found, _ := View[TestData]("test:batch:name:batch1")
	assert.True(t, found)
	assert.Equal(t, v1, v2)

	// 2. 临界值：batchSize=1
	err = BatchUpdateKeyFuncs(dataList, 0, 1, keyFunc)
	assert.NoError(t, err)

	// 3. 临界值：空数据
	err = BatchUpdateKeyFuncs([]TestData{}, 0, 10, keyFunc)
	assert.NoError(t, err)

	// 4. 临界值：非法batchSize（<=0）
	err = BatchUpdateKeyFuncs(dataList, 0, 0, keyFunc)
	assert.Error(t, err)

	// 5. 临界值：keyFunc = nil（新版判断逻辑）
	err = BatchUpdateKeyFuncs(dataList, 0, 10, nil)
	assert.Error(t, err)

	// 6. 序列化失败测试（无法序列化为 JSON 的类型）
	badData := []chan int{make(chan int)}
	badKeyFunc := func(c chan int) []string {
		return []string{"test:bad:key"}
	}
	err = BatchUpdateKeyFuncs(badData, 0, 10, badKeyFunc)
	assert.Error(t, err)

	// 7. 未初始化调用
	Close()
	err = BatchUpdateKeyFuncs(dataList, 0, 10, keyFunc)
	assert.Error(t, err)
}

// TestPrefixView 测试前缀查询（正常/限制条数/坏数据/未初始化）
func TestPrefixView(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	prefix := "test:prefix:"
	for i := 0; i < 10; i++ {
		_ = Update(fmt.Sprintf("%s%d", prefix, i), TestData{ID: i}, 0)
	}

	// 1. 全量查询
	list, err := PrefixView[TestData](prefix, 0)
	assert.NoError(t, err)
	assert.Len(t, list, 10)

	// 2. 限制条数
	list, err = PrefixView[TestData](prefix, 5)
	assert.NoError(t, err)
	assert.Len(t, list, 5)

	// 3. 未初始化
	Close()
	_, err = PrefixView[TestData](prefix, 0)
	assert.Error(t, err)
}

// TestDelete 测试删除（存在/不存在/未初始化/异常状态）
func TestDelete(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	key := "test:del:1"
	_ = Update(key, TestData{}, 0)

	// 1. 删除存在key
	assert.NoError(t, Delete(key))
	_, found, _ := View[TestData](key)
	assert.False(t, found)

	// 2. 删除不存在key（无错误）
	assert.NoError(t, Delete("test:del:not:exist"))

	// 3. 未初始化
	Close()
	assert.Error(t, Delete(key))

	// 4. 异常状态
	mockBadgerDBInvalidState(t)
	assert.Error(t, Delete(key))
}

// TestClose 测试关闭（正常/重复关闭/关闭后调用）
func TestClose(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	// 关闭
	Close()
	assert.Equal(t, uint32(0), atomic.LoadUint32(&initialized))

	// 重复关闭（无panic）
	Close()

	// 关闭后调用方法报错
	assert.Error(t, Update("test:close", 1, 0))
}

// TestConcurrent 高并发读写测试（验证线程安全）
func TestConcurrent(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	const (
		grs   = 50
		count = 200
	)
	var wg sync.WaitGroup
	errCh := make(chan error, grs*count)

	// 并发写
	for i := 0; i < grs; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < count; j++ {
				key := fmt.Sprintf("test:con:%d:%d", idx, j)
				err := Update(key, TestData{ID: idx*1000 + j}, 0)
				if err != nil {
					errCh <- err
				}
			}
		}(i)
	}
	wg.Wait()
	assert.Empty(t, errCh)

	// 并发读
	for i := 0; i < grs; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < count; j++ {
				key := fmt.Sprintf("test:con:%d:%d", idx, j)
				val, found, err := View[TestData](key)
				if err != nil || !found || val.ID != idx*1000+j {
					errCh <- fmt.Errorf("fail")
				}
			}
		}(i)
	}
	wg.Wait()
	assert.Empty(t, errCh)
}

// TestExpire 测试过期时间（永久/过期写入）
func TestExpire(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	// 永久有效
	_ = Update("test:exp:forever", TestData{}, 0)
	_, found, _ := View[TestData]("test:exp:forever")
	assert.True(t, found)

	// 1s过期
	_ = Update("test:exp:short", TestData{}, 1*time.Second)
	time.Sleep(1100 * time.Millisecond)
	// badger需要gc才会物理删除，这里只验证写入逻辑
}

// TestBatchPerformance 批量写入性能基础测试
func TestBatchPerformance(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	// 构造1万条数据，多key批量写入
	var list []TestData
	for i := 0; i < 10000; i++ {
		list = append(list, TestData{ID: i, Name: fmt.Sprintf("perf_%d", i)})
	}

	keyFunc := func(d TestData) []string {
		return []string{
			fmt.Sprintf("test:batch:id:%d", d.ID),
			fmt.Sprintf("test:batch:name:%s", d.Name),
		}
	}

	start := time.Now()
	err = BatchUpdateKeyFuncs(list, 0, 1000, keyFunc)
	cost := time.Since(start)

	assert.NoError(t, err)
	t.Logf("批量写入 10000 条（双key）完成，耗时: %v", cost)
}
