package badgerutil

import (
	"errors"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
)

var (
	badgerDB    *badger.DB
	initialized uint32
	batchSize   = 1000 // 批量写入的大小
)

// if dbPath=="memory"，则使用内存模式;
//
// if dbPath!="memory"，则使用文件模式，路径为绝对路径dbPath
func Init(dbPath string) error {
	if !atomic.CompareAndSwapUint32(&initialized, 0, 1) {
		return errors.New("badgerDB is already initialized")
	}
	var err error
	// 优化配置以适应高并发场景
	// 如果 dbPath 是关键字 "memory"，则切换为纯内存模式
	var opts badger.Options
	if dbPath == "memory" {
		opts = badger.DefaultOptions("").WithInMemory(true) // 开启内存模式
		opts = opts.WithValueThreshold(1024*1024 + 1)       // 1MB

	} else {
		err = os.MkdirAll(dbPath, os.ModePerm)
		if err != nil {
			belogs.Error("Init(): MkdirAll fail, dbPath:", dbPath, err)
			atomic.StoreUint32(&initialized, 0)
			badgerDB = nil
			return err
		}
		opts = badger.DefaultOptions(dbPath)
		opts = opts.WithValueThreshold(256 * 1024) // 磁盘模式保持 256KB
	}
	opts = opts.WithNumVersionsToKeep(1)
	opts = opts.WithValueLogFileSize(256 << 20) // 256MB 单个文件日志文件 缩小 vlog 文件大小，便于后续 GC 回收
	opts = opts.WithCompactL0OnClose(true)
	opts = opts.WithCompression(options.Snappy) //(options.ZSTD)  Snappy比ZSTD 快 5-10 倍，压缩率稍低
	opts = opts.WithBlockCacheSize(2 << 30)     // 2GB块缓存
	opts = opts.WithIndexCacheSize(512 << 20)   // 512MB索引缓存

	opts = opts.WithMemTableSize(256 * 1024 * 1024) // 256MB内存表, <=小于系统内存/4
	opts = opts.WithNumMemtables(runtime.NumCPU())  // cpunum个内存表
	opts = opts.WithNumCompactors(runtime.NumCPU()) // 增加压缩器数量
	opts = opts.WithSyncWrites(false)               // 关闭同步写，提升性能
	opts = opts.WithNumLevelZeroTables(10)          // 增大L0表阈值，减少压缩触发
	opts = opts.WithNumLevelZeroTablesStall(20)     // 增大stall阈值，避免写阻塞
	opts = opts.WithNumGoroutines(runtime.NumCPU()) // 增加并发goroutine数量

	badgerDB, err = badger.Open(opts)
	if err != nil {
		belogs.Error("Init(): Open with options fail, opts:", opts, err)
		atomic.StoreUint32(&initialized, 0)
		badgerDB = nil
		return err
	}

	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if atomic.LoadUint32(&initialized) == 0 {
				//return ticker
				belogs.Debug("badgerDB.init(): close RunValueLogGC ticker")
				return
			}
			if badgerDB == nil {
				continue
			}
			// 循环 GC 直到没有可回收文件
			for {
				err := badgerDB.RunValueLogGC(0.5) // 丢弃率 50%
				if err == badger.ErrNoRewrite {
					break // 无文件可GC
				}
				if err != nil {
					belogs.Error("badgerDB.init(): RunValueLogGC fail", err)
					break
				}
			}
		}
	}()

	return nil
}

func Close() {
	if badgerDB != nil {
		badgerDB.RunValueLogGC(0.5)
		badgerDB.Close()
	}
	atomic.StoreUint32(&initialized, 0)
}

func DropAll() error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}
	return badgerDB.DropAll()
}

// //////////////////////////////////////////////////
// base funcs
// ///////////////////////////////////////////////////

// expire: 过期时间（<=0表示永不过期）
func Update[T any](key string, value T, expire time.Duration) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}

	valueBytes := jsonutil.MarshalJsonBytes(value)
	if valueBytes == nil {
		return errors.New("failed to marshal value to JSON bytes")
	}
	expireAt := uint64(0)
	if expire > 0 {
		expireAt = uint64(time.Now().Add(expire).Unix())
	}
	return badgerDB.Update(func(txn *badger.Txn) error {
		entry := &badger.Entry{
			Key:       []byte(key),
			Value:     valueBytes,
			ExpiresAt: expireAt,
		}
		return txn.SetEntry(entry)
	})

}

// updateWithTxn 内部方法：使用外部传入的事务更新数据，不对外暴露
// 注意：此函数不会提交事务，仅在事务内执行设置操作，由调用方负责提交/回滚
func UpdateWithTxn[T any](txn *badger.Txn, key string,
	value T, expireAt uint64) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil || txn == nil {
		return errors.New("badgerDB is not initialized")
	}

	valueBytes := jsonutil.MarshalJsonBytes(value)
	if valueBytes == nil {
		return errors.New("failed to marshal value to JSON bytes")
	}
	entry := &badger.Entry{
		Key:       []byte(key),
		Value:     valueBytes,
		ExpiresAt: expireAt,
	}
	return txn.SetEntry(entry)
}

// UpdateWithBatch 内部方法：使用 WriteBatch 进行更新
func UpdateWithBatch[T any](batch *badger.WriteBatch, key string, value T, expireAt uint64) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil || batch == nil {
		return errors.New("badgerDB is not initialized")
	}

	valueBytes := jsonutil.MarshalJsonBytes(value)
	if valueBytes == nil {
		return errors.New("failed to marshal value to JSON bytes")
	}

	entry := &badger.Entry{
		Key:       []byte(key),
		Value:     valueBytes,
		ExpiresAt: expireAt,
	}

	return batch.SetEntry(entry)
}

// 返回值：value、是否找到、错误
func View[T any](key string) (T, bool, error) {
	var zero T
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return zero, false, errors.New("badgerDB is not initialized")
	}
	var value []byte
	err := badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // 键不存在返回nil，不抛错
			}
			belogs.Error("View(): Get fail, key:", key, err)
			return err
		}
		// 复制value（避免引用底层内存），并释放资源
		val, err := item.ValueCopy(nil)
		if err != nil {
			belogs.Error("View(): ValueCopy fail, item:", item, err)
			return err
		}
		value = val
		return nil
	})

	if value == nil {
		return zero, false, err
	}
	//belogs.Debug("("View(): get value", string(value))
	var result T
	err = jsonutil.UnmarshalJsonBytes(value, &result)
	if err != nil {
		belogs.Error("View(): UnmarshalJsonBytes fail, value:", string(value), err)
		return zero, false, err
	}
	return result, true, err
}

// []T 需要少于65535，如果大于，则调用ViewBatchPaged
//
// # ViewBatch 批量读取多个 key，在一个 View 事务中完成
//
// 返回值按 keys 顺序一一对应：values[i] 是 keys[i] 的值，exists[i] 表示是否找到
func viewBatchImpl[T any](keys []string) ([]T, []bool, error) {
	var zeroT T
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return nil, nil, errors.New("badgerDB is not initialized")
	}
	if len(keys) == 0 {
		return []T{}, []bool{}, nil
	}

	values := make([]T, len(keys))
	exists := make([]bool, len(keys))

	err := badgerDB.View(func(txn *badger.Txn) error {
		for i, key := range keys {
			item, err := txn.Get([]byte(key))
			if err != nil {
				if err == badger.ErrKeyNotFound {
					exists[i] = false
					values[i] = zeroT
					continue
				}
				belogs.Error("viewBatchImpl(): txn.Get fail, key:", key, "index:", i, err)
				return err
			}

			val, err := item.ValueCopy(nil)
			if err != nil {
				belogs.Error("viewBatchImpl(): ValueCopy fail, key:", key, "index:", i, err)
				return err
			}
			if len(val) == 0 {
				exists[i] = false
				values[i] = zeroT
				continue
			}

			var result T
			if err := jsonutil.UnmarshalJsonBytes(val, &result); err != nil {
				belogs.Error("viewBatchImpl(): UnmarshalJsonBytes fail, key:", key,
					"value:", string(val), err)
				return err
			}
			values[i] = result
			exists[i] = true
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return values, exists, nil
}

// ViewBatchPaged 分页批量读取，避免单事务大于65535
//
// 注意：返回的已经是全部值
func ViewBatch[T any](keys []string) ([]T, []bool, error) {
	pageSize := conf.DefaultInt("cache::badgerPageSize", 5000)
	if pageSize <= 0 {
		pageSize = 5000
	}

	values := make([]T, 0, len(keys))
	exists := make([]bool, 0, len(keys))

	for i := 0; i < len(keys); i += pageSize {
		end := i + pageSize
		if end > len(keys) {
			end = len(keys)
		}

		pageValues, pageExists, err := ViewBatch[T](keys[i:end])
		if err != nil {
			belogs.Error("ViewBatch(): ViewBatch fail, page:", i/pageSize,
				"range:", i, "-", end, err)
			return nil, nil, err
		}
		values = append(values, pageValues...)
		exists = append(exists, pageExists...)
	}

	return values, exists, nil
}

// Exists 判断key是否存在，不读取value，高性能
func Exists(key string) (bool, error) {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return false, errors.New("badgerDB is not initialized")
	}

	var exists bool
	err := badgerDB.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				exists = false
				return nil
			}
			belogs.Error("Exists(): check key fail, key:", key, err)
			return err
		}
		exists = true
		return nil
	})
	return exists, err
}

func Delete(key string) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}
	return badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}
func DeleteWithTxn(txn *badger.Txn, key string) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil || txn == nil {
		return errors.New("badgerDB is not initialized")
	}
	return txn.Delete([]byte(key))
}

// DeleteWithBatch 内部方法：使用 WriteBatch 进行删除
func DeleteWithBatch(batch *badger.WriteBatch, key string) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil || batch == nil {
		return errors.New("badgerDB is not initialized")
	}

	return batch.Delete([]byte(key))
}

// limit: <=0表示不限制返回数量
func PrefixView[T any](prefixStr string, limit int) ([]T, error) {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return nil, errors.New("badgerDB is not initialized")
	}
	results := make([]T, 0)
	prefix := []byte(prefixStr)
	err := badgerDB.View(func(txn *badger.Txn) error {
		// 构建前缀迭代器
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 100 // 预取大小，优化性能
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close() // 确保迭代器关闭

		count := 0
		// 遍历前缀匹配的KV
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if limit > 0 && count >= limit {
				break
			}
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				belogs.Error("PrefixView(): ValueCopy fail, item:", item, err)
				return err
			}

			result := make([]T, 0)
			err = jsonutil.UnmarshalJsonBytes(val, &result)
			if err != nil {
				//belogs.Debug("("PrefixView(): UnmarshalJsonBytes list fail, will try single model again, value:", string(val), err)
				var resultOne T
				err = jsonutil.UnmarshalJsonBytes(val, &resultOne)
				if err != nil {
					belogs.Error("PrefixView(): UnmarshalJsonBytes single and list both fail, value:", string(val), err)
					results = nil
					return err
				}
				results = append(results, resultOne)
			} else {
				results = append(results, result...)
			}
			count++
		}
		return nil
	})
	return results, err
}

// limit: <=0表示不限制返回数量
func ViewKeyByPrefix(prefixStr string, limit int) ([]string, error) {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return nil, errors.New("badgerDB is not initialized")
	}
	keys := make([]string, 0)
	prefix := []byte(prefixStr)
	err := badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // 关键优化：不读Value
		opts.AllVersions = false    // 不读历史版本
		count := 0
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if limit > 0 && count >= limit {
				break
			}
			key := it.Item().KeyCopy(nil) // 安全拷贝
			keys = append(keys, string(key))
			count++
		}
		return nil
	})
	return keys, err
}

func NewBatch() (*badger.WriteBatch, error) {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return nil, errors.New("badgerDB is not initialized")
	}
	return badgerDB.NewWriteBatch(), nil
}
func BatchCancel(batch *badger.WriteBatch) {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return
	}
	batch.Cancel()
}
func BatchFlush(batch *badger.WriteBatch) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}
	if err := batch.Flush(); err != nil {
		//	belogs.Error("Flush(): Flush final batch fail, err:", err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////
// advance funcs
/////////////////////////////////////////////////////
/*
// Append 追加功能：key存在则将value追加到列表，不存在则直接存储（自动转为数组）
// expire: 过期时间（<=0表示永不过期）
func Append[T any](key string, value T, expire time.Duration) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}

	expireAt := uint64(0)
	if expire > 0 {
		expireAt = uint64(time.Now().Add(expire).Unix())
	}

	return badgerDB.Update(func(txn *badger.Txn) error {
		var values []T
		// 1. 查询key是否存在
		item, err := txn.Get([]byte(key))
		if err != nil && err != badger.ErrKeyNotFound {
			belogs.Error("Append(): get key fail, key:", key, err)
			return err
		}

		// 2. 存在则读取原有数据
		if err == nil {
			valCopy, err := item.ValueCopy(nil)
			if err != nil {
				belogs.Error("Append(): ValueCopy fail, key:", key, err)
				return err
			}
			//belogs.Debug("("Append(): get valCopy:", string(valCopy))
			// 反序列化为数组
			if err := jsonutil.UnmarshalJsonBytes(valCopy, &values); err != nil {
				belogs.Error("Append(): Unmarshal existing value fail, key:", key, err)
				return err
			}
		}

		// 3. 追加新数据
		values = append(values, value)
		//belogs.Debug("Append(): new values:", jsonutil.MarshalJson(values))
		// 4. 序列化新数组
		newValueBytes := jsonutil.MarshalJsonBytes(values)
		if newValueBytes == nil {
			return errors.New("failed to marshal appended value to JSON bytes")
		}

		// 5. 写入数据库
		entry := &badger.Entry{
			Key:       []byte(key),
			Value:     newValueBytes,
			ExpiresAt: expireAt,
		}
		return txn.SetEntry(entry)
	})
}

// AppendWithTxn 内部方法：使用外部传入的事务进行追加操作，不对外暴露
// 注意：此函数不会提交事务，仅在事务内执行读写，由调用方负责提交/回滚
func AppendWithTxn[T any](txn *badger.Txn,
	key string, value T, expireAt uint64) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil || txn == nil {
		return errors.New("badgerDB is not initialized")
	}

	var values []T

	// 1. 查询 key 是否存在
	item, err := txn.Get([]byte(key))
	if err != nil && err != badger.ErrKeyNotFound {
		belogs.Error("AppendWithTxn(): get key fail, key:", key, err)
		return err
	}

	// 2. 存在则读取原有数据并反序列化
	if err == nil {
		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			belogs.Error("AppendWithTxn(): ValueCopy fail, key:", key, err)
			return err
		}

		if err := jsonutil.UnmarshalJsonBytes(valCopy, &values); err != nil {
			belogs.Error("AppendWithTxn(): Unmarshal existing value fail, key:", key, err)
			return err
		}
	}

	// 3. 追加新值
	values = append(values, value)

	// 4. 序列化新数组
	newValueBytes := jsonutil.MarshalJsonBytes(values)
	if newValueBytes == nil {
		return errors.New("AppendWithTxn: failed to marshal appended value to JSON bytes")
	}

	// 5. 使用传入的事务设置条目（带过期时间）
	entry := &badger.Entry{
		Key:       []byte(key),
		Value:     newValueBytes,
		ExpiresAt: expireAt,
	}
	return txn.SetEntry(entry)
}

// AppendWithBatch 内部方法：使用 WriteBatch 进行追加
// 注意：WriteBatch 不支持读取，必须先手动读一遍
// AppendWithBatch 内部方法：使用 WriteBatch 进行追加
// 🔥 重点：不再需要外部传入 txn，内部自动创建/销毁
func AppendWithBatch[T any](batch *badger.WriteBatch,
	key string, value T, expireAt uint64) error {

	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil || batch == nil {
		return errors.New("badgerDB is not initialized")
	}

	var values []T

	// ==============================================
	// 🔥 内部自动创建只读事务，外部完全不用传！
	// ==============================================
	txn := badgerDB.NewTransaction(false)
	defer txn.Discard()

	// 1. 查询 key 是否存在（内部 txn 读取）
	item, err := txn.Get([]byte(key))
	if err != nil && err != badger.ErrKeyNotFound {
		belogs.Error("AppendWithBatch(): get key fail, key:", key, err)
		return err
	}

	// 2. 存在则读取原有数据
	if err == nil {
		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			belogs.Error("AppendWithBatch(): ValueCopy fail, key:", key, err)
			return err
		}

		if err := jsonutil.UnmarshalJsonBytes(valCopy, &values); err != nil {
			belogs.Error("AppendWithBatch(): Unmarshal existing value fail, key:", key, err)
			return err
		}
	}

	// 3. 追加新值
	values = append(values, value)

	// 4. 序列化
	newValueBytes := jsonutil.MarshalJsonBytes(values)
	if newValueBytes == nil {
		return errors.New("AppendWithBatch: failed to marshal appended value to JSON bytes")
	}

	// 5. 写入 batch
	entry := &badger.Entry{
		Key:       []byte(key),
		Value:     newValueBytes,
		ExpiresAt: expireAt,
	}

	return batch.SetEntry(entry)
}

const (
	MAINKEY_TO_VALUE    = "_value"
	MAINKEY_TO_OUTERKEY = "_outerkey"
)

// BatchUpdateKeyFunc
// T: 泛型数据类型
// datas: 待批量写入的数据切片
// expire: 过期时间，<=0表示永不过期
// batchSize: 每批次写入的数量，必须大于0
// mainKeyFunc，outerKeyFunc:
//
//	outerKey1, outerKey2, subk3 --> mainKey --> value
//	mainKey --> [outerKey1,outerKey2,outerKey3...]
//
// 方便当del时，传入一个outerKey1，找到对应mainKey，删除value外，还能根据maiKey找到其他的outerKey对应
func BatchUpdateByMultiKeys[T any](datas []T, expire time.Duration, batchSize int,
	mainKeyFunc func(T) string,
	outerUpdateKeyFunc func(T) []string,
	outerAppendKeyFunc func(T) []string) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}
	if batchSize <= 0 {
		return errors.New("batchSize must be greater than 0")
	}
	if mainKeyFunc == nil {
		return errors.New("mainKeyFunc cannot be empty")
	}
	if outerUpdateKeyFunc == nil {
		return errors.New("outerUpdateKeyFunc cannot be empty")
	}
	if outerAppendKeyFunc == nil {
		return errors.New("outerAppendKeyFunc cannot be empty")
	}
	if len(datas) == 0 {
		return nil
	}

	expireAt := uint64(0)
	if expire > 0 {
		expireAt = uint64(time.Now().Add(expire).Unix())
	}

	batch := badgerDB.NewWriteBatch()
	defer batch.Cancel()

	for dataIdx, value := range datas {
		mainKey := mainKeyFunc(value)
		mainEntry := &badger.Entry{
			Key:       []byte(mainKey + MAINKEY_TO_VALUE),
			Value:     jsonutil.MarshalJsonBytes(value),
			ExpiresAt: expireAt,
		}
		if err := batch.SetEntry(mainEntry); err != nil {
			belogs.Error("BatchUpdateKeyFunc(): SetEntry mainEntry fail, mainKey:", mainKey, err)
			return err
		}
		//belogs.Debug("BatchUpdateKeyFunc(): SetEntry mainEntry, mainKey:", mainKey)

		outerKeys := outerUpdateKeyFunc(value)
		for _, outerKey := range outerKeys {
			outerEntry := &badger.Entry{
				Key:       []byte(outerKey),
				Value:     []byte(mainKey), // ✅ 修复：直接存字符串字节，不要JSON序列化（去掉双引号）
				ExpiresAt: expireAt,
			}
			if err := batch.SetEntry(outerEntry); err != nil {
				belogs.Error("BatchUpdateKeyFunc(): SetEntry outerEntry fail, outerKey:", outerKey, err)
				return err
			}
			//belogs.Debug("("BatchUpdateKeyFunc(): SetEntry outerEntry, outerKey:", outerKey)
		}

		mainOuterEntry := &badger.Entry{
			Key:       []byte(mainKey + MAINKEY_TO_OUTERKEY),
			Value:     jsonutil.MarshalJsonBytes(outerKeys), // 切片必须JSON序列化，保留
			ExpiresAt: expireAt,
		}
		if err := batch.SetEntry(mainOuterEntry); err != nil {
			belogs.Error("BatchUpdateKeyFunc(): SetEntry mainOuterEntry fail, mainKey:", mainKey, err)
			return err
		}
		//belogs.Debug("BatchUpdateKeyFunc(): SetEntry mainOuterEntry, mainKey:", mainKey)

		if (dataIdx+1)%batchSize == 0 {
			if err := batch.Flush(); err != nil {
				belogs.Error("BatchUpdateKeyFunc(): Flush every batchSize fail, err:", err)
				return err
			}
			batch.Cancel()
			batch = badgerDB.NewWriteBatch()
		}
	}

	if err := batch.Flush(); err != nil {
		belogs.Error("BatchUpdateKeyFunc(): Flush final batch fail, err:", err)
		return err
	}

	return nil
}

// ViewByMultiKeys：通过 outerKey 查询最终的 Value 数据
// 执行流程：
// 1. 根据 outerKey 获取 mainKey
// 2. 根据 mainKey 查询对应的真实 Value 数据
// 3. 反序列化 Value 为指定泛型类型并返回
// T: 泛型数据类型（与写入时的类型一致）
// outerKey: 外部查询键
// 返回值：查询到的Value指针、错误信息
func ViewByMultiKeys[T any](outerKey string) (*T, error) {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return nil, errors.New("badgerDB is not initialized")
	}
	if outerKey == "" {
		return nil, errors.New("outerKey cannot be empty")
	}

	var result T
	err := badgerDB.View(func(txn *badger.Txn) error {
		mainKeyItem, err := txn.Get([]byte(outerKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				//belogs.Debug("("ViewByMultiKeys(): outerKey not found, outerKey:", outerKey)
				return badger.ErrKeyNotFound
			}
			belogs.Error("ViewByMultiKeys(): get mainKey by outerKey fail, outerKey:", outerKey, err)
			return err
		}

		mainKeyBytes, err := mainKeyItem.ValueCopy(nil)
		if err != nil {
			belogs.Error("ViewByMultiKeys(): ValueCopy mainKey fail, outerKey:", outerKey, err)
			return err
		}
		mainKey := string(mainKeyBytes) // ✅ 修复：直接转字符串，无引号
		//belogs.Debug("ViewByMultiKeys(): get mainKey by outerKey success, outerKey:", outerKey, "mainKey:", mainKey)

		valueKey := mainKey + MAINKEY_TO_VALUE
		valueItem, err := txn.Get([]byte(valueKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				//belogs.Debug("("ViewByMultiKeys(): value not found by mainKey, mainKey:", mainKey)
				return badger.ErrKeyNotFound
			}
			belogs.Error("ViewByMultiKeys(): get value by mainKey fail, valueKey:", valueKey, err)
			return err
		}

		valueBytes, err := valueItem.ValueCopy(nil)
		if err != nil {
			belogs.Error("ViewByMultiKeys(): ValueCopy value fail, valueKey:", valueKey, err)
			return err
		}

		if err := jsonutil.UnmarshalJsonBytes(valueBytes, &result); err != nil {
			belogs.Error("ViewByMultiKeys(): UnmarshalJson value to T fail, valueKey:", valueKey, err)
			return err
		}
		return nil
	})

	if err != nil {
		belogs.Error("ViewByMultiKeys(): fail:", err)
		return nil, err
	}
	return &result, nil
}

// DeleteByOuterKey：通过 outerKey 删除整条数据（含清理所有关联 outerKey）
// 执行流程：
// 1. 根据 outerKey 获取 mainKey
// 2. 删除 mainKey 对应的真实 value 数据
// 3. 遍历并删除所有 value = mainKey 的 outerKey（完整清理）
func DeleteByMultiKeys(outerKey string) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}

	return badgerDB.Update(func(txn *badger.Txn) error {
		mainKeyItem, err := txn.Get([]byte(outerKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				//belogs.Debug("("DeleteByMultiKeys(): outerKey not found, outerKey:", outerKey)
				return nil
			}
			belogs.Error("DeleteByMultiKeys(): get mainKey by outerKey fail, outerKey:", outerKey, err)
			return err
		}

		mainKeyBytes, err := mainKeyItem.ValueCopy(nil)
		if err != nil {
			belogs.Error("DeleteByMultiKeys(): ValueCopy mainKey fail, outerKey:", outerKey, err)
			return err
		}
		mainKey := string(mainKeyBytes) // ✅ 修复：直接转字符串，无引号

		// 删除 value
		err = txn.Delete([]byte(mainKey + MAINKEY_TO_VALUE))
		if err != nil {
			belogs.Error("DeleteByMultiKeys(): delete main value fail, key:", mainKey+MAINKEY_TO_VALUE, err)
			return err
		}

		// 获取并删除所有 outerKeys
		outerKeysItems, err := txn.Get([]byte(mainKey + MAINKEY_TO_OUTERKEY))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			belogs.Error("DeleteByMultiKeys(): Get outerKeys fail, key:", mainKey+MAINKEY_TO_OUTERKEY, err)
			return err
		}
		vals, err := outerKeysItems.ValueCopy(nil)
		if err != nil {
			belogs.Error("DeleteByMultiKeys(): ValueCopy fail", err)
			return err
		}
		var outKeys []string
		if err = jsonutil.UnmarshalJsonBytes(vals, &outKeys); err != nil {
			belogs.Error("DeleteByMultiKeys(): Unmarshal outerKeys fail", err)
			return err
		}
		for _, outKey := range outKeys {
			if err = txn.Delete([]byte(outKey)); err != nil {
				belogs.Error("DeleteByMultiKeys(): delete outKey fail, outKey:", outKey, err)
				return err
			}
		}

		// 删除 outerKeys 映射
		err = txn.Delete([]byte(mainKey + MAINKEY_TO_OUTERKEY))
		if err != nil {
			belogs.Error("DeleteByMultiKeys(): delete main outerKey fail", err)
			return err
		}

		//belogs.Debug("("DeleteByMultiKeys(): success, outerKey:", outerKey, "mainKey:", mainKey)
		return nil
	})
}
*/
