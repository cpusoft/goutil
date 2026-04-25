package badgerutil

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	// 测试内存模式（速度最快）
	err := Init("memory")
	assert.NoError(t, err)
	defer Close()

	// 基础 CRUD
	t.Run("CRUD", TestCRUD)
	t.Run("Exists", TestExists)
	t.Run("Append", TestAppend)
	t.Run("Expire", TestExpire)

	// 事务
	t.Run("UpdateWithTxn", TestUpdateWithTxn)
	t.Run("AppendWithTxn", TestAppendWithTxn)
	t.Run("DeleteWithTxn", TestDeleteWithTxn)

	// 批量 WriteBatch
	t.Run("UpdateWithBatch", TestUpdateWithBatch)
	t.Run("AppendWithBatch", TestAppendWithBatch)
	t.Run("DeleteWithBatch", TestDeleteWithBatch)

	// 前缀查询
	t.Run("PrefixView", TestPrefixView)

	// 边界 & 异常
	t.Run("EmptyKey", TestEmptyKey)
	t.Run("NotFoundView", TestNotFoundView)

	// 组合 & 压力
	t.Run("BatchMixed", TestBatchMixed)
	t.Run("StressTest", TestStressTest)
}

// TestCRUD 基础增删改查
func TestCRUD(t *testing.T) {
	key := "test:crud"
	val := "hello badger"

	// Update
	err := Update(key, val, 0)
	assert.NoError(t, err)

	// View
	res, found, err := View[string](key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, val, res)

	// Delete
	err = Delete(key)
	assert.NoError(t, err)

	// Check deleted
	_, found, _ = View[string](key)
	assert.False(t, found)
}

// TestExists 判断 key 是否存在
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

// TestAppend 测试列表追加
func TestAppend(t *testing.T) {
	key := "test:append"

	// 第一次 append
	err := Append(key, 1, 0)
	assert.NoError(t, err)

	// 第二次
	err = Append(key, 2, 0)
	assert.NoError(t, err)

	// 查看
	arr, found, err := View[[]int](key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, []int{1, 2}, arr)
}

// TestExpire 测试过期
func TestExpire(t *testing.T) {
	key := "test:expire"
	_ = Update(key, "will expire", 100*time.Millisecond)

	// 存在
	_, found, _ := View[string](key)
	assert.True(t, found)

	time.Sleep(120 * time.Millisecond)

	// 已过期
	_, found, _ = View[string](key)
	assert.False(t, found)
}

// TestUpdateWithTxn 事务更新
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

// TestAppendWithTxn 事务追加
func TestAppendWithTxn(t *testing.T) {
	key := "test:txn:append"
	txn := badgerDB.NewTransaction(true)
	defer txn.Discard()

	expireAt := uint64(time.Now().Add(time.Minute).Unix())
	_ = AppendWithTxn(txn, key, "a", expireAt)
	_ = AppendWithTxn(txn, key, "b", expireAt)

	err := txn.Commit()
	assert.NoError(t, err)

	arr, found, _ := View[[]string](key)
	assert.True(t, found)
	assert.Equal(t, []string{"a", "b"}, arr)
}

// TestDeleteWithTxn 事务删除
func TestDeleteWithTxn(t *testing.T) {
	key := "test:txn:del"
	_ = Update(key, "tmp", 0)

	txn := badgerDB.NewTransaction(true)
	defer txn.Discard()
	err := DeleteWithTxn(txn, key)
	assert.NoError(t, err)
	txn.Commit()

	exists, _ := Exists(key)
	assert.False(t, exists)
}

// TestUpdateWithBatch 批量 Update
func TestUpdateWithBatch(t *testing.T) {
	batch := badgerDB.NewWriteBatch()
	defer batch.Cancel()

	key := "test:batch:update"
	expireAt := uint64(time.Now().Add(time.Minute).Unix())
	err := UpdateWithBatch(batch, key, "batch value", expireAt)
	assert.NoError(t, err)

	err = batch.Flush()
	assert.NoError(t, err)

	val, found, _ := View[string](key)
	assert.True(t, found)
	assert.Equal(t, "batch value", val)
}

// TestAppendWithBatch 批量 Append
func TestAppendWithBatch(t *testing.T) {
	key := "test:batch:append"
	batch := badgerDB.NewWriteBatch()
	defer batch.Cancel()

	txn := badgerDB.NewTransaction(false)
	defer txn.Discard()

	expireAt := uint64(time.Now().Add(time.Minute).Unix())
	err := AppendWithBatch(txn, batch, key, "x", expireAt)
	assert.NoError(t, err)
	err = AppendWithBatch(txn, batch, key, "y", expireAt)
	assert.NoError(t, err)

	err = batch.Flush()
	assert.NoError(t, err)

	arr, found, _ := View[[]string](key)
	assert.True(t, found)
	assert.Equal(t, []string{"x", "y"}, arr)
}

// TestDeleteWithBatch 批量删除
func TestDeleteWithBatch(t *testing.T) {
	key := "test:batch:del"
	_ = Update(key, "to delete", 0)

	batch := badgerDB.NewWriteBatch()
	defer batch.Cancel()
	err := DeleteWithBatch(batch, key)
	assert.NoError(t, err)
	batch.Flush()

	exists, _ := Exists(key)
	assert.False(t, exists)
}

// TestPrefixView 前缀查询
func TestPrefixView(t *testing.T) {
	_ = Update("pre:a", "va", 0)
	_ = Update("pre:b", "vb", 0)
	_ = Update("pre:c", "vc", 0)
	_ = Update("other:x", "vx", 0)

	arr, err := PrefixView[string]("pre:", 0)
	assert.NoError(t, err)
	assert.Len(t, arr, 3)
}

// TestNotFoundView 查询不存在 key
func TestNotFoundView(t *testing.T) {
	_, found, err := View[int]("not:exist")
	assert.NoError(t, err)
	assert.False(t, found)
}

// TestEmptyKey 空 key 测试
func TestEmptyKey(t *testing.T) {
	err := Update("", "empty key test", 0)
	assert.NoError(t, err)

	val, found, _ := View[string]("")
	assert.True(t, found)
	assert.Equal(t, "empty key test", val)
}

// TestBatchMixed 组合批量：update + append + delete
func TestBatchMixed(t *testing.T) {
	batch := badgerDB.NewWriteBatch()
	defer batch.Cancel()

	readTxn := badgerDB.NewTransaction(false)
	defer readTxn.Discard()

	expireAt := uint64(time.Now().Add(time.Minute).Unix())

	_ = UpdateWithBatch(batch, "mix:u", "update", expireAt)
	_ = AppendWithBatch(readTxn, batch, "mix:a", "item1", expireAt)
	_ = UpdateWithBatch(batch, "mix:d", "delete me", expireAt)
	_ = DeleteWithBatch(batch, "mix:d")

	err := batch.Flush()
	assert.NoError(t, err)

	u, _, _ := View[string]("mix:u")
	assert.Equal(t, "update", u)

	a, _, _ := View[[]string]("mix:a")
	assert.Equal(t, []string{"item1"}, a)

	exists, _ := Exists("mix:d")
	assert.False(t, exists)
}

// TestStressTest 高并发压力测试
func TestStressTest(t *testing.T) {
	const n = 5000
	done := make(chan bool)

	// 并发写入
	for i := 0; i < n; i++ {
		go func(i int) {
			key := "stress:k"
			_ = Append(key, i, 0)
			done <- true
		}(i)
	}

	// 等待完成
	for i := 0; i < n; i++ {
		<-done
	}

	arr, _, _ := View[[]int]("stress:k")
	t.Log("total items in list:", len(arr))
	assert.True(t, len(arr) > 0)
}

// TestFileMode 测试文件 DB（可选）
func TestFileMode(t *testing.T) {
	path := "./tmp_badger_test"
	_ = os.RemoveAll(path)

	err := Init(path)
	assert.NoError(t, err)
	defer Close()
	defer os.RemoveAll(path)

	_ = Update("file:test", "file mode ok", 0)
	val, found, _ := View[string]("file:test")
	assert.True(t, found)
	assert.Equal(t, "file mode ok", val)
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
	belogs.Info("TestBatchUpdateByKey success")
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
	belogs.Info("TestBatchUpdateByKey_FuncCheck success")
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
	belogs.Info("TestBatchUpdateByKey_Expire success")
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
	belogs.Info("TestBatchUpdateByKey_Performance success")
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
	belogs.Info("TestBatchUpdateByKey_Concurrent success")
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
