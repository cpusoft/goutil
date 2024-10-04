package badgedb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/dgraph-io/badger/v4"
)

var (
	globalDB BadgeDB
	once     sync.Once
	mu       sync.RWMutex
)

var ErrDBNotInitialized = errors.New("database not initialized")

// InitDB initializes the database
func InitDB(options badger.Options) error {
	var err error
	once.Do(func() {
		globalDB = NewBadgeDB()
		err = globalDB.Init(options)
	})
	return err
}

// CloseDB closes the database connection
func CloseDB() error {
	mu.Lock()
	defer mu.Unlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}
	return globalDB.Close()
}

// RunWithTxn runs a function within a transaction
func RunWithTxn(txnFunc func(txn *badger.Txn) error) error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}
	return globalDB.RunWithTxn(txnFunc)
}

// Insert inserts a key-value pair into the database
func Insert[T any](key string, value T) error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}

	return globalDB.Insert(key, value)
}

func InsertWithTxn[T any](txn *badger.Txn, key string, value T) error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}
	return globalDB.InsertWithTxn(txn, key, value)
}

// Delete deletes a key-value pair from the database
func Delete(key string) error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}
	return globalDB.Delete(key)
}

func DeleteWithTxn(txn *badger.Txn, key string) error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}
	return globalDB.DeleteWithTxn(txn, key)
}

// BatchInsert inserts multiple key-value pairs into the database
func BatchInsert[T any](data map[string]T) map[string]error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return map[string]error{"error": ErrDBNotInitialized}
	}

	convertedData := make(map[string]any, len(data))
	for k, v := range data {
		convertedData[k] = v
	}
	return globalDB.BatchInsert(convertedData)
}

// Get retrieves a value from the database by key
func Get[T any](key string) (T, bool, error) {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		var zero T
		return zero, false, ErrDBNotInitialized
	}
	value, exists, err := globalDB.Get(key)
	if err != nil {
		var zero T
		return zero, false, err
	}

	if !exists {
		var zero T
		return zero, false, nil
	}

	var typedValue T
	err = unmarshalValue[T](value, &typedValue)

	if err != nil {
		var zero T
		return zero, false, fmt.Errorf("type assertion to %T failed", typedValue)
	}

	return typedValue, exists, nil
}

func GetWithTxn[T any](txn *badger.Txn, key string) (T, bool, error) {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		var zero T
		return zero, false, ErrDBNotInitialized
	}
	value, exists, err := globalDB.GetWithTxn(txn, key)
	if err != nil {
		var zero T
		return zero, false, err
	}

	if !exists {
		var zero T
		return zero, false, nil
	}
	var typedValue T
	err = unmarshalValue[T](value, &typedValue)

	if err != nil {
		var zero T
		return zero, false, fmt.Errorf("type assertion to %T failed", typedValue)
	}

	return typedValue, exists, nil
}

// MGet retrieves multiple values from the database by keys
func MGet[T any](keys []string) (map[string]T, map[string]error, error) {
	results, errors, err := globalDB.MGet(keys)
	if err != nil {
		return nil, nil, err
	}

	typedResults := make(map[string]T)
	for key, value := range results {

		var typedValue T
		err = unmarshalValue[T](value, &typedValue)

		if err != nil {
			return nil, nil, fmt.Errorf("type assertion to %T failed for key %s", typedValue, key)
		}
		typedResults[key] = typedValue
	}

	return typedResults, errors, nil
}

// Append appends a value to a list stored at key

func Append[T any](key string, value T) error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}
	return globalDB.Append(key, value)
}

func AppendWithTxn[T any](txn *badger.Txn, key string, value T) error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}
	return globalDB.AppendWithTxn(txn, key, value)
}

// GetList retrieves a list stored at key
func GetList[T any](key string) ([]T, error) {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return nil, ErrDBNotInitialized
	}

	anyList, err := globalDB.GetList(key)
	if err != nil {
		return nil, err
	}

	typedList := make([]T, len(anyList))
	for i, v := range anyList {
		var typedValue T
		err = unmarshalValue[T](v, &typedValue)

		if err != nil {
			return nil, fmt.Errorf("type assertion to %T failed for element at index %d", typedValue, i)
		}
		typedList[i] = typedValue
	}

	return typedList, nil
}

func GetListWithTxn[T any](txn *badger.Txn, key string) ([]T, error) {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return nil, ErrDBNotInitialized
	}

	anyList, err := globalDB.GetListWithTxn(txn, key)
	if err != nil {
		return nil, err
	}

	typedList := make([]T, len(anyList))
	for i, v := range anyList {
		var typedValue T
		err = unmarshalValue[T](v, &typedValue)

		if err != nil {
			return nil, fmt.Errorf("type assertion to %T failed for element at index %d", typedValue, i)
		}
		typedList[i] = typedValue
	}

	return typedList, nil
}

// StoreWithCompositeKey stores data with a composite key
func StoreWithCompositeKey(entity string, id string, columns map[string]string) error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}
	return globalDB.StoreWithCompositeKey(entity, id, columns)
}
func StoreWithCompositeKeyWithTxn(txn *badger.Txn, entity string, id string, columns map[string]string) error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}
	return globalDB.StoreWithCompositeKeyWithTxn(txn, entity, id, columns)
}

// QueryByCompositeKey queries data by composite key
func QueryByCompositeKey[T any](entity string, columns map[string]string) (T, bool, error) {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		var zeroV T
		return zeroV, false, ErrDBNotInitialized
	}
	var typedValue T
	value, exist, err := globalDB.QueryByCompositeKey(entity, columns)
	if err != nil {
		return typedValue, exist, err
	}

	err = unmarshalValue[T](value, &typedValue)

	if err != nil {
		var zero T
		return zero, exist, fmt.Errorf("type assertion to %T failed", typedValue)
	}
	return typedValue, exist, nil
}

func QueryByCompositeKeyWithTxn[T any](txn *badger.Txn, entity string, columns map[string]string) (T, bool, error) {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		var zeroV T
		return zeroV, false, ErrDBNotInitialized
	}
	var typedValue T
	value, exist, err := globalDB.QueryByCompositeKeyWithTxn(txn, entity, columns)
	if err != nil {
		return typedValue, exist, err
	}

	err = unmarshalValue[T](value, &typedValue)

	if err != nil {
		var zero T
		return zero, exist, fmt.Errorf("type assertion to %T failed", typedValue)
	}
	return typedValue, exist, nil
}

// BatchQueryByPrefix queries multiple items by prefix
func BatchQueryByPrefix[T any](prefix string) (map[string]T, map[string]error, error) {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return nil, nil, ErrDBNotInitialized
	}
	anyMap, errMap, err := globalDB.BatchQueryByPrefix(prefix)
	if err != nil {
		return nil, nil, err
	}

	typedList := make(map[string]T, len(anyMap))
	for key, v := range anyMap {
		var typedValue T
		err = unmarshalValue[T](v, &typedValue)

		if err != nil {
			return nil, nil, fmt.Errorf("type assertion to %T failed for element at index %s", typedValue, key)
		}
		typedList[key] = typedValue
	}
	return typedList, errMap, nil
}

func BatchQueryByPrefixWithTxn[T any](txn *badger.Txn, prefix string) (map[string]T, map[string]error, error) {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return nil, nil, ErrDBNotInitialized
	}
	anyMap, errMap, err := globalDB.BatchQueryByPrefixWithTxn(txn, prefix)
	if err != nil {
		return nil, nil, err
	}

	typedList := make(map[string]T, len(anyMap))
	for key, v := range anyMap {
		var typedValue T
		err = unmarshalValue[T](v, &typedValue)

		if err != nil {
			return nil, nil, fmt.Errorf("type assertion to %T failed for element at index %s", typedValue, key)
		}
		typedList[key] = typedValue
	}
	return typedList, errMap, nil
}

// HSet sets a field in a hash stored at key
func HSet[T any](key, field string, value T) error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}
	return globalDB.HSet(key, field, value)
}

// HGet retrieves a field from a hash stored at key
func HGet[T any](key, field string) (T, error) {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		var zero T
		return zero, ErrDBNotInitialized
	}

	value, err := globalDB.HGet(key, field)
	if err != nil {
		var zero T
		return zero, nil

	}

	var typedValue T
	err = unmarshalValue[T](value, &typedValue)

	if err != nil {
		var zero T
		return zero, fmt.Errorf("assert failed, value:%v", value)
	}

	return typedValue, nil
}

// HDel deletes a field from a hash stored at key
func HDel(key, field string) error {
	mu.RLock()
	defer mu.RUnlock()
	if globalDB == nil {
		return ErrDBNotInitialized
	}
	return globalDB.HDel(key, field)
}

func marshalValue(v any) ([]byte, error) {
	// Use JSON serialization to preserve type information
	return json.Marshal(v)
}

func unmarshalValue[T any](data []byte, v *T) error {
	switch any(v).(type) {
	case *string:
		*v = any(string(data)).(T) // 处理为字符串
	case *int:
		// 尝试将字节数据转换为整数
		i, err := strconv.Atoi(string(data))
		if err != nil {
			return err
		}
		*v = any(i).(T)
	case *float64:
		// 尝试将字节数据转换为浮点数
		f, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return err
		}
		*v = any(f).(T)
	default:
		// 先检查数据是否是有效的 JSON
		if json.Valid(data) {
			return json.Unmarshal(data, v)
		} else {
			// 如果数据不是 JSON 格式，返回错误
			return fmt.Errorf("invalid data format: %s", string(data))
		}
	}
	return nil
}
