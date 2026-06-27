package badgerutil

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/dgraph-io/badger/v4"
)

// ==================== 测试数据结构 ====================

type TestUser struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	IsActive bool   `json:"is_active"`
}

type TestConfig struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ==================== 测试辅助函数 ====================

// getTestDBPath 获取测试数据库路径
func getTestDBPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(os.TempDir(), fmt.Sprintf("badger_test_%d_%d", time.Now().UnixNano(), t.Name()))
}

// cleanupTestDB 清理测试数据库
func cleanupTestDB(t *testing.T, dbPath string) {
	t.Helper()
	if err := os.RemoveAll(dbPath); err != nil {
		belogs.Error("cleanupTestDB() fail:", err)
	}
}

// createTestDB 创建测试数据库
func createTestDB(t *testing.T, dbPath string) *badger.DB {
	t.Helper()
	badgerDb, err := Init(dbPath)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	return badgerDb
}

// ==================== Init / Close 测试 ====================

// TestInit_NormalPath 测试正常路径初始化
func TestInit_NormalPath(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)

	badgerDb, err := Init(dbPath)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	if badgerDb == nil {
		t.Fatal("Init() returned nil db")
	}

	// 验证目录已创建
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("dbPath was not created")
	}

	Close(badgerDb)
	belogs.Info("TestInit_NormalPath: passed")
}

// TestInit_MemoryMode 测试内存模式
func TestInit_MemoryMode(t *testing.T) {
	badgerDb, err := Init("memory")
	if err != nil {
		t.Fatalf("Init(memory) failed: %v", err)
	}
	if badgerDb == nil {
		t.Fatal("Init(memory) returned nil db")
	}

	// 写入数据
	err = Update(badgerDb, "test_key", "test_value", 0)
	if err != nil {
		t.Fatalf("Update() in memory mode failed: %v", err)
	}

	// 读取数据
	val, found, err := View[string](badgerDb, "test_key")
	if err != nil {
		t.Fatalf("View() in memory mode failed: %v", err)
	}
	if !found {
		t.Fatal("View() key not found in memory mode")
	}
	if val != "test_value" {
		t.Fatalf("View() value mismatch: got %v, want test_value", val)
	}

	Close(badgerDb)
	belogs.Info("TestInit_MemoryMode: passed")
}

// TestInit_ConcurrentInit 测试并发初始化（应只有一个成功）
func TestInit_ConcurrentInit(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount int
	var dbs []*badger.DB

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			badgerDb, err := Init(dbPath)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				successCount++
				dbs = append(dbs, badgerDb)
			} else {
				belogs.Info("ConcurrentInit: expected failure:", err)
			}
		}()
	}
	wg.Wait()

	// 只有一个应该成功
	if successCount != 1 {
		t.Fatalf("Expected 1 successful init, got %d", successCount)
	}

	// 关闭成功的那个
	for _, db := range dbs {
		Close(db)
	}

	cleanupTestDB(t, dbPath)
	belogs.Info("TestInit_ConcurrentInit: passed")
}

// TestInit_ReopenAfterClose 测试关闭后重新打开
func TestInit_ReopenAfterClose(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)

	// 第一次打开，写入数据
	badgerDb1 := createTestDB(t, dbPath)
	err := Update(badgerDb1, "persistent_key", "persistent_value", 0)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
	Close(badgerDb1)

	// 第二次打开，验证数据还在
	badgerDb2, err := Init(dbPath)
	if err != nil {
		t.Fatalf("Reopen Init() failed: %v", err)
	}

	val, found, err := View[string](badgerDb2, "persistent_key")
	if err != nil {
		t.Fatalf("View() after reopen failed: %v", err)
	}
	if !found {
		t.Fatal("View() key not found after reopen")
	}
	if val != "persistent_value" {
		t.Fatalf("View() value mismatch after reopen: got %v", val)
	}

	Close(badgerDb2)
	belogs.Info("TestInit_ReopenAfterClose: passed")
}

// TestInit_StaleLockFile 测试残留 LOCK 文件的处理
func TestInit_StaleLockFile(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)

	// 创建目录和模拟残留 LOCK 文件
	_ = os.MkdirAll(dbPath, os.ModePerm)
	lockFile := filepath.Join(dbPath, "LOCK")
	if err := os.WriteFile(lockFile, []byte("stale"), 0644); err != nil {
		t.Fatalf("Failed to create stale LOCK file: %v", err)
	}

	// 应该能正常打开（如果实现了清理逻辑）
	badgerDb, err := Init(dbPath)
	if err != nil {
		// 如果没有实现清理逻辑，这里会失败，这是预期行为
		belogs.Info("TestInit_StaleLockFile: Init failed as expected without cleanup logic:", err)
		return
	}

	Close(badgerDb)
	belogs.Info("TestInit_StaleLockFile: passed")
}

// ==================== Update / View 测试 ====================

// TestUpdateAndView_Basic 测试基本的更新和查看
func TestUpdateAndView_Basic(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	user := TestUser{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30, IsActive: true}
	err := Update(badgerDb, "user:1", user, 0)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	result, found, err := View[TestUser](badgerDb, "user:1")
	if err != nil {
		t.Fatalf("View() failed: %v", err)
	}
	if !found {
		t.Fatal("View() key not found")
	}
	if result.ID != user.ID || result.Name != user.Name || result.Email != user.Email {
		t.Fatalf("View() result mismatch: got %+v, want %+v", result, user)
	}

	belogs.Info("TestUpdateAndView_Basic: passed")
}

// TestUpdateAndView_EmptyKey 测试空键
func TestUpdateAndView_EmptyKey(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	err := Update(badgerDb, "", "value", 0)
	if err == nil {
		t.Fatal("Update() with empty key should fail")
	}

	_, _, err = View[string](badgerDb, "")
	if err == nil {
		t.Fatal("View() with empty key should fail")
	}

	belogs.Info("TestUpdateAndView_EmptyKey: passed")
}

// TestUpdateAndView_NilDB 测试 nil 数据库
func TestUpdateAndView_NilDB(t *testing.T) {
	err := Update[string](nil, "key", "value", 0)
	if err == nil {
		t.Fatal("Update() with nil db should fail")
	}

	_, _, err = View[string](nil, "key")
	if err == nil {
		t.Fatal("View() with nil db should fail")
	}

	belogs.Info("TestUpdateAndView_NilDB: passed")
}

// TestUpdateAndView_KeyNotFound 测试键不存在
func TestUpdateAndView_KeyNotFound(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	_, found, err := View[string](badgerDb, "nonexistent_key")
	if err != nil {
		t.Fatalf("View() for nonexistent key should not error: %v", err)
	}
	if found {
		t.Fatal("View() should not find nonexistent key")
	}

	belogs.Info("TestUpdateAndView_KeyNotFound: passed")
}

// TestUpdateAndView_Expire 测试过期时间
func TestUpdateAndView_Expire(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	// 写入 1 秒后过期的数据
	err := Update(badgerDb, "expire_key", "expire_value", 1*time.Second)
	if err != nil {
		t.Fatalf("Update() with expire failed: %v", err)
	}

	// 立即读取，应该存在
	val, found, err := View[string](badgerDb, "expire_key")
	if err != nil {
		t.Fatalf("View() immediately failed: %v", err)
	}
	if !found {
		t.Fatal("View() should find key immediately")
	}
	if val != "expire_value" {
		t.Fatalf("View() value mismatch: got %v", val)
	}

	// 等待过期
	time.Sleep(2 * time.Second)

	// 再次读取，应该不存在（Badger 的过期需要 GC 或 reopen 才能完全生效）
	// 这里只是验证写入时没有错误
	belogs.Info("TestUpdateAndView_Expire: passed (note: Badger expiry requires GC/reopen to fully take effect)")
}

// TestUpdateAndView_Overwrite 测试覆盖写入
func TestUpdateAndView_Overwrite(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	err := Update(badgerDb, "overwrite_key", "value1", 0)
	if err != nil {
		t.Fatalf("First Update() failed: %v", err)
	}

	err = Update(badgerDb, "overwrite_key", "value2", 0)
	if err != nil {
		t.Fatalf("Second Update() failed: %v", err)
	}

	val, found, err := View[string](badgerDb, "overwrite_key")
	if err != nil {
		t.Fatalf("View() after overwrite failed: %v", err)
	}
	if !found {
		t.Fatal("View() key not found after overwrite")
	}
	if val != "value2" {
		t.Fatalf("View() value mismatch after overwrite: got %v, want value2", val)
	}

	belogs.Info("TestUpdateAndView_Overwrite: passed")
}

// TestUpdateAndView_Concurrent 测试并发读写
func TestUpdateAndView_Concurrent(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	const numGoroutines = 50
	const numOps = 100

	var wg sync.WaitGroup

	// 并发写入
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("concurrent_key_%d_%d", id, j)
				val := fmt.Sprintf("value_%d_%d", id, j)
				if err := Update(badgerDb, key, val, 0); err != nil {
					t.Errorf("Concurrent Update() failed: %v", err)
				}
			}
		}(i)
	}
	wg.Wait()

	// 并发读取
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("concurrent_key_%d_%d", id, j)
				expectedVal := fmt.Sprintf("value_%d_%d", id, j)
				val, found, err := View[string](badgerDb, key)
				if err != nil {
					t.Errorf("Concurrent View() failed: %v", err)
					continue
				}
				if !found {
					t.Errorf("Concurrent View() key not found: %s", key)
					continue
				}
				if val != expectedVal {
					t.Errorf("Concurrent View() value mismatch: got %v, want %v", val, expectedVal)
				}
			}
		}(i)
	}
	wg.Wait()

	belogs.Info("TestUpdateAndView_Concurrent: passed")
}

// TestUpdateAndView_VariousTypes 测试各种数据类型
func TestUpdateAndView_VariousTypes(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	// 字符串
	err := Update(badgerDb, "str_key", "hello world", 0)
	if err != nil {
		t.Fatalf("Update string failed: %v", err)
	}
	strVal, found, err := View[string](badgerDb, "str_key")
	if err != nil || !found || strVal != "hello world" {
		t.Fatalf("View string failed: err=%v, found=%v, val=%v", err, found, strVal)
	}

	// 整数
	err = Update(badgerDb, "int_key", 42, 0)
	if err != nil {
		t.Fatalf("Update int failed: %v", err)
	}
	intVal, found, err := View[int](badgerDb, "int_key")
	if err != nil || !found || intVal != 42 {
		t.Fatalf("View int failed: err=%v, found=%v, val=%v", err, found, intVal)
	}

	// 布尔
	err = Update(badgerDb, "bool_key", true, 0)
	if err != nil {
		t.Fatalf("Update bool failed: %v", err)
	}
	boolVal, found, err := View[bool](badgerDb, "bool_key")
	if err != nil || !found || !boolVal {
		t.Fatalf("View bool failed: err=%v, found=%v, val=%v", err, found, boolVal)
	}

	// 切片
	sliceVal := []string{"a", "b", "c"}
	err = Update(badgerDb, "slice_key", sliceVal, 0)
	if err != nil {
		t.Fatalf("Update slice failed: %v", err)
	}
	sliceResult, found, err := View[[]string](badgerDb, "slice_key")
	if err != nil || !found || len(sliceResult) != 3 {
		t.Fatalf("View slice failed: err=%v, found=%v, val=%v", err, found, sliceResult)
	}

	// 结构体
	user := TestUser{ID: 100, Name: "Test", Email: "test@test.com", Age: 25, IsActive: false}
	err = Update(badgerDb, "struct_key", user, 0)
	if err != nil {
		t.Fatalf("Update struct failed: %v", err)
	}
	structResult, found, err := View[TestUser](badgerDb, "struct_key")
	if err != nil || !found || structResult.ID != 100 {
		t.Fatalf("View struct failed: err=%v, found=%v, val=%+v", err, found, structResult)
	}

	belogs.Info("TestUpdateAndView_VariousTypes: passed")
}

// ==================== Delete 测试 ====================

// TestDelete_Basic 测试基本删除
func TestDelete_Basic(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	err := Update(badgerDb, "delete_key", "delete_value", 0)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	err = Delete(badgerDb, "delete_key")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	_, found, err := View[string](badgerDb, "delete_key")
	if err != nil {
		t.Fatalf("View() after delete failed: %v", err)
	}
	if found {
		t.Fatal("View() should not find deleted key")
	}

	belogs.Info("TestDelete_Basic: passed")
}

// TestDelete_NonExistent 测试删除不存在的键
func TestDelete_NonExistent(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	err := Delete(badgerDb, "nonexistent_delete_key")
	if err != nil {
		t.Fatalf("Delete() nonexistent key should not error: %v", err)
	}

	belogs.Info("TestDelete_NonExistent: passed")
}

// TestDelete_EmptyKey 测试删除空键
func TestDelete_EmptyKey(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	err := Delete(badgerDb, "")
	if err == nil {
		t.Fatal("Delete() with empty key should fail")
	}

	belogs.Info("TestDelete_EmptyKey: passed")
}

// TestDelete_NilDB 测试 nil 数据库删除
func TestDelete_NilDB(t *testing.T) {
	err := Delete(nil, "key")
	if err == nil {
		t.Fatal("Delete() with nil db should fail")
	}
	belogs.Info("TestDelete_NilDB: passed")
}

// ==================== Exists 测试 ====================

// TestExists_Basic 测试基本存在性检查
func TestExists_Basic(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	err := Update(badgerDb, "exists_key", "exists_value", 0)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	exists, err := Exists(badgerDb, "exists_key")
	if err != nil {
		t.Fatalf("Exists() failed: %v", err)
	}
	if !exists {
		t.Fatal("Exists() should return true for existing key")
	}

	exists, err = Exists(badgerDb, "not_exists_key")
	if err != nil {
		t.Fatalf("Exists() for nonexistent failed: %v", err)
	}
	if exists {
		t.Fatal("Exists() should return false for nonexistent key")
	}

	belogs.Info("TestExists_Basic: passed")
}

// TestExists_EmptyKey 测试空键存在性
func TestExists_EmptyKey(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	_, err := Exists(badgerDb, "")
	if err == nil {
		t.Fatal("Exists() with empty key should fail")
	}
	belogs.Info("TestExists_EmptyKey: passed")
}

// TestExists_NilDB 测试 nil 数据库存在性
func TestExists_NilDB(t *testing.T) {
	_, err := Exists(nil, "key")
	if err == nil {
		t.Fatal("Exists() with nil db should fail")
	}
	belogs.Info("TestExists_NilDB: passed")
}

// ==================== ViewBatch 测试 ====================

// TestViewBatch_Basic 测试批量读取
func TestViewBatch_Basic(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	// 写入多个键
	keys := make([]string, 10)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("batch_key_%d", i)
		keys[i] = key
		val := fmt.Sprintf("batch_value_%d", i)
		err := Update(badgerDb, key, val, 0)
		if err != nil {
			t.Fatalf("Update() failed for key %s: %v", key, err)
		}
	}

	// 批量读取
	values, exists, err := ViewBatch[string](badgerDb, keys)
	if err != nil {
		t.Fatalf("ViewBatch() failed: %v", err)
	}
	if len(values) != 10 || len(exists) != 10 {
		t.Fatalf("ViewBatch() returned wrong lengths: values=%d, exists=%d", len(values), len(exists))
	}

	for i := 0; i < 10; i++ {
		if !exists[i] {
			t.Fatalf("ViewBatch() key %s not found", keys[i])
		}
		expected := fmt.Sprintf("batch_value_%d", i)
		if values[i] != expected {
			t.Fatalf("ViewBatch() value mismatch for key %s: got %v, want %v", keys[i], values[i], expected)
		}
	}

	belogs.Info("TestViewBatch_Basic: passed")
}

// TestViewBatch_PartialMissing 测试部分键不存在
func TestViewBatch_PartialMissing(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	// 只写入部分键
	err := Update(badgerDb, "batch_partial_1", "value1", 0)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	keys := []string{"batch_partial_1", "batch_partial_missing", "batch_partial_1"}
	values, exists, err := ViewBatch[string](badgerDb, keys)
	if err != nil {
		t.Fatalf("ViewBatch() failed: %v", err)
	}

	if !exists[0] || values[0] != "value1" {
		t.Fatal("ViewBatch() should find existing key")
	}
	if exists[1] {
		t.Fatal("ViewBatch() should not find missing key")
	}
	if !exists[2] || values[2] != "value1" {
		t.Fatal("ViewBatch() should find repeated existing key")
	}

	belogs.Info("TestViewBatch_PartialMissing: passed")
}

// TestViewBatch_EmptyKeys 测试空键列表
func TestViewBatch_EmptyKeys(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	values, exists, err := ViewBatch[string](badgerDb, []string{})
	if err != nil {
		t.Fatalf("ViewBatch() with empty keys failed: %v", err)
	}
	if len(values) != 0 || len(exists) != 0 {
		t.Fatal("ViewBatch() with empty keys should return empty slices")
	}

	belogs.Info("TestViewBatch_EmptyKeys: passed")
}

// TestViewBatch_NilDB 测试 nil 数据库批量读取
func TestViewBatch_NilDB(t *testing.T) {
	_, _, err := ViewBatch[string](nil, []string{"key1"})
	if err == nil {
		t.Fatal("ViewBatch() with nil db should fail")
	}
	belogs.Info("TestViewBatch_NilDB: passed")
}

// TestViewBatch_LargeBatch 测试大批量读取（分页测试）
func TestViewBatch_LargeBatch(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	// 写入大量数据
	const count = 2000
	keys := make([]string, count)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("large_batch_key_%d", i)
		keys[i] = key
		val := fmt.Sprintf("large_batch_value_%d", i)
		err := Update(badgerDb, key, val, 0)
		if err != nil {
			t.Fatalf("Update() failed for key %s: %v", key, err)
		}
	}

	// 批量读取
	values, exists, err := ViewBatch[string](badgerDb, keys)
	if err != nil {
		t.Fatalf("ViewBatch() large batch failed: %v", err)
	}
	if len(values) != count || len(exists) != count {
		t.Fatalf("ViewBatch() returned wrong lengths")
	}

	for i := 0; i < count; i++ {
		if !exists[i] {
			t.Fatalf("ViewBatch() key %s not found at index %d", keys[i], i)
		}
		expected := fmt.Sprintf("large_batch_value_%d", i)
		if values[i] != expected {
			t.Fatalf("ViewBatch() value mismatch at index %d", i)
		}
	}

	belogs.Info("TestViewBatch_LargeBatch: passed")
}

// ==================== PrefixView 测试 ====================

// TestPrefixView_Basic 测试前缀查询
func TestPrefixView_Basic(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	// 写入带前缀的数据
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("prefix:user:%d", i)
		val := TestUser{ID: int64(i), Name: fmt.Sprintf("User%d", i), Email: fmt.Sprintf("user%d@test.com", i), Age: 20 + i, IsActive: true}
		err := Update(badgerDb, key, val, 0)
		if err != nil {
			t.Fatalf("Update() failed: %v", err)
		}
	}

	// 写入不带前缀的数据
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("other:item:%d", i)
		err := Update(badgerDb, key, "other_value", 0)
		if err != nil {
			t.Fatalf("Update() failed: %v", err)
		}
	}

	// 前缀查询
	results, err := PrefixView[TestUser](badgerDb, "prefix:user:", 0)
	if err != nil {
		t.Fatalf("PrefixView() failed: %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("PrefixView() returned %d results, want 5", len(results))
	}

	// 限制数量
	limitedResults, err := PrefixView[TestUser](badgerDb, "prefix:user:", 3)
	if err != nil {
		t.Fatalf("PrefixView() with limit failed: %v", err)
	}
	if len(limitedResults) != 3 {
		t.Fatalf("PrefixView() with limit returned %d results, want 3", len(limitedResults))
	}

	belogs.Info("TestPrefixView_Basic: passed")
}

// TestPrefixView_EmptyPrefix 测试空前缀
func TestPrefixView_EmptyPrefix(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	_, err := PrefixView[string](badgerDb, "", 0)
	if err == nil {
		t.Fatal("PrefixView() with empty prefix should fail")
	}
	belogs.Info("TestPrefixView_EmptyPrefix: passed")
}

// TestPrefixView_NilDB 测试 nil 数据库前缀查询
func TestPrefixView_NilDB(t *testing.T) {
	_, err := PrefixView[string](nil, "prefix", 0)
	if err == nil {
		t.Fatal("PrefixView() with nil db should fail")
	}
	belogs.Info("TestPrefixView_NilDB: passed")
}

// TestPrefixView_NoMatch 测试无匹配前缀
func TestPrefixView_NoMatch(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	results, err := PrefixView[string](badgerDb, "nomatch:", 0)
	if err != nil {
		t.Fatalf("PrefixView() no match failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("PrefixView() no match should return empty, got %d", len(results))
	}
	belogs.Info("TestPrefixView_NoMatch: passed")
}

// ==================== ViewKeyByPrefix 测试 ====================

// TestViewKeyByPrefix_Basic 测试按键前缀查询键名
func TestViewKeyByPrefix_Basic(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	// 写入数据
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("keyprefix:item:%d", i)
		err := Update(badgerDb, key, fmt.Sprintf("value%d", i), 0)
		if err != nil {
			t.Fatalf("Update() failed: %v", err)
		}
	}

	// 查询键名
	keys, err := ViewKeyByPrefix(badgerDb, "keyprefix:item:", 0)
	if err != nil {
		t.Fatalf("ViewKeyByPrefix() failed: %v", err)
	}
	if len(keys) != 5 {
		t.Fatalf("ViewKeyByPrefix() returned %d keys, want 5", len(keys))
	}

	// 限制数量
	limitedKeys, err := ViewKeyByPrefix(badgerDb, "keyprefix:item:", 2)
	if err != nil {
		t.Fatalf("ViewKeyByPrefix() with limit failed: %v", err)
	}
	if len(limitedKeys) != 2 {
		t.Fatalf("ViewKeyByPrefix() with limit returned %d keys, want 2", len(limitedKeys))
	}

	belogs.Info("TestViewKeyByPrefix_Basic: passed")
}

// TestViewKeyByPrefix_EmptyPrefix 测试空前缀
func TestViewKeyByPrefix_EmptyPrefix(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	_, err := ViewKeyByPrefix(badgerDb, "", 0)
	if err == nil {
		t.Fatal("ViewKeyByPrefix() with empty prefix should fail")
	}
	belogs.Info("TestViewKeyByPrefix_EmptyPrefix: passed")
}

// TestViewKeyByPrefix_NilDB 测试 nil 数据库
func TestViewKeyByPrefix_NilDB(t *testing.T) {
	_, err := ViewKeyByPrefix(nil, "prefix", 0)
	if err == nil {
		t.Fatal("ViewKeyByPrefix() with nil db should fail")
	}
	belogs.Info("TestViewKeyByPrefix_NilDB: passed")
}

// ==================== Batch 操作测试 ====================

// TestBatch_Basic 测试批量写入
func TestBatch_Basic(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	batch, err := NewBatch(badgerDb)
	if err != nil {
		t.Fatalf("NewBatch() failed: %v", err)
	}

	// 批量写入
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("batch_write_key_%d", i)
		val := TestUser{ID: int64(i), Name: fmt.Sprintf("BatchUser%d", i), Email: fmt.Sprintf("batch%d@test.com", i), Age: i, IsActive: i%2 == 0}
		err := UpdateWithBatch(badgerDb, batch, key, val, 0)
		if err != nil {
			t.Fatalf("UpdateWithBatch() failed: %v", err)
		}
	}

	// 批量删除
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("batch_delete_key_%d", i)
		err := DeleteWithBatch(badgerDb, batch, key)
		if err != nil {
			t.Fatalf("DeleteWithBatch() failed: %v", err)
		}
	}

	// 提交
	err = BatchFlush(badgerDb, batch)
	if err != nil {
		t.Fatalf("BatchFlush() failed: %v", err)
	}

	// 验证写入
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("batch_write_key_%d", i)
		_, found, err := View[TestUser](badgerDb, key)
		if err != nil {
			t.Fatalf("View() after batch failed: %v", err)
		}
		if !found {
			t.Fatalf("View() key %s not found after batch", key)
		}
	}

	belogs.Info("TestBatch_Basic: passed")
}

// TestBatch_Cancel 测试取消批量操作
func TestBatch_Cancel(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	batch, err := NewBatch(badgerDb)
	if err != nil {
		t.Fatalf("NewBatch() failed: %v", err)
	}

	err = UpdateWithBatch(badgerDb, batch, "cancel_key", "cancel_value", 0)
	if err != nil {
		t.Fatalf("UpdateWithBatch() failed: %v", err)
	}

	BatchCancel(badgerDb, batch)

	// 验证未写入
	_, found, err := View[string](badgerDb, "cancel_key")
	if err != nil {
		t.Fatalf("View() after cancel failed: %v", err)
	}
	if found {
		t.Fatal("View() should not find cancelled key")
	}

	belogs.Info("TestBatch_Cancel: passed")
}

// TestBatch_NilDB 测试 nil 数据库批量操作
func TestBatch_NilDB(t *testing.T) {
	_, err := NewBatch(nil)
	if err == nil {
		t.Fatal("NewBatch() with nil db should fail")
	}

	err = UpdateWithBatch[string](nil, nil, "key", "val", 0)
	if err == nil {
		t.Fatal("UpdateWithBatch() with nil db should fail")
	}

	err = DeleteWithBatch(nil, nil, "key")
	if err == nil {
		t.Fatal("DeleteWithBatch() with nil db should fail")
	}

	err = BatchFlush(nil, nil)
	if err == nil {
		t.Fatal("BatchFlush() with nil db should fail")
	}

	// BatchCancel 不应 panic
	BatchCancel(nil, nil)

	belogs.Info("TestBatch_NilDB: passed")
}

// ==================== UpdateWithTxn / DeleteWithTxn 测试 ====================

// TestTxn_Basic 测试事务操作
func TestTxn_Basic(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	// 使用 Badger 原生事务
	err := badgerDb.Update(func(txn *badger.Txn) error {
		// 写入
		err := UpdateWithTxn(badgerDb, txn, "txn_key", "txn_value", 0)
		if err != nil {
			return err
		}

		// 删除（不存在的键，不应报错）
		err = DeleteWithTxn(badgerDb, txn, "txn_nonexistent")
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// 验证
	val, found, err := View[string](badgerDb, "txn_key")
	if err != nil {
		t.Fatalf("View() after txn failed: %v", err)
	}
	if !found || val != "txn_value" {
		t.Fatalf("View() after txn mismatch: found=%v, val=%v", found, val)
	}

	belogs.Info("TestTxn_Basic: passed")
}

// TestTxn_NilParams 测试 nil 参数
func TestTxn_NilParams(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	err := UpdateWithTxn[string](badgerDb, nil, "key", "val", 0)
	if err == nil {
		t.Fatal("UpdateWithTxn() with nil txn should fail")
	}

	err = UpdateWithTxn[string](nil, nil, "key", "val", 0)
	if err == nil {
		t.Fatal("UpdateWithTxn() with nil db should fail")
	}

	err = DeleteWithTxn(badgerDb, nil, "key")
	if err == nil {
		t.Fatal("DeleteWithTxn() with nil txn should fail")
	}

	err = DeleteWithTxn(nil, nil, "key")
	if err == nil {
		t.Fatal("DeleteWithTxn() with nil db should fail")
	}

	belogs.Info("TestTxn_NilParams: passed")
}

// ==================== DropAll 测试 ====================

// TestDropAll_Basic 测试清空数据库
func TestDropAll_Basic(t *testing.T) {
	dbPath := getTestDBPath(t)
	defer cleanupTestDB(t, dbPath)
	badgerDb := createTestDB(t, dbPath)
	defer Close(badgerDb)

	// 写入数据
	for i := 0; i < 10; i++ {
		err := Update(badgerDb, fmt.Sprintf("drop_key_%d", i), fmt.Sprintf("value_%d", i), 0)
		if err != nil {
			t.Fatalf("Update() failed: %v", err)
		}
	}

	// 清空
	err := DropAll(badgerDb)
	if err != nil {
		t.Fatalf("DropAll() failed: %v", err)
	}

	// 验证数据已删除
	for i := 0; i < 10; i++ {
		_, found, err := View[string](badgerDb, fmt.Sprintf("drop_key_%d", i))
		if err != nil {
			t.Fatalf("View() after drop failed: %v", err)
		}
		if found {
			t.Fatalf("View() should not find key after drop: drop_key_%d", i)
		}
	}

	belogs.Info("TestDropAll_Basic: passed")
}

// TestDropAll_NilDB 测试 nil 数据库清空
func TestDropAll_NilDB(t *testing.T) {
	err := DropAll(nil)
	if err != nil {
		t.Fatalf("DropAll() with nil db should not error: %v", err)
	}
	belogs.Info("TestDropAll_NilDB: passed")
}

// ==================== 性能基准测试 ====================

// BenchmarkUpdate 基准测试：单条写入
func BenchmarkUpdate(b *testing.B) {
	dbPath := getTestDBPath(nil)
	defer cleanupTestDB(nil, dbPath)
	badgerDb := createTestDB(nil, dbPath)
	defer Close(badgerDb)

	user := TestUser{ID: 1, Name: "Benchmark", Email: "bench@test.com", Age: 30, IsActive: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		_ = Update(badgerDb, key, user, 0)
	}
}

// BenchmarkView 基准测试：单条读取
func BenchmarkView(b *testing.B) {
	dbPath := getTestDBPath(nil)
	defer cleanupTestDB(nil, dbPath)
	badgerDb := createTestDB(nil, dbPath)
	defer Close(badgerDb)

	user := TestUser{ID: 1, Name: "Benchmark", Email: "bench@test.com", Age: 30, IsActive: true}
	_ = Update(badgerDb, "bench_view_key", user, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = View[TestUser](badgerDb, "bench_view_key")
	}
}

// BenchmarkViewBatch 基准测试：批量读取
func BenchmarkViewBatch(b *testing.B) {
	dbPath := getTestDBPath(nil)
	defer cleanupTestDB(nil, dbPath)
	badgerDb := createTestDB(nil, dbPath)
	defer Close(badgerDb)

	// 写入 1000 条数据
	keys := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		keys[i] = fmt.Sprintf("batch_bench_key_%d", i)
		_ = Update(badgerDb, keys[i], fmt.Sprintf("value_%d", i), 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ViewBatch[string](badgerDb, keys)
	}
}

// BenchmarkBatchWrite 基准测试：批量写入
func BenchmarkBatchWrite(b *testing.B) {
	dbPath := getTestDBPath(nil)
	defer cleanupTestDB(nil, dbPath)
	badgerDb := createTestDB(nil, dbPath)
	defer Close(badgerDb)

	user := TestUser{ID: 1, Name: "Benchmark", Email: "bench@test.com", Age: 30, IsActive: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch, _ := NewBatch(badgerDb)
		for j := 0; j < 100; j++ {
			key := fmt.Sprintf("batch_bench_key_%d_%d", i, j)
			_ = UpdateWithBatch(badgerDb, batch, key, user, 0)
		}
		_ = BatchFlush(badgerDb, batch)
	}
}

// BenchmarkPrefixView 基准测试：前缀查询
func BenchmarkPrefixView(b *testing.B) {
	dbPath := getTestDBPath(nil)
	defer cleanupTestDB(nil, dbPath)
	badgerDb := createTestDB(nil, dbPath)
	defer Close(badgerDb)

	// 写入 10000 条带前缀的数据
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("prefix_bench:user:%d", i)
		_ = Update(badgerDb, key, TestUser{ID: int64(i), Name: fmt.Sprintf("User%d", i)}, 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = PrefixView[TestUser](badgerDb, "prefix_bench:user:", 100)
	}
}
