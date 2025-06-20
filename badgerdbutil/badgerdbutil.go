package badgerdbutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	badgedb "github.com/cpusoft/goutil/badgerdbutil/badgerdb"
	"github.com/cpusoft/goutil/belogs"
	"github.com/dgraph-io/badger/v4"
)

var (
	pool        *badgedb.Pool
	initialized uint32
)

var (
	ErrDBNotInitialized = errors.New("database not initialized")
	ErrDBAlreadyInit    = errors.New("database already initialized")
)

// InitDB 初始化数据库
func InitDB(opts badger.Options) error {
	if !atomic.CompareAndSwapUint32(&initialized, 0, 1) {
		return ErrDBAlreadyInit
	}

	var err error
	pool, err = badgedb.NewPool(opts)
	if err != nil {
		atomic.StoreUint32(&initialized, 0)
		return err
	}
	return nil
}

// CloseDB 关闭数据库
func CloseDB() error {
	if !atomic.CompareAndSwapUint32(&initialized, 1, 0) {
		return ErrDBNotInitialized
	}

	if pool != nil {
		err := pool.Close()
		pool = nil
		return err
	}

	return nil
}

// GetPoolStats 获取连接池统计信息
func GetPoolStats() (badgedb.PoolStats, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return badgedb.PoolStats{}, ErrDBNotInitialized
	}
	return pool.Stats(), nil
}

// Insert inserts a key-value pair into the database
func Insert[T any](key string, value T) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.Insert(key, value)
	})
}

func InsertWithTxn[T any](txn *badger.Txn, key string, value T) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.InsertWithTxn(txn, key, value)
	})
}

// Delete deletes a key-value pair from the database
func Delete(key string) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.Delete(key)
	})
}

func DeleteWithTxn(txn *badger.Txn, key string) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.DeleteWithTxn(txn, key)
	})
}

// BatchInsert inserts multiple key-value pairs into the database
func BatchInsert[T any](data map[string]T) map[string]error {
	if atomic.LoadUint32(&initialized) == 0 {
		return map[string]error{"error": ErrDBNotInitialized}
	}

	var result map[string]error
	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		convertedData := make(map[string]any, len(data))
		for k, v := range data {
			convertedData[k] = v
		}
		result = db.BatchInsert(convertedData)
		return nil
	})

	if err != nil {
		return map[string]error{"error": err}
	}
	return result
}

// BatchDel delete multiple key-value pairs from the database
func BatchDelete[T any](data map[string]T) map[string]error {
	if atomic.LoadUint32(&initialized) == 0 {
		return map[string]error{"error": ErrDBNotInitialized}
	}

	var result map[string]error
	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		for k, _ := range data {
			//convertedData[k] = v
			err := db.Delete(k)
			if err != nil {
				belogs.Error("BatchDelete failed, k:", k, " err:", err)
				result[k] = err
			}
		}
		return nil
	})

	if err != nil {
		return map[string]error{"error": err}
	}
	return result
}

// Get retrieves a value from the database by key
func Get[T any](key string) (T, bool, error) {
	var result T
	var exists bool

	if atomic.LoadUint32(&initialized) == 0 {
		return result, false, ErrDBNotInitialized
	}

	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		value, e, err := db.Get(key)
		if err != nil {
			return err
		}

		exists = e
		if !exists {
			return nil
		}

		return badgedb.UnMarshalValue[T](value, &result)
	})

	if err != nil {
		var zero T
		return zero, false, err
	}

	return result, exists, nil
}

func GetWithTxn[T any](txn *badger.Txn, key string) (T, bool, error) {
	var result T
	var exists bool

	if atomic.LoadUint32(&initialized) == 0 {
		return result, false, ErrDBNotInitialized
	}

	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		value, e, err := db.GetWithTxn(txn, key)
		if err != nil {
			return err
		}

		exists = e
		if !exists {
			return nil
		}

		return badgedb.UnMarshalValue[T](value, &result)
	})

	if err != nil {
		var zero T
		return zero, false, err
	}

	return result, exists, nil
}

// MGet retrieves multiple values from the database by keys
func MGet[T any](keys []string) (map[string]T, map[string]error, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return nil, nil, ErrDBNotInitialized
	}

	var typedResults map[string]T
	var resultErrors map[string]error

	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		results, errors, err := db.MGet(keys)
		if err != nil {
			return err
		}

		typedResults = make(map[string]T)
		for key, value := range results {
			var typedValue T
			if err := badgedb.UnMarshalValue[T](value, &typedValue); err != nil {
				return fmt.Errorf("type assertion to %T failed for key %s", typedValue, key)
			}
			typedResults[key] = typedValue
		}
		resultErrors = errors
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return typedResults, resultErrors, nil
}

func MGetWithTxn[T any](txn *badger.Txn, keys []string) (map[string]T, map[string]error, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return nil, nil, ErrDBNotInitialized
	}

	var typedResults map[string]T
	var resultErrors map[string]error

	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		results, errors, err := db.MGetWithTxn(txn, keys)
		if err != nil {
			return err
		}

		typedResults = make(map[string]T)
		for key, value := range results {
			var typedValue T
			if err := badgedb.UnMarshalValue[T](value, &typedValue); err != nil {
				return fmt.Errorf("type assertion to %T failed for key %s", typedValue, key)
			}
			typedResults[key] = typedValue
		}
		resultErrors = errors
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return typedResults, resultErrors, nil
}

// RunWithTxn runs a function within a transaction
func RunWithTxn(txnFunc func(txn *badger.Txn) error) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConnTxn(txnFunc)
}

// Append appends a value to a list stored at key
func Append[T any](key string, value T) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.Append(key, value)
	})
}

func AppendWithTxn[T any](txn *badger.Txn, key string, value T) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.AppendWithTxn(txn, key, value)
	})
}

// GetList retrieves a list stored at key
func GetList[T any](key string) ([]T, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return nil, ErrDBNotInitialized
	}

	var result []T
	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		anyList, err := db.GetList(key)
		if err != nil {
			return err
		}

		result = make([]T, len(anyList))
		for i, v := range anyList {
			var typedValue T
			if err := badgedb.UnMarshalValue[T](v, &typedValue); err != nil {
				return fmt.Errorf("type assertion to %T failed for element at index %d", typedValue, i)
			}
			result[i] = typedValue
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetListWithTxn[T any](txn *badger.Txn, key string) ([]T, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return nil, ErrDBNotInitialized
	}

	var result []T
	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		anyList, err := db.GetListWithTxn(txn, key)
		if err != nil {
			return err
		}

		result = make([]T, len(anyList))
		for i, v := range anyList {
			var typedValue T
			if err := badgedb.UnMarshalValue[T](v, &typedValue); err != nil {
				return fmt.Errorf("type assertion to %T failed for element at index %d", typedValue, i)
			}
			result[i] = typedValue
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// StoreWithCompositeKey stores data with a composite key
func StoreWithCompositeKey(entity string, id string, columns map[string]string) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.StoreWithCompositeKey(entity, id, columns)
	})
}

func StoreWithCompositeKeyWithTxn(txn *badger.Txn, entity string, id string, columns map[string]string) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.StoreWithCompositeKeyWithTxn(txn, entity, id, columns)
	})
}

// QueryByCompositeKey queries data by composite key
func QueryByCompositeKey[T any](entity string, columns map[string]string) (T, bool, error) {
	var result T
	var exists bool

	if atomic.LoadUint32(&initialized) == 0 {
		return result, false, ErrDBNotInitialized
	}

	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		value, e, err := db.QueryByCompositeKey(entity, columns)
		if err != nil {
			return err
		}

		exists = e
		if !exists {
			return nil
		}

		return badgedb.UnMarshalValue[T](value, &result)
	})

	if err != nil {
		var zero T
		return zero, exists, err
	}

	return result, exists, nil
}

func QueryByCompositeKeyWithTxn[T any](txn *badger.Txn, entity string, columns map[string]string) (T, bool, error) {
	var result T
	var exists bool

	if atomic.LoadUint32(&initialized) == 0 {
		return result, false, ErrDBNotInitialized
	}

	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		value, e, err := db.QueryByCompositeKeyWithTxn(txn, entity, columns)
		if err != nil {
			return err
		}

		exists = e
		if !exists {
			return nil
		}

		return badgedb.UnMarshalValue[T](value, &result)
	})

	if err != nil {
		var zero T
		return zero, exists, err
	}

	return result, exists, nil
}

// BatchQueryByPrefix queries multiple items by prefix
func BatchQueryByPrefix[T any](prefix string) (map[string]T, map[string]error, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return nil, nil, ErrDBNotInitialized
	}

	var typedResults map[string]T
	var resultErrors map[string]error

	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		anyMap, errors, err := db.BatchQueryByPrefix(prefix)
		if err != nil {
			return err
		}

		typedResults = make(map[string]T)
		for key, v := range anyMap {
			var typedValue T
			if err := badgedb.UnMarshalValue[T](v, &typedValue); err != nil {
				return fmt.Errorf("type assertion to %T failed for element at index %s", typedValue, key)
			}
			typedResults[key] = typedValue
		}
		resultErrors = errors
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return typedResults, resultErrors, nil
}

func BatchQueryByPrefixWithTxn[T any](txn *badger.Txn, prefix string) (map[string]T, map[string]error, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return nil, nil, ErrDBNotInitialized
	}

	var typedResults map[string]T
	var resultErrors map[string]error

	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		anyMap, errors, err := db.BatchQueryByPrefixWithTxn(txn, prefix)
		if err != nil {
			return err
		}

		typedResults = make(map[string]T)
		for key, v := range anyMap {
			var typedValue T
			if err := badgedb.UnMarshalValue[T](v, &typedValue); err != nil {
				return fmt.Errorf("type assertion to %T failed for element at index %s", typedValue, key)
			}
			typedResults[key] = typedValue
		}
		resultErrors = errors
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return typedResults, resultErrors, nil
}

// HSet sets a field in a hash stored at key
func HSet[T any](key, field string, value T) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.HSet(key, field, value)
	})
}

// HGet retrieves a field from a hash stored at key
func HGet[T any](key, field string) (T, error) {
	var result T

	if atomic.LoadUint32(&initialized) == 0 {
		return result, ErrDBNotInitialized
	}

	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		value, err := db.HGet(key, field)
		if err != nil {
			return err
		}

		return badgedb.UnMarshalValue[T](value, &result)
	})

	if err != nil {
		var zero T
		return zero, err
	}

	return result, nil
}

// HDel deletes a field from a hash stored at key
func HDel(key, field string) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.HDel(key, field)
	})
}

// AddToSet adds a value to a set stored at key
func AddToSet[T any](key string, value T) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.AddToSet(key, value)
	})
}

func ClearSet(key string) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.ClearSet(key)
	})
}

func AddToSetWithTxn[T any](txn *badger.Txn, key string, value T) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.AddToSetWithTxn(txn, key, value)
	})
}

// GetSet retrieves a set stored at key
func GetSet[T any](key string) ([]T, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return nil, ErrDBNotInitialized
	}

	var result []T
	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		anyList, err := db.GetSet(key)
		if err != nil {
			return err
		}

		result = make([]T, len(anyList))
		for i, v := range anyList {
			var typedValue T
			if err := badgedb.UnMarshalValue[T](v, &typedValue); err != nil {
				return fmt.Errorf("type assertion to %T failed for element at index %d", typedValue, i)
			}
			result[i] = typedValue
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetSetWithTxn[T any](txn *badger.Txn, key string) ([]T, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return nil, ErrDBNotInitialized
	}

	var result []T
	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		anyList, err := db.GetSetWithTxn(txn, key)
		if err != nil {
			return err
		}

		result = make([]T, len(anyList))
		for i, v := range anyList {
			var typedValue T
			if err := badgedb.UnMarshalValue[T](v, &typedValue); err != nil {
				return fmt.Errorf("type assertion to %T failed for element at index %d", typedValue, i)
			}
			result[i] = typedValue
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DefaultQueryOptions 返回默认的查询选项
func DefaultQueryOptions() badgedb.QueryOptions {
	return badgedb.QueryOptions{
		PrefetchValues: true,
		PrefetchSize:   100, // 默认预取100条记录
		Reverse:        false,
	}
}

// BatchQueryByPrefixStream 提供流式查询接口
func BatchQueryByPrefixStream(prefix string, opts badgedb.QueryOptions, callback func(key []byte, value []byte) error) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}

	return pool.WithConn(func(db badgedb.BadgeDB) error {
		iterator, err := db.BatchQueryByPrefixWithOptions(prefix, opts)
		if err != nil {
			return err
		}
		defer iterator.Close()

		for iterator.Valid() {
			key := iterator.Key()
			value, err := iterator.Value()
			if err != nil {
				return err
			}

			if err := callback(key, value); err != nil {
				return err
			}

			iterator.Next()
		}

		return nil
	})
}

// BatchQueryByPrefixStreamWithTxn 提供带事务的流式查询接口
func BatchQueryByPrefixStreamWithTxn[T any](txn *badger.Txn, prefix string,
	opts badgedb.QueryOptions, callback func(key string, value T) error) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}

	return pool.WithConn(func(db badgedb.BadgeDB) error {
		iterator, err := db.BatchQueryByPrefixWithTxnAndOptions(txn, prefix, opts)
		if err != nil {
			return err
		}
		defer iterator.Close()

		for iterator.Valid() {
			key := iterator.Key()
			valueBytes, err := iterator.Value()
			if err != nil {
				return err
			}

			var value T
			if err := badgedb.UnMarshalValue(valueBytes, &value); err != nil {
				return err
			}

			if err := callback(string(key), value); err != nil {
				return err
			}

			iterator.Next()
		}

		return nil
	})
}

// BatchQueryByPrefixStreamWithBatch 提供带批处理的流式查询接口
func BatchQueryByPrefixStreamWithBatch[T any](prefix string, batchSize int,
	callback func(batch []T) error) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}

	return pool.WithConn(func(db badgedb.BadgeDB) error {
		opts := DefaultQueryOptions()
		iterator, err := db.BatchQueryByPrefixWithOptions(prefix, opts)
		if err != nil {
			return err
		}
		defer iterator.Close()

		batch := make([]T, 0, batchSize)
		for iterator.Valid() {
			valueBytes, err := iterator.Value()
			if err != nil {
				return err
			}

			var value T
			if err := badgedb.UnMarshalValue(valueBytes, &value); err != nil {
				// 如果解析失败，尝试直接使用原始值
				var ok bool
				switch any((*new(T))).(type) {
				case int:
					// 对于整数类型，尝试解析数字
					i, err := strconv.Atoi(strings.Trim(string(valueBytes), "\""))
					if err != nil {
						return err
					}
					value, ok = any(i).(T)
					if !ok {
						return fmt.Errorf("failed to convert %v to target type", i)
					}
				case string:
					// 对于字符串类型，直接使用
					s := strings.Trim(string(valueBytes), "\"")
					value, ok = any(s).(T)
					if !ok {
						return fmt.Errorf("failed to convert %v to target type", s)
					}
				default:
					// 其他类型，返回原始错误
					return err
				}
			}

			batch = append(batch, value)

			// 当批次达到指定大小时，处理批次
			if len(batch) >= batchSize {
				if err := callback(batch); err != nil {
					return err
				}
				batch = batch[:0] // 清空批次
			}

			// 移动到下一个元素
			iterator.Next()
		}

		// 处理最后一个不完整的批次
		if len(batch) > 0 {
			if err := callback(batch); err != nil {
				return err
			}
		}

		return nil
	})
}

// BatchAddToSet adds multiple values to a set stored at key
func BatchAddToSet[T any](key string, values []T) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	anyValues := make([]any, len(values))
	for i, v := range values {
		anyValues[i] = v
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.BatchAddToSet(key, anyValues)
	})
}

// BatchAddToSetWithTxn adds multiple values to a set stored at key within a transaction
func BatchAddToSetWithTxn[T any](txn *badger.Txn, key string, values []T) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	anyValues := make([]any, len(values))
	for i, v := range values {
		anyValues[i] = v
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.BatchAddToSetWithTxn(txn, key, anyValues)
	})
}

// BatchRemoveFromSet removes multiple values from a set stored at key
func BatchRemoveFromSet[T any](key string, values []T) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	anyValues := make([]any, len(values))
	for i, v := range values {
		anyValues[i] = v
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.BatchRemoveFromSet(key, anyValues)
	})
}

// BatchRemoveFromSetWithTxn removes multiple values from a set stored at key within a transaction
func BatchRemoveFromSetWithTxn[T any](txn *badger.Txn, key string, values []T) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	anyValues := make([]any, len(values))
	for i, v := range values {
		anyValues[i] = v
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.BatchRemoveFromSetWithTxn(txn, key, anyValues)
	})
}

// GetSetWithPrefix retrieves all sets with a given prefix
func GetSetWithPrefix[T any](prefix string) (map[string][]T, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return nil, ErrDBNotInitialized
	}

	var result map[string][]T
	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		anyMap, err := db.GetSetWithPrefix(prefix)
		if err != nil {
			return err
		}

		result = make(map[string][]T)
		for key, values := range anyMap {
			typedValues := make([]T, len(values))
			for i, v := range values {
				var typedValue T
				if err := badgedb.UnMarshalValue[T](v, &typedValue); err != nil {
					return fmt.Errorf("type assertion to %T failed for element at index %d in set %s", typedValue, i, key)
				}
				typedValues[i] = typedValue
			}
			result[key] = typedValues
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetSetWithPrefixTxn retrieves all sets with a given prefix within a transaction
func GetSetWithPrefixTxn[T any](txn *badger.Txn, prefix string) (map[string][]T, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return nil, ErrDBNotInitialized
	}

	var result map[string][]T
	err := pool.WithConn(func(db badgedb.BadgeDB) error {
		anyMap, err := db.GetSetWithPrefixTxn(txn, prefix)
		if err != nil {
			return err
		}

		result = make(map[string][]T)
		for key, values := range anyMap {
			typedValues := make([]T, len(values))
			for i, v := range values {
				var typedValue T
				if err := badgedb.UnMarshalValue[T](v, &typedValue); err != nil {
					return fmt.Errorf("type assertion to %T failed for element at index %d in set %s", typedValue, i, key)
				}
				typedValues[i] = typedValue
			}
			result[key] = typedValues
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// ClearSetWithPrefixBatched 分批删除指定前缀的集合
// batchSize 参数控制每个事务中处理的最大键数
func ClearSetWithPrefixBatched(prefix string, batchSize int) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.ClearSetWithPrefixFast(prefix)
	})
}

// ClearSetWithPrefixFast 使用DropPrefix方法快速删除指定前缀的所有键
func ClearSetWithPrefixFast(prefix string) error {
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrDBNotInitialized
	}
	return pool.WithConn(func(db badgedb.BadgeDB) error {
		return db.ClearSetWithPrefixFast(prefix)
	})
}
