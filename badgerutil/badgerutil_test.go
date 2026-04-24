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
// 测试用结构体
type TestData struct {
	ID   int
	Name string
}

// 获取临时DB路径，测试后自动清理
func getTempDBPath(t *testing.T) string {
	dir := filepath.Join(os.TempDir(), "badger_test_"+t.Name())
	_ = os.RemoveAll(dir)
	t.Cleanup(func() {
		Close()
		_ = os.RemoveAll(dir)
	})
	return dir
}

// -------------------------- 基础功能测试 --------------------------

// TestInit 初始化测试
func TestInit(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), atomic.LoadUint32(&initialized))
	assert.NotNil(t, badgerDB)

	// 重复初始化
	err = Init(path)
	assert.ErrorContains(t, err, "already initialized")

	// 初始化失败
	atomic.StoreUint32(&initialized, 0)
	badgerDB = nil
	err = Init("/dev/null/invalid_path")
	assert.Error(t, err)
	assert.Equal(t, uint32(0), atomic.LoadUint32(&initialized))
}

// TestUpdate_View 基础读写
func TestUpdate_View(t *testing.T) {
	path := getTempDBPath(t)
	_ = Init(path)

	key := "test:user:1"
	val := TestData{ID: 1, Name: "alpha"}

	// 写入
	err := Update(key, val, 10*time.Minute)
	assert.NoError(t, err)

	// 读取
	res, found, err := View[TestData](key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, val, res)

	// 不存在
	_, found, _ = View[TestData]("test:not:exist")
	assert.False(t, found)

	// 未初始化
	Close()
	assert.Error(t, Update(key, val, 0))
}

// TestAppend 测试追加功能
func TestAppend(t *testing.T) {
	path := getTempDBPath(t)
	_ = Init(path)

	key := "test:append:list"

	// 第一次 append
	err := Append(key, "log-1", 0)
	assert.NoError(t, err)

	// 第二次 append
	err = Append(key, "log-2", 0)
	assert.NoError(t, err)

	// 读取数组
	list, found, err := View[[]string](key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, []string{"log-1", "log-2"}, list)
}

// TestExists 测试存在判断
func TestExists(t *testing.T) {
	path := getTempDBPath(t)
	_ = Init(path)

	key := "test:exists:1"
	_ = Update(key, TestData{}, 0)

	exists, err := Exists(key)
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = Exists("test:not:exists")
	assert.NoError(t, err)
	assert.False(t, exists)
}

// TestDelete 普通删除
func TestDelete(t *testing.T) {
	path := getTempDBPath(t)
	_ = Init(path)

	key := "test:del:1"
	_ = Update(key, TestData{}, 0)

	err := Delete(key)
	assert.NoError(t, err)

	_, found, _ := View[TestData](key)
	assert.False(t, found)
}

// -------------------------- 高级功能：BatchUpdateKeyFuncs --------------------------

func TestBatchUpdateKeyFuncs(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	defer Close()

	// 测试数据
	datas := []TestData{
		{ID: 101, Name: "jack"},
		{ID: 102, Name: "rose"},
	}

	// mainKey 规则
	mainKeyFunc := func(d TestData) string {
		return fmt.Sprintf("main:%d", d.ID)
	}

	// outerKey 规则（多索引）
	outerKeyFunc := func(d TestData) []string {
		return []string{
			fmt.Sprintf("outer:id:%d", d.ID),
			fmt.Sprintf("outer:name:%s", d.Name),
		}
	}

	// 批量写入
	err = BatchUpdateKeyFuncs(datas, 0, 1, mainKeyFunc, outerKeyFunc)
	assert.NoError(t, err)

	// ✅ 验证 outerKey -> mainKey -> value
	// 通过 outerKey 查询
	v, found, err := View[TestData]("outer:id:101")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, 101, v.ID)

	// 通过另一个 outerKey 查询
	v2, found, _ := View[TestData]("outer:name:jack")
	assert.True(t, found)
	assert.Equal(t, v, v2)

	// 临界值测试
	// 空数据
	err = BatchUpdateKeyFuncs([]TestData{}, 0, 10, mainKeyFunc, outerKeyFunc)
	assert.NoError(t, err)

	// batchSize=0
	err = BatchUpdateKeyFuncs(datas, 0, 0, mainKeyFunc, outerKeyFunc)
	assert.ErrorContains(t, err, "batchSize must be greater")

	// mainKeyFunc=nil
	err = BatchUpdateKeyFuncs(datas, 0, 10, nil, outerKeyFunc)
	assert.ErrorContains(t, err, "mainKeyFunc")

	// outerKeyFunc=nil
	err = BatchUpdateKeyFuncs(datas, 0, 10, mainKeyFunc, nil)
	assert.ErrorContains(t, err, "outerKeyFunc")
}

// -------------------------- DeleteByOuterKey 级联删除测试（核心） --------------------------

func TestDeleteByOuterKey(t *testing.T) {
	path := getTempDBPath(t)
	err := Init(path)
	assert.NoError(t, err)
	defer Close()

	// 1. 构造数据
	data := TestData{ID: 999, Name: "test_delete"}
	mainKeyFunc := func(d TestData) string { return fmt.Sprintf("main:%d", d.ID) }
	outerKeyFunc := func(d TestData) []string {
		return []string{"outer:a", "outer:b", "outer:c"}
	}

	// 2. 批量写入
	err = BatchUpdateKeyFuncs([]TestData{data}, 0, 10, mainKeyFunc, outerKeyFunc)
	assert.NoError(t, err)

	// 3. 删除任意一个 outerKey
	err = DeleteByOuterKey("outer:b")
	assert.NoError(t, err)

	// 4. 验证：所有 outerKey 都被删除
	for _, k := range []string{"outer:a", "outer:b", "outer:c"} {
		exists, _ := Exists(k)
		assert.False(t, exists, k+" 应该被删除")
	}

	// 5. 验证 mainKey value 被删除
	mainValueKey := mainKeyFunc(data) + MAINKEY_TO_VALUE
	exists, _ := Exists(mainValueKey)
	assert.False(t, exists)

	// 6. 验证 mainKey->outerKeys 被删除
	mainOuterKey := mainKeyFunc(data) + MAINKEY_TO_OUTERKEY
	exists, _ = Exists(mainOuterKey)
	assert.False(t, exists)

	t.Log("✅ DeleteByOuterKey 级联删除全部成功")
}

// TestDeleteByOuterKey_CornerCase 临界测试
func TestDeleteByOuterKey_CornerCase(t *testing.T) {
	path := getTempDBPath(t)
	_ = Init(path)

	// 删除不存在的 outerKey
	err := DeleteByOuterKey("outer:not:exist")
	assert.NoError(t, err)

	// 未初始化
	Close()
	err = DeleteByOuterKey("outer:x")
	assert.Error(t, err)
}

// -------------------------- PrefixView --------------------------

func TestPrefixView(t *testing.T) {
	path := getTempDBPath(t)
	_ = Init(path)

	for i := 0; i < 5; i++ {
		_ = Update(fmt.Sprintf("pre:%d", i), TestData{ID: i}, 0)
	}

	list, err := PrefixView[TestData]("pre:", 0)
	assert.NoError(t, err)
	assert.Len(t, list, 5)

	list, err = PrefixView[TestData]("pre:", 2)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

// -------------------------- 并发 & 压力测试 --------------------------

func TestConcurrent_Update_View(t *testing.T) {
	path := getTempDBPath(t)
	_ = Init(path)

	const (
		Goroutines = 20
		Times      = 200
	)
	var wg sync.WaitGroup
	errCh := make(chan error, Goroutines*Times)

	// 并发写
	for i := 0; i < Goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < Times; j++ {
				key := fmt.Sprintf("con:%d:%d", idx, j)
				e := Update(key, TestData{ID: idx*1000 + j}, 0)
				if e != nil {
					errCh <- e
				}
			}
		}(i)
	}
	wg.Wait()
	assert.Empty(t, errCh)

	// 并发读
	for i := 0; i < Goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < Times; j++ {
				key := fmt.Sprintf("con:%d:%d", idx, j)
				val, found, e := View[TestData](key)
				if e != nil || !found || val.ID != idx*1000+j {
					errCh <- fmt.Errorf("read error")
				}
			}
		}(i)
	}
	wg.Wait()
	assert.Empty(t, errCh)
}

// TestBatch_Performance 批量写入压力/性能测试
func TestBatch_Performance(t *testing.T) {
	path := getTempDBPath(t)
	_ = Init(path)

	const N = 10000
	var datas []TestData
	for i := 0; i < N; i++ {
		datas = append(datas, TestData{ID: i, Name: fmt.Sprintf("user%d", i)})
	}

	mainKey := func(d TestData) string { return fmt.Sprintf("m:%d", d.ID) }
	outerKey := func(d TestData) []string {
		return []string{fmt.Sprintf("o:id:%d", d.ID), fmt.Sprintf("o:name:%s", d.Name)}
	}

	start := time.Now()
	err := BatchUpdateKeyFuncs(datas, 0, 1000, mainKey, outerKey)
	cost := time.Since(start)

	assert.NoError(t, err)
	t.Logf("✅ 性能测试：%d 条数据（双索引）批量写入完成，耗时: %v", N, cost)
}

// TestClose 关闭测试
func TestClose(t *testing.T) {
	path := getTempDBPath(t)
	_ = Init(path)
	Close()
	assert.Equal(t, uint32(0), atomic.LoadUint32(&initialized))
	_ = Update("test", 1, 0)
}
