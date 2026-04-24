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
		badgerDB = nil
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

////////////////////////////////////////////////////
// base funcs
/////////////////////////////////////////////////////

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
			belogs.Debug("Append(): get valCopy:", string(valCopy))
			// 反序列化为数组
			if err := jsonutil.UnmarshalJsonBytes(valCopy, &values); err != nil {
				belogs.Error("Append(): Unmarshal existing value fail, key:", key, err)
				return err
			}
		}

		// 3. 追加新数据
		values = append(values, value)
		belogs.Debug("Append(): new values:", jsonutil.MarshalJson(values))
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
	belogs.Debug("View(): get value", string(value))
	var result T
	err = jsonutil.UnmarshalJsonBytes(value, &result)
	if err != nil {
		belogs.Error("View(): UnmarshalJsonBytes fail, value:", string(value), err)
		return zero, false, err
	}
	return result, true, err
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

////////////////////////////////////////////////////
// advance funcs
/////////////////////////////////////////////////////

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
func BatchUpdateKeyFuncs[T any](datas []T, expire time.Duration, batchSize int,
	mainKeyFunc func(T) string, outerKeyFunc func(T) []string) error {
	// 1. 基础校验
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}
	// 校验批次大小合法性
	if batchSize <= 0 {
		return errors.New("batchSize must be greater than 0")
	}
	// 校验keyFuncs不能为空
	if mainKeyFunc == nil {
		return errors.New("mainKeyFunc cannot be empty")
	}
	if outerKeyFunc == nil {
		return errors.New("outerKeyFunc cannot be empty")
	}
	// 空数据直接返回
	if len(datas) == 0 {
		return nil
	}

	// 2. 计算过期时间
	expireAt := uint64(0)
	if expire > 0 {
		expireAt = uint64(time.Now().Add(expire).Unix())
	}

	// 3. 创建批量写入对象
	batch := badgerDB.NewWriteBatch()
	defer batch.Cancel() // 函数退出时确保取消未提交的批次

	// 4. 遍历数据，多key写入
	for dataIdx, value := range datas {
		// 序列化value（多个key共用同一个value）
		// mainKey --> ValueBytes
		// outerKey1, outerKey2,... --> mainKey
		mainKey := mainKeyFunc(value)
		mainEntry := &badger.Entry{
			Key:       []byte(mainKey + MAINKEY_TO_VALUE),
			Value:     jsonutil.MarshalJsonBytes(value),
			ExpiresAt: expireAt,
		}
		// mainKey --> value
		if err := batch.SetEntry(mainEntry); err != nil {
			belogs.Error("BatchUpdateKeyFunc(): SetEntry mainEntry fail, mainKey:", mainKey,
				"mainEntry", jsonutil.MarshalJson(mainEntry), err)
			return err
		}
		belogs.Debug("BatchUpdateKeyFunc(): SetEntry mainEntry, mainKey:", mainKey,
			"mainEntry", jsonutil.MarshalJson(mainEntry))

		// outerKeys --> mainKey
		outerKeys := outerKeyFunc(value)
		for _, outerKey := range outerKeys {
			outerEntry := &badger.Entry{
				Key:       []byte(outerKey),
				Value:     []byte(mainKey), //jsonutil.MarshalJsonBytes(mainKey),
				ExpiresAt: expireAt,
			}
			if err := batch.SetEntry(outerEntry); err != nil {
				belogs.Error("BatchUpdateKeyFunc(): range outerKeys, SetEntry outerEntry fail, outerKey:", outerKey,
					"outerEntry", jsonutil.MarshalJson(outerEntry), err)
				return err
			}
			belogs.Debug("BatchUpdateKeyFunc(): range outerKeys, SetEntry outerEntry, outerKey:", outerKey,
				"outerEntry", jsonutil.MarshalJson(outerEntry))
		}
		// mainKey --> []outerKeys
		mainOuterEntry := &badger.Entry{
			Key:       []byte(mainKey + MAINKEY_TO_OUTERKEY),
			Value:     jsonutil.MarshalJsonBytes(outerKeys),
			ExpiresAt: expireAt,
		}
		if err := batch.SetEntry(mainOuterEntry); err != nil {
			belogs.Error("BatchUpdateKeyFunc(): SetEntry mainOuterEntry fail, mainKey:", mainKey,
				"mainOuterEntry", jsonutil.MarshalJson(mainOuterEntry), err)
			return err
		}
		belogs.Debug("BatchUpdateKeyFunc(): SetEntry mainOuterEntry, mainKey:", mainKey,
			"mainOuterEntry", jsonutil.MarshalJson(mainOuterEntry))

		// 达到批次大小，提交并重建批次
		if (dataIdx+1)%batchSize == 0 {
			if err := batch.Flush(); err != nil {
				belogs.Error("BatchUpdateKeyFunc(): Flush every batchSize fail, err:", err)
				return err
			}
			batch.Cancel()                   // 释放旧批次资源
			batch = badgerDB.NewWriteBatch() // 新建批次
		}
	}

	// 5. 刷新最后一批不足batchSize的数据
	if err := batch.Flush(); err != nil {
		belogs.Error("BatchUpdateKeyFunc(): Flush final batch fail, err:", err)
		return err
	}

	return nil
}

// DeleteByOuterKey：通过 outerKey 删除整条数据（含清理所有关联 outerKey）
// 执行流程：
// 1. 根据 outerKey 获取 mainKey
// 2. 删除 mainKey 对应的真实 value 数据
// 3. 遍历并删除所有 value = mainKey 的 outerKey（完整清理）
func DeleteByOuterKey(outerKey string) error {
	if atomic.LoadUint32(&initialized) == 0 || badgerDB == nil {
		return errors.New("badgerDB is not initialized")
	}

	return badgerDB.Update(func(txn *badger.Txn) error {
		// 1. 根据 outerKey 获取 mainKey
		mainKeyItem, err := txn.Get([]byte(outerKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				belogs.Debug("DeleteByOuterKey(): outerKey not found, outerKey:", outerKey)
				return nil
			}
			belogs.Error("DeleteByOuterKey(): get mainKey by outerKey fail, outerKey:", outerKey, err)
			return err
		}

		// 读取 mainKey 字符串
		mainKeyBytes, err := mainKeyItem.ValueCopy(nil)
		if err != nil {
			belogs.Error("DeleteByOuterKey(): ValueCopy mainKey fail, outerKey:", outerKey, err)
			return err
		}
		mainKey := string(mainKeyBytes)

		// 2. 删除 mainKey --> value
		err = txn.Delete([]byte(mainKey + MAINKEY_TO_VALUE))
		if err != nil {
			belogs.Error("DeleteByOuterKey(): delete mainKey+MAINKEY_TO_VALUE fail, mainKey+MAINKEY_TO_VALUE:", mainKey+MAINKEY_TO_VALUE, err)
			return err
		}

		// 2. 根据 mainKey 找到 outerKeys ，然后删除 outerKeys --> mainKey
		outerKeysItems, err := txn.Get([]byte(mainKey + MAINKEY_TO_OUTERKEY))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // 键不存在返回nil，不抛错
			}
			belogs.Error("DeleteByOuterKey(): Get fail, key:", mainKey+MAINKEY_TO_OUTERKEY, err)
			return err
		}
		vals, err := outerKeysItems.ValueCopy(nil)
		if err != nil {
			belogs.Error("DeleteByOuterKey(): ValueCopy fail, outerKeysItems:", outerKeysItems, err)
			return err
		}
		outKeys := make([]string, 0)
		err = jsonutil.UnmarshalJson(string(vals), &outKeys)
		if err != nil {
			belogs.Error("DeleteByOuterKey(): UnmarshalJson outKeys fail, vals:", vals, err)
			return err
		}
		for _, outKey := range outKeys {
			err = txn.Delete([]byte(outKey))
			if err != nil {
				belogs.Error("DeleteByOuterKey(): delete outKey fail, outKey:", outKey, err)
				return err
			}
		}

		// 3. 删除 mainKey-->outKeys,
		err = txn.Delete([]byte(mainKey + MAINKEY_TO_OUTERKEY))
		if err != nil {
			belogs.Error("DeleteByOuterKey(): delete mainKey+MAINKEY_TO_VALUE fail, mainKey+MAINKEY_TO_VALUE:", mainKey, err)
			return err
		}

		belogs.Info("DeleteByOuterKey(): success, outerKey:", outerKey,
			"deleted mainKey:", mainKey, " deete outKeys:", outKeys)
		return nil
	})
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
				belogs.Debug("PrefixView(): UnmarshalJsonBytes list fail, will try single model again, value:", string(val), err)
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
