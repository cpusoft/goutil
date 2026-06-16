package badgerutil

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	Close()
	err := Init("memory") //"/root/rpki/data/cache")
	assert.NoError(t, err)
	defer Close()

	keys, err := ViewKeyByPrefix("aki:", 0)
	if err != nil {
		fmt.Println("TestKey(): ViewKeyByPrefix fail",
			err)
		return
	}
	fmt.Println("keys", keys)

	err = badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		fmt.Println("=== 开始遍历 ===")
		for it.Rewind(); it.Valid(); it.Next() {
			k := it.Item().KeyCopy(nil)
			v, err := it.Item().ValueCopy(nil)
			if err != nil {
				fmt.Println("TestKey(): ValueCopy fail", err)
				return err
			}
			kStr := string(k)
			vStr := string(v)
			if strings.HasPrefix(kStr, "aki:") {
				fmt.Printf("KEY: %s\nVALUE: %s\n---\n", kStr, vStr)
			}
			fmt.Printf("count:%d\nKEY: %s\nVALUE: %s\n---\n", count, string(k), string(v))
			count++
		}
		return nil
	})

}
func TestAll(t *testing.T) {
	Close()
	err := Init("memory")
	assert.NoError(t, err)
	defer Close()

	t.Run("CRUD", TestCRUD)
	t.Run("Exists", TestExists)
	t.Run("Expire", TestExpire)

	t.Run("UpdateWithTxn", TestUpdateWithTxn)
	t.Run("DeleteWithTxn", TestDeleteWithTxn)

	t.Run("UpdateWithBatch", TestUpdateWithBatch)
	t.Run("DeleteWithBatch", TestDeleteWithBatch)

	t.Run("PrefixView", TestPrefixView)
	t.Run("NotFoundView", TestNotFoundView)

}

func TestCRUD(t *testing.T) {
	key := "test:crud"
	val := "hello badger"

	err := Update(key, val, 0)
	assert.NoError(t, err)

	res, found, err := View[string](key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, val, res)

	err = Delete(key)
	assert.NoError(t, err)

	_, found, _ = View[string](key)
	assert.False(t, found)
}

func TestExists(t *testing.T) {
	key := "test:exists"
	_ = Update(key, "1", 0)

	exists, err := Exists(key)
	assert.NoError(t, err)
	assert.True(t, exists)

	_ = Delete(key)
	exists, err = Exists(key)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestExpire(t *testing.T) {
	key := "test:expire"
	_ = Update(key, "will expire", 2*time.Second)

	_, found, _ := View[string](key)
	assert.True(t, found)

	time.Sleep(2500 * time.Millisecond)

	_, found, _ = View[string](key)
	assert.False(t, found)
}

func TestUpdateWithTxn(t *testing.T) {
	key := "test:txn:update"
	txn := badgerDB.NewTransaction(true)
	defer txn.Discard()

	expireAt := uint64(time.Now().Add(time.Minute).Unix())
	err := UpdateWithTxn(txn, key, "txn update", expireAt)
	assert.NoError(t, err)

	err = txn.Commit()
	assert.NoError(t, err)

	val, found, _ := View[string](key)
	assert.True(t, found)
	assert.Equal(t, "txn update", val)
}

func TestDeleteWithTxn(t *testing.T) {
	key := "test:txn:del"
	Update(key, "tmp", 0)

	txn := badgerDB.NewTransaction(true)
	defer txn.Discard()
	DeleteWithTxn(txn, key)
	txn.Commit()

	exists, _ := Exists(key)
	assert.False(t, exists)
}

func TestUpdateWithBatch(t *testing.T) {
	batch := badgerDB.NewWriteBatch()
	defer batch.Cancel()

	key := "test:batch:update"
	expireAt := uint64(time.Now().Add(time.Minute).Unix())
	UpdateWithBatch(batch, key, "batch value", expireAt)

	batch.Flush()

	val, found, _ := View[string](key)
	assert.True(t, found)
	assert.Equal(t, "batch value", val)
}

func TestDeleteWithBatch(t *testing.T) {
	key := "test:batch:del"
	Update(key, "to delete", 0)

	batch := badgerDB.NewWriteBatch()
	defer batch.Cancel()
	DeleteWithBatch(batch, key)
	batch.Flush()

	exists, _ := Exists(key)
	assert.False(t, exists)
}

func TestPrefixView(t *testing.T) {
	Update("pre:a", "va", 0)
	Update("pre:b", "vb", 0)
	Update("pre:c", "vc", 0)
	Update("other:x", "vx", 0)

	arr, err := PrefixView[string]("pre:", 0)
	assert.NoError(t, err)
	assert.Len(t, arr, 3)
}

func TestNotFoundView(t *testing.T) {
	_, found, err := View[int]("not:exist")
	assert.NoError(t, err)
	assert.False(t, found)
}
func TestFileMode(t *testing.T) {
	path := "./tmp_badger_test"
	os.RemoveAll(path)

	Init(path)
	defer Close()
	defer os.RemoveAll(path)

	Update("file:test", "ok", 0)
	val, found, _ := View[string]("file:test")
	assert.True(t, found)
	assert.Equal(t, "ok", val)
}

//////////////////////////////////////////////////////////////////////
/*
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
	belogs.Debug("TestBatchUpdateByMultiKeys success")
}

// TestViewByMultiKeys 测试多键查询（通过任意outerKey查value）
func TestViewByMultiKeys(t *testing.T) {
	t.Parallel()
	// 1. 写入测试数据
	datas := genTestData(1)
	_ = BatchUpdateByMultiKeys(datas, 0, 10, getMainKey, getOuterKeys)
	testData := datas[0]
	outerKey := "outer:name:" + testData.Name
	t.Log("outerKey", outerKey)

	// 2. 通过outerKey查询``
	res, err := ViewByMultiKeys[TestModel](outerKey)
	if err != nil || res.ID != testData.ID {
		t.Fatalf("ViewByMultiKeys fail: err=%v, res=%+v", err, res)
	}

	// 3. 查询不存在的outerKey
	_, err = ViewByMultiKeys[TestModel]("outer:not-exist")
	if err == nil {
		t.Fatal("query not exist outerKey should return error")
	}
	belogs.Debug("TestViewByMultiKeys success")
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
	belogs.Debug("TestDeleteByMultiKeys success")
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
	belogs.Debug("TestBatchPerformance success")
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
	belogs.Debug("TestConcurrentPerformance success")
}

// ------------------------------
// 四、BatchUpdateByKey 单主键批量测试
// ------------------------------

// TestBatchUpdateByKey 测试单主键批量写入（临界值全覆盖）
func TestBatchUpdateByKey(t *testing.T) {
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
		{"batchSize大于数据量", genTestData(3), 100, 0, false},
		{"带过期时间", genTestData(3), 5, 1 * time.Second, false},
		{"非法batchSize=0", genTestData(2), 0, 0, true},
		{"非法batchSize负数", genTestData(2), -5, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := BatchUpdateByKey(tt.datas, tt.expire, tt.batchSize, func(data TestModel) string {
				return "batch:single:" + data.ID
			})
			if (err != nil) != tt.wantErr {
				t.Fatalf("BatchUpdateByKey error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	belogs.Debug("TestBatchUpdateByKey success")
}

// TestBatchUpdateByKey_FuncCheck 测试写入后可正常读取
func TestBatchUpdateByKey_FuncCheck(t *testing.T) {
	t.Parallel()
	// 写入测试数据
	datas := genTestData(5)
	batchSize := 2
	keyPrefix := "batch:check:"

	err := BatchUpdateByKey(datas, 0, batchSize, func(data TestModel) string {
		return keyPrefix + data.ID
	})
	if err != nil {
		t.Fatalf("BatchUpdateByKey fail: %v", err)
	}

	// 验证每条数据都能正常 View 读取
	for _, data := range datas {
		key := keyPrefix + data.ID
		res, found, err := View[TestModel](key)
		if err != nil || !found || res.ID != data.ID {
			t.Fatalf("check key=%s fail: err=%v, found=%v, id=%s",
				key, err, found, res.ID)
		}
	}
	belogs.Debug("TestBatchUpdateByKey_FuncCheck success")
}

// TestBatchUpdateByKey_Expire 测试批量过期
func TestBatchUpdateByKey_Expire(t *testing.T) {
	t.Parallel()
	datas := genTestData(2)
	keyPrefix := "batch:expire:"

	// 写入 100ms 过期
	err := BatchUpdateByKey(datas, 100*time.Millisecond, 10, func(data TestModel) string {
		return keyPrefix + data.ID
	})
	if err != nil {
		t.Fatal(err)
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 验证全部过期
	for _, data := range datas {
		key := keyPrefix + data.ID
		_, found, _ := View[TestModel](key)
		if found {
			t.Fatalf("key %s should be expired", key)
		}
	}
	belogs.Debug("TestBatchUpdateByKey_Expire success")
}

// TestBatchUpdateByKey_Performance 单主键批量写入压力测试
func TestBatchUpdateByKey_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}
	t.Parallel()

	const count = 100000
	datas := genTestData(count)
	batchSize := 1000
	keyPrefix := "batch:perf:"

	start := time.Now()
	err := BatchUpdateByKey(datas, 0, batchSize, func(data TestModel) string {
		return keyPrefix + data.ID
	})
	if err != nil {
		t.Fatal(err)
	}

	cost := time.Since(start)
	t.Logf("单主键批量写入 %d 条, 耗时: %v, 平均: %.2f op/s",
		count, cost, float64(count)/cost.Seconds())
	belogs.Debug("TestBatchUpdateByKey_Performance success")
}

// TestBatchUpdateByKey_Concurrent 单主键并发读写测试
func TestBatchUpdateByKey_Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}
	t.Parallel()

	const (
		writeCount       = 1000
		readGoroutine    = 100
		readPerGoroutine = 100
	)
	keyPrefix := "batch:concurrent:"
	datas := genTestData(writeCount)

	// 先批量写入
	err := BatchUpdateByKey(datas, 0, 100, func(data TestModel) string {
		return keyPrefix + data.ID
	})
	if err != nil {
		t.Fatal(err)
	}

	// 并发读取
	start := time.Now()
	var wg sync.WaitGroup
	errCh := make(chan error, readGoroutine)
	wg.Add(readGoroutine)

	for i := 0; i < readGoroutine; i++ {
		go func(idx int) {
			defer wg.Done()
			data := datas[idx%writeCount]
			key := keyPrefix + data.ID
			for j := 0; j < readPerGoroutine; j++ {
				_, found, err := View[TestModel](key)
				if err != nil || !found {
					errCh <- fmt.Errorf("read fail key=%s, err=%v, found=%v", key, err, found)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errCh)
	for e := range errCh {
		t.Fatal(e)
	}

	totalRead := readGoroutine * readPerGoroutine
	cost := time.Since(start)
	t.Logf("单主键并发读取 %d 次, %d 协程, 耗时: %v, 平均: %.2f op/s",
		totalRead, readGoroutine, cost, float64(totalRead)/cost.Seconds())
	belogs.Debug("TestBatchUpdateByKey_Concurrent success")
}

// ------------------------------
// 工具函数
// ------------------------------

// genTestData 生成指定数量的测试数据
func genTestData(n int) []TestModel {
	datas := make([]TestModel, 0, n)
	for i := 0; i < n; i++ {
		datas = append(datas, TestModel{
			ID:        fmt.Sprintf("%d", i),
			Name:      fmt.Sprintf("test-%d", i),
			Timestamp: time.Now().Unix(),
		})
	}
	return datas
}
*/
