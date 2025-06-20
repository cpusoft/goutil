package badgerdbutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")

	InitDB(opts)

	defer CloseDB()
	// *************** 1.insert  str **************
	err := Insert("name", "张三")
	assert.Nil(t, err)
	// *************** 2. Get  **************
	name, exist, err := Get[string]("name")
	assert.Nil(t, err)
	assert.Equal(t, true, exist)

	assert.Equal(t, "张三", strings.Trim(name, "\""))
	// *************** 3. Delete  **************
	err = Delete("name")
	assert.Nil(t, err)
	// *************** 4. Get  **************
	_, exist, err = Get[string]("name")
	assert.Nil(t, err)
	assert.Equal(t, false, exist)
	// *************** 4. Get  **************
	// *************** 4. Get  **************
	// *************** 4. Get  **************
	// *************** 4. Get  **************
	type Student struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Age  int64  `json:"age"`
	}
	student := Student{
		ID:   "123456789",
		Name: "王五",
		Age:  int64(13),
	}
	err = Insert[Student](student.ID, student)
	assert.Nil(t, err)

	tStu, exist, err := Get[Student](student.ID)
	assert.Nil(t, err)
	assert.Equal(t, true, exist)
	assert.Equal(t, "王五", tStu.Name)

}

func TestInsertWithTxn(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")

	InitDB(opts)

	defer CloseDB()

	RunWithTxn(func(txn *badger.Txn) error {
		// *************** 1.insert  **************
		err := InsertWithTxn(txn, "name", "李四")
		assert.Nil(t, err)
		// *************** 2. Get  **************
		name, exist, err := GetWithTxn[string](txn, "name")
		assert.Nil(t, err)
		assert.Equal(t, true, exist)

		assert.Equal(t, "李四", strings.Trim(name, "\""))

		return nil
	})

	RunWithTxn(func(txn *badger.Txn) error {

		// *************** 3. Delete  **************
		err := DeleteWithTxn(txn, "name")
		assert.Nil(t, err)
		// *************** 4. Get  **************
		_, exist, err := GetWithTxn[string](txn, "name")
		assert.Nil(t, err)
		assert.Equal(t, false, exist)

		return nil
	})

}

func TestUtilStoreWithCompositeKey(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")

	InitDB(opts)

	defer CloseDB()

	columns := map[string]string{
		"name": "zhangsan",
		"age":  "30",
	}

	err := StoreWithCompositeKey("user", "123", columns)
	assert.Nil(t, err)

	columnsToQuery := map[string]string{
		"name": "zhangsan",
		"age":  "30",
	}

	result, exist, err := QueryByCompositeKey[string]("user", columnsToQuery)
	assert.Nil(t, err)
	assert.Equal(t, true, exist)
	assert.Equal(t, "123:user", result)
}

func TestBatchQueryByPrefixStream(t *testing.T) {
	opts := badger.DefaultOptions("./badgerdb")
	err := InitDB(opts)
	assert.Nil(t, err)
	defer CloseDB()

	// 清理测试数据
	RunWithTxn(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			_ = txn.Delete(it.Item().Key())
		}
		return nil
	})

	t.Run("Test_Small_Dataset_String", func(t *testing.T) {
		smallPrefix := "small_test:"

		// 清理之前的测试数据
		RunWithTxn(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()

			prefix := []byte(smallPrefix)
			for it.Seek(prefix); it.Valid() && bytes.HasPrefix(it.Item().Key(), prefix); it.Next() {
				_ = txn.Delete(it.Item().Key())
			}
			return nil
		})

		// 准备测试数据
		testData := map[string]string{
			smallPrefix + "key1": "value1",
			smallPrefix + "key2": "value2",
			smallPrefix + "key3": "value3",
			"other:key1":         "value4", // 不同前缀
		}

		for k, v := range testData {
			err := Insert(k, v)
			assert.Nil(t, err)
		}

		// 测试流式查询
		count := 0
		results := make(map[string]string)
		err := BatchQueryByPrefixStream(smallPrefix, DefaultQueryOptions(), func(key []byte, value []byte) error {
			count++
			results[string(key)] = string(value)
			return nil
		})

		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, "value1", strings.Trim(results[smallPrefix+"key1"], "\""))
	})

	t.Run("Test_Large_Dataset_Int", func(t *testing.T) {
		largePrefix := "large_test:"

		// 清理之前的测试数据
		RunWithTxn(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()

			prefix := []byte(largePrefix)
			for it.Seek(prefix); it.Valid() && bytes.HasPrefix(it.Item().Key(), prefix); it.Next() {
				_ = txn.Delete(it.Item().Key())
			}
			return nil
		})

		// 准备大量测试数据
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("%skey%d", largePrefix, i)
			err := Insert(key, i)
			assert.Nil(t, err)
		}

		count := 0
		sum := 0
		err := BatchQueryByPrefixStream(largePrefix, DefaultQueryOptions(), func(key []byte, value []byte) error {
			count++
			val, _ := strconv.Atoi(strings.Trim(string(value), "\""))
			sum += val
			return nil
		})

		assert.Nil(t, err)
		assert.Equal(t, 1000, count)
	})

	t.Run("Test_Struct_Dataset", func(t *testing.T) {
		type TestStruct struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}

		// 准备结构体测试数据
		testStructs := []TestStruct{
			{ID: 1, Name: "Test1"},
			{ID: 2, Name: "Test2"},
			{ID: 3, Name: "Test3"},
		}

		// 清理之前的测试数据
		RunWithTxn(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()

			prefix := []byte("struct:")
			for it.Seek(prefix); it.Valid() && bytes.HasPrefix(it.Item().Key(), prefix); it.Next() {
				_ = txn.Delete(it.Item().Key())
			}
			return nil
		})

		for _, ts := range testStructs {
			key := fmt.Sprintf("struct:key%d", ts.ID)
			err := Insert(key, ts)
			assert.Nil(t, err)
		}

		count := 0
		results := make([]TestStruct, 0)
		err := BatchQueryByPrefixStream("struct:", DefaultQueryOptions(), func(key []byte, value []byte) error {
			count++
			var ts TestStruct
			err := json.Unmarshal(value, &ts)
			assert.Nil(t, err)
			results = append(results, ts)
			return nil
		})

		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, "Test1", results[0].Name)
	})

	t.Run("Test_Error_Handling", func(t *testing.T) {
		expectedError := errors.New("test error")

		// 使用一个存在的前缀，确保回调函数被调用
		smallPrefix := "small_test:"

		err := BatchQueryByPrefixStream(smallPrefix, DefaultQueryOptions(), func(key []byte, value []byte) error {
			// 返回一个错误，应该中断迭代并返回该错误
			return expectedError
		})

		assert.Equal(t, expectedError, err)
	})

	t.Run("Test_Empty_Prefix", func(t *testing.T) {
		// 使用一个不存在的前缀
		nonExistentPrefix := "nonexistent_prefix_that_should_not_match_anything:"

		count := 0
		err := BatchQueryByPrefixStream(nonExistentPrefix, DefaultQueryOptions(), func(key []byte, value []byte) error {
			count++
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Test_Batch_Processing", func(t *testing.T) {
		// 准备批量测试数据
		batchSize := 100
		totalItems := 550 // 不是批次大小的整数倍
		batchPrefix := "batch_test:"

		// 清理之前的测试数据
		RunWithTxn(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()

			prefix := []byte(batchPrefix)
			for it.Seek(prefix); it.Valid() && bytes.HasPrefix(it.Item().Key(), prefix); it.Next() {
				_ = txn.Delete(it.Item().Key())
			}
			return nil
		})

		// 插入测试数据
		for i := 0; i < totalItems; i++ {
			key := fmt.Sprintf("%skey%d", batchPrefix, i)
			value := fmt.Sprintf("value%d", i)
			err := Insert(key, value)
			assert.Nil(t, err)
		}

		processedCount := 0
		batchCount := 0
		lastBatchSize := 0

		err := BatchQueryByPrefixStreamWithBatch[string](batchPrefix, batchSize, func(batch []string) error {
			batchCount++
			lastBatchSize = len(batch)
			processedCount += len(batch)
			return nil
		})

		assert.Nil(t, err)
		assert.Equal(t, totalItems, processedCount)
		assert.Equal(t, 6, batchCount)     // 5个完整批次(100)加1个不完整批次(50)
		assert.Equal(t, 50, lastBatchSize) // 最后一个批次应该有50个项目
	})
}

// 测试并发场景
func TestBatchQueryByPrefixStreamConcurrent(t *testing.T) {
	opts := badger.DefaultOptions("./badgerdb")
	err := InitDB(opts)
	assert.Nil(t, err)
	defer CloseDB()

	// 准备测试数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("concurrent:key%d", i)
		err := Insert(key, i)
		assert.Nil(t, err)
	}

	var wg sync.WaitGroup
	concurrentQueries := 5

	results := make([]int, concurrentQueries)
	for i := 0; i < concurrentQueries; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			count := 0
			err := BatchQueryByPrefixStream("concurrent:", DefaultQueryOptions(), func(key []byte, value []byte) error {
				count++
				return nil
			})
			assert.Nil(t, err)
			results[index] = count
		}(i)
	}

	wg.Wait()

	// 验证所有查询都返回了相同的结果数量
	for i := 1; i < concurrentQueries; i++ {
		assert.Equal(t, results[0], results[i])
		assert.Equal(t, 1000, results[i])
	}
}

// 测试性能
func BenchmarkBatchQueryByPrefixStream(b *testing.B) {
	opts := badger.DefaultOptions("./badgerdb")
	err := InitDB(opts)
	if err != nil {
		b.Fatal(err)
	}
	defer CloseDB()

	benchPrefix := "bench_test:"

	// 清理之前的测试数据
	RunWithTxn(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(benchPrefix)
		for it.Seek(prefix); it.Valid() && bytes.HasPrefix(it.Item().Key(), prefix); it.Next() {
			_ = txn.Delete(it.Item().Key())
		}
		return nil
	})

	// 准备基准测试数据
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("%skey%d", benchPrefix, i)
		value := fmt.Sprintf("value%d", i)
		err := Insert(key, value)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	b.Run("Sequential_Read", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := BatchQueryByPrefixStream(benchPrefix, DefaultQueryOptions(), func(key []byte, value []byte) error {
				return nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Batch_Read", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := BatchQueryByPrefixStreamWithBatch[string](benchPrefix, 1000, func(batch []string) error {
				return nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestBatchSetOperations(t *testing.T) {
	//// 创建临时目录作为测试数据库路径
	//tempDir, err := os.MkdirTemp("", "badger-test-*")
	//if err != nil {
	//	t.Fatalf("无法创建临时目录: %v", err)
	//}
	//defer os.RemoveAll(tempDir)
	//
	//// 设置badger数据库选项
	//dbPath := filepath.Join(tempDir, "badger")
	//opts := badger.DefaultOptions(dbPath).WithLogger(nil)
	//
	//// 初始化数据库
	//err = InitDB(opts)
	//if err != nil {
	//	t.Fatalf("初始化数据库失败: %v", err)
	//}
	//defer CloseDB()

	opts := badger.DefaultOptions("./badgerdb")
	err := InitDB(opts)
	assert.Nil(t, err)
	defer CloseDB()

	// 测试用的前缀
	testPrefix := "TEST:PREFIX:UPDATE"

	// 1. 测试大批量添加元素
	t.Run("批量添加元素", func(t *testing.T) {
		// 准备要添加的大量数据
		const numItems = 10000
		testValues := make([]string, numItems)
		for i := 0; i < numItems; i++ {
			testValues[i] = "value-" + strconv.Itoa(i)
		}

		// 批量添加到集合
		err := BatchAddToSet(testPrefix, testValues)
		assert.NoError(t, err, "批量添加元素失败")

		// 验证是否成功添加
		err = RunWithTxn(func(txn *badger.Txn) error {
			// 获取集合中的所有元素
			setMap, err := GetSetWithPrefixTxn[string](txn, testPrefix)
			if err != nil {
				return err
			}

			// 检查元素数量
			var totalElements int
			for _, values := range setMap {
				totalElements += len(values)
			}
			assert.Equal(t, numItems, totalElements, "添加的元素数量与预期不符")
			return nil
		})
		assert.NoError(t, err, "验证添加结果失败")
	})

	// 2. 测试获取所有键
	t.Run("获取所有键", func(t *testing.T) {
		var keys []string
		err := RunWithTxn(func(txn *badger.Txn) error {
			// 模拟 GetRecentUpdateKeyWithTxn 函数
			setMap, err := GetSetWithPrefixTxn[string](txn, testPrefix)
			if err != nil {
				return err
			}

			// 将所有值转换为字符串
			for _, values := range setMap {
				for _, value := range values {
					keys = append(keys, value)
				}
			}
			return nil
		})
		assert.NoError(t, err, "获取所有键失败")
		assert.Equal(t, 10000, len(keys), "键的数量与预期不符")
	})

	// 3. 测试分批删除元素 - 简单数据集
	t.Run("分批删除元素", func(t *testing.T) {
		// 使用分批删除功能
		startTime := time.Now()
		err := ClearSetWithPrefixBatched(testPrefix, 100)
		duration := time.Since(startTime)

		t.Logf("删除10,000个元素耗时: %v", duration)
		assert.NoError(t, err, "分批删除元素失败")

		// 验证是否全部删除
		err = RunWithTxn(func(txn *badger.Txn) error {
			setMap, err := GetSetWithPrefixTxn[string](txn, testPrefix)
			if err != nil {
				return err
			}

			var totalElements int
			for _, values := range setMap {
				totalElements += len(values)
			}
			assert.Equal(t, 0, totalElements, "集合没有被完全清空")
			return nil
		})
		assert.NoError(t, err, "验证删除结果失败")
	})

	// 4. 测试大型事务分批处理（模拟真实情况）
	t.Run("真实场景模拟", func(t *testing.T) {
		// 先添加大量数据
		const numItems = 100000 // 增加数据量以更好地测试性能
		testValues := make([]string, numItems)
		for i := 0; i < numItems; i++ {
			testValues[i] = "bigvalue-" + strconv.Itoa(i) + "-" + strconv.Itoa(i*i)
		}

		// 记录添加开始时间
		addStartTime := time.Now()

		// 分批添加到集合
		batchSize := 5000 // 增加批处理大小以加快添加速度
		for i := 0; i < numItems; i += batchSize {
			end := i + batchSize
			if end > numItems {
				end = numItems
			}
			err := BatchAddToSet(testPrefix, testValues[i:end])
			assert.NoError(t, err, "批量添加元素失败")
		}

		addDuration := time.Since(addStartTime)
		t.Logf("添加100,000个元素耗时: %v, 平均速率: %.2f 元素/秒",
			addDuration, float64(numItems)/addDuration.Seconds())

		// 验证添加成功
		var totalAdded int
		err := RunWithTxn(func(txn *badger.Txn) error {
			setMap, err := GetSetWithPrefixTxn[string](txn, testPrefix)
			if err != nil {
				return err
			}

			for _, values := range setMap {
				totalAdded += len(values)
			}
			return nil
		})
		assert.NoError(t, err, "验证添加结果失败")
		t.Logf("成功添加 %d 个元素", totalAdded)

		// 使用优化后的分批删除功能，测试不同批处理大小的性能
		testBatchSizes := []int{100, 1000, 5000, 10000}
		for _, bSize := range testBatchSizes {
			// 克隆前缀，以便每次测试使用不同前缀
			testPrefixWithBatch := fmt.Sprintf("%s:BATCH_%d", testPrefix, bSize)

			// 先复制数据到新前缀
			err = RunWithTxn(func(txn *badger.Txn) error {
				setMap, err := GetSetWithPrefixTxn[string](txn, testPrefix)
				if err != nil {
					return err
				}

				// 复制部分数据用于测试
				count := 0
				for _, values := range setMap {
					for _, value := range values {
						if count < 50000 { // 只复制5万条，加快测试速度
							err = AddToSetWithTxn(txn, testPrefixWithBatch, string(value))
							if err != nil {
								return err
							}
							count++
						} else {
							break
						}
					}
					if count >= 50000 {
						break
					}
				}
				return nil
			})
			assert.NoError(t, err, "复制数据失败")

			// 使用我们的分批删除功能
			startTime := time.Now()
			err := ClearSetWithPrefixBatched(testPrefixWithBatch, bSize)
			duration := time.Since(startTime)

			t.Logf("批处理大小 %d: 删除约50,000个元素耗时: %v, 平均速率: %.2f 元素/秒",
				bSize, duration, 50000.0/duration.Seconds())

			assert.NoError(t, err, "分批删除大量元素失败")

			// 验证是否全部删除
			err = RunWithTxn(func(txn *badger.Txn) error {
				setMap, err := GetSetWithPrefixTxn[string](txn, testPrefixWithBatch)
				if err != nil {
					return err
				}

				var remainingElements int
				for _, values := range setMap {
					remainingElements += len(values)
				}
				assert.Equal(t, 0, remainingElements, "集合没有被完全清空")
				return nil
			})
			assert.NoError(t, err, "验证删除结果失败")
		}

		// 最后测试删除全部数据的性能
		startTime := time.Now()
		err = ClearSetWithPrefixBatched(testPrefix, 10000) // 使用较大的批处理大小
		duration := time.Since(startTime)

		t.Logf("删除全部数据(约%d个元素)耗时: %v, 平均速率: %.2f 元素/秒",
			totalAdded, duration, float64(totalAdded)/duration.Seconds())

		assert.NoError(t, err, "分批删除全部元素失败")

		// 验证是否全部删除
		err = RunWithTxn(func(txn *badger.Txn) error {
			setMap, err := GetSetWithPrefixTxn[string](txn, testPrefix)
			if err != nil {
				return err
			}

			var remainingElements int
			for _, values := range setMap {
				remainingElements += len(values)
			}
			assert.Equal(t, 0, remainingElements, "集合没有被完全清空")
			return nil
		})
		assert.NoError(t, err, "验证最终删除结果失败")
	})

	// 5. 新增测试: 测试超大数据集的性能（选择性运行）
	if testing.Short() {
		t.Skip("跳过超大数据集测试")
	}

	t.Run("超大数据集性能测试", func(t *testing.T) {
		// 添加40万条数据用于性能测试
		const numItems = 400000

		// 清理可能存在的旧数据
		ClearSetWithPrefixBatched("LARGE_TEST_PREFIX", 10000)

		// 分批添加数据
		t.Log("开始添加40万条测试数据...")
		addStartTime := time.Now()

		batchSize := 10000
		for i := 0; i < numItems; i += batchSize {
			end := i + batchSize
			if end > numItems {
				end = numItems
			}

			// 创建批次数据
			testValues := make([]string, end-i)
			for j := 0; j < end-i; j++ {
				testValues[j] = fmt.Sprintf("large-test-value-%d", i+j)
			}

			// 添加到集合
			err := BatchAddToSet("LARGE_TEST_PREFIX", testValues)
			assert.NoError(t, err, "批量添加超大数据集元素失败")

			// 报告进度
			if (i+batchSize)%100000 == 0 || i+batchSize >= numItems {
				t.Logf("已添加 %d/%d 条数据...", i+batchSize, numItems)
			}
		}

		addDuration := time.Since(addStartTime)
		t.Logf("添加40万条数据耗时: %v, 平均速率: %.2f 元素/秒",
			addDuration, float64(numItems)/addDuration.Seconds())

		// 测试删除性能
		t.Log("开始删除40万条测试数据...")
		deleteStartTime := time.Now()

		err := ClearSetWithPrefixBatched("LARGE_TEST_PREFIX", 10000)
		deleteDuration := time.Since(deleteStartTime)

		t.Logf("删除40万条数据耗时: %v, 平均速率: %.2f 元素/秒",
			deleteDuration, float64(numItems)/deleteDuration.Seconds())

		assert.NoError(t, err, "分批删除超大数据集失败")

		// 验证删除成功
		err = RunWithTxn(func(txn *badger.Txn) error {
			setMap, err := GetSetWithPrefixTxn[string](txn, "LARGE_TEST_PREFIX")
			if err != nil {
				return err
			}

			var remainingElements int
			for _, values := range setMap {
				remainingElements += len(values)
			}
			assert.Equal(t, 0, remainingElements, "超大数据集没有被完全清空")
			return nil
		})
		assert.NoError(t, err, "验证超大数据集删除结果失败")
	})
}
