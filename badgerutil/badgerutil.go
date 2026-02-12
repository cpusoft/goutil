package badgerutil

import (
	"errors"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
)

var (
	badgerDB    *badger.DB
	initialized uint32
	batchSize   = 1000 // 批量写入的大小
)

func Init(dbPath string) error {
	if !atomic.CompareAndSwapUint32(&initialized, 0, 1) {
		return errors.New("badgerDB is already initialized")
	}
	var err error
	// 优化配置以适应高并发场景
	opts := badger.DefaultOptions(dbPath)

	opts = opts.WithMemTableSize(256 * 1024 * 1024) // 128MB内存表, <=小于系统内存/4
	opts = opts.WithNumMemtables(runtime.NumCPU())  // cpunum个内存表
	opts = opts.WithValueLogFileSize(1 << 30)       // 1G日志文件
	opts = opts.WithNumCompactors(runtime.NumCPU()) // 增加压缩器数量
	opts = opts.WithSyncWrites(false)               // 关闭同步写，提升性能
	opts = opts.WithCompression(options.None)       // 关闭压缩，减少CPU开销
	opts = opts.WithNumLevelZeroTables(10)          // 增大L0表阈值，减少压缩触发
	opts = opts.WithNumLevelZeroTablesStall(20)     // 增大stall阈值，避免写阻塞

	opts = opts.WithNumGoroutines(runtime.NumCPU()) // 增加并发goroutine数量
	opts = opts.WithBlockCacheSize(100 << 20)       // 100MB块缓存
	opts = opts.WithIndexCacheSize(50 << 20)        // 50MB索引缓存
	opts = opts.WithValueThreshold(1024)            // 1KB以下的值内联存储

	badgerDB, err = badger.Open(opts)
	if err != nil {
		belogs.Error("Init(): Open with options fail, opts:", opts, err)
		atomic.StoreUint32(&initialized, 0)
		return err
	}

	return nil
}

func Close() {
	if badgerDB != nil {
		badgerDB.RunValueLogGC(0.5)
		badgerDB.Close()
	}
	atomic.StoreUint32(&initialized, 0)
}

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

func BatchUpdateMap[T any](datas map[string]T, expire time.Duration) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}

	batch := badgerDB.NewWriteBatch()
	defer batch.Cancel() // 确保在出错时取消
	expireAt := uint64(0)
	if expire > 0 {
		expireAt = uint64(time.Now().Add(expire).Unix())
	}
	i := 0
	for key, value := range datas {
		valueBytes := jsonutil.MarshalJsonBytes(value)
		if valueBytes == nil {
			belogs.Error("BatchUpdateMap(): MarshalJsonBytes fail, value:", value)
			return errors.New("failed to marshal value to JSON bytes")
		}
		entry := &badger.Entry{
			Key:       []byte(key),
			Value:     valueBytes,
			ExpiresAt: expireAt,
		}
		if err := batch.SetEntry(entry); err != nil {
			belogs.Error("BatchUpdateMap(): SetEntry fail, entry:", entry, err)
			return err
		}
		// 达到批次大小，提交Batch
		if (i+1)%batchSize == 0 {
			if err := batch.Flush(); err != nil {
				belogs.Error("BatchUpdateMap(): Flush every batchSize fail, entry:", entry, err)
				return err
			}
			batch.Cancel()                   // 释放当前批次资源
			batch = badgerDB.NewWriteBatch() // 创建新批次
		}
		i++
	}
	// 刷新批量写入
	return batch.Flush()
}

func BatchUpdateKeyFunc[T any](datas []T, expire time.Duration, keyFunc func(T) string) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}

	batch := badgerDB.NewWriteBatch()
	defer batch.Cancel() // 确保在出错时取消
	expireAt := uint64(0)
	if expire > 0 {
		expireAt = uint64(time.Now().Add(expire).Unix())
	}
	for i, value := range datas {
		key := keyFunc(value)
		valueBytes := jsonutil.MarshalJsonBytes(value)
		if valueBytes == nil {
			belogs.Error("BatchUpdateKeyFunc(): MarshalJsonBytes fail, value:", value)
			return errors.New("failed to marshal value to JSON bytes")
		}
		entry := &badger.Entry{
			Key:       []byte(key),
			Value:     valueBytes,
			ExpiresAt: expireAt,
		}
		if err := batch.SetEntry(entry); err != nil {
			belogs.Error("BatchUpdateKeyFunc(): SetEntry fail, entry:", jsonutil.MarshalJson(entry), err)
			return err
		}
		// 达到批次大小，提交Batch
		if (i+1)%batchSize == 0 {
			if err := batch.Flush(); err != nil {
				belogs.Error("BatchUpdateKeyFunc(): Flush every batchSize fail, entry:", jsonutil.MarshalJson(entry), err)
				return err
			}
			batch.Cancel()                   // 释放当前批次资源
			batch = badgerDB.NewWriteBatch() // 创建新批次
		}
	}
	// 刷新批量写入
	return batch.Flush()
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
			belogs.Error("View(): ValueCopy fail, item:", jsonutil.MarshalJson(item), err)
			return err
		}
		value = val
		return nil
	})

	if value == nil {
		return zero, false, err
	}
	var result T
	err = jsonutil.UnmarshalJsonBytes(value, &result)
	if err != nil {
		belogs.Error("View(): UnmarshalJsonBytes fail, value:", string(value), err)
		return zero, false, err
	}
	return result, true, err
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
				belogs.Error("PrefixView(): ValueCopy fail, item:", jsonutil.MarshalJson(item), err)
				return err
			}

			var result T
			err = jsonutil.UnmarshalJsonBytes(val, &result)
			if err != nil {
				belogs.Error("PrefixView(): UnmarshalJsonBytes fail, value:", string(val), err)
				return err
			}
			results = append(results, result)
			count++
		}
		return nil
	})
	return results, err
}
func Delete(key string) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}
	return badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}
