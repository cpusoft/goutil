package badgedb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/dgraph-io/badger/v4"
)

var (
	initialized uint32
	dbPool      sync.Pool
	options     atomic.Pointer[badger.Options]
)

var (
	ErrDBNotInitialized = errors.New("database not initialized")
	ErrDBAlreadyInit    = errors.New("database already initialized")
)

type DBManager struct {
	db BadgeDB
}

func InitDB(opts badger.Options) error {
	if !atomic.CompareAndSwapUint32(&initialized, 0, 1) {
		return ErrDBAlreadyInit
	}

	optsCopy := opts
	options.Store(&optsCopy)

	// 初始化连接池
	dbPool.New = func() interface{} {
		db := NewBadgeDB()
		err := db.Init(opts)
		if err != nil {
			return nil
		}
		return &DBManager{db: db}
	}

	// 预热连接池
	warmupSize := 4 // 可以根据需要调整
	for i := 0; i < warmupSize; i++ {
		manager := dbPool.Get().(*DBManager)
		if manager != nil {
			dbPool.Put(manager)
		}
	}

	return nil
}

func getManager() (*DBManager, error) {
	if atomic.LoadUint32(&initialized) == 0 {
		return nil, ErrDBNotInitialized
	}

	manager := dbPool.Get().(*DBManager)
	if manager == nil {
		opts := options.Load()
		if opts == nil {
			return nil, ErrDBNotInitialized
		}

		db := NewBadgeDB()
		if err := db.Init(*opts); err != nil {
			return nil, err
		}
		manager = &DBManager{db: db}
	}
	return manager, nil
}

func releaseManager(manager *DBManager) {
	if manager != nil {
		dbPool.Put(manager)
	}
}

func CloseDB() error {
	if !atomic.CompareAndSwapUint32(&initialized, 1, 0) {
		return ErrDBNotInitialized
	}

	// 清理连接池中的所有连接
	for {
		manager := dbPool.Get().(*DBManager)
		if manager == nil {
			break
		}
		if err := manager.db.Close(); err != nil {
			return err
		}
	}

	return nil
}

func withDB(op func(BadgeDB) error) error {
	manager, err := getManager()
	if err != nil {
		return err
	}
	defer releaseManager(manager)

	return op(manager.db)
}

func withDBResult[T any](op func(BadgeDB) (T, error)) (T, error) {
	manager, err := getManager()
	if err != nil {
		var zero T
		return zero, err
	}
	defer releaseManager(manager)

	return op(manager.db)
}

// Insert inserts a key-value pair into the database
func Insert[T any](key string, value T) error {
	return withDB(func(db BadgeDB) error {
		return db.Insert(key, value)
	})
}

func InsertWithTxn[T any](txn *badger.Txn, key string, value T) error {
	return withDB(func(db BadgeDB) error {
		return db.InsertWithTxn(txn, key, value)
	})
}

// Delete deletes a key-value pair from the database
func Delete(key string) error {
	return withDB(func(db BadgeDB) error {
		return db.Delete(key)
	})
}

func DeleteWithTxn(txn *badger.Txn, key string) error {
	return withDB(func(db BadgeDB) error {
		return db.DeleteWithTxn(txn, key)
	})
}

// BatchInsert inserts multiple key-value pairs into the database
func BatchInsert[T any](data map[string]T) map[string]error {
	manager, err := getManager()
	if err != nil {
		return map[string]error{"error": err}
	}
	defer releaseManager(manager)

	convertedData := make(map[string]any, len(data))
	for k, v := range data {
		convertedData[k] = v
	}
	return manager.db.BatchInsert(convertedData)
}

// Get retrieves a value from the database by key
func Get[T any](key string) (T, bool, error) {
	var result T
	var exists bool

	err := withDB(func(db BadgeDB) error {
		value, e, err := db.Get(key)
		if err != nil {
			return err
		}

		exists = e
		if !exists {
			return nil
		}

		return unmarshalValue[T](value, &result)
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

	err := withDB(func(db BadgeDB) error {
		value, e, err := db.GetWithTxn(txn, key)
		if err != nil {
			return err
		}

		exists = e
		if !exists {
			return nil
		}

		return unmarshalValue[T](value, &result)
	})

	if err != nil {
		var zero T
		return zero, false, err
	}

	return result, exists, nil
}

// MGet retrieves multiple values from the database by keys
func MGet[T any](keys []string) (map[string]T, map[string]error, error) {
	manager, err := getManager()
	if err != nil {
		return nil, nil, err
	}
	defer releaseManager(manager)

	results, errors, err := manager.db.MGet(keys)
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

func MGetWithTxn[T any](txn *badger.Txn, keys []string) (map[string]T, map[string]error, error) {
	manager, err := getManager()
	if err != nil {
		return nil, nil, err
	}
	defer releaseManager(manager)

	results, errors, err := manager.db.MGetWithTxn(txn, keys)
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
	return withDB(func(db BadgeDB) error {
		return db.Append(key, value)
	})
}

func AppendWithTxn[T any](txn *badger.Txn, key string, value T) error {
	return withDB(func(db BadgeDB) error {
		return db.AppendWithTxn(txn, key, value)
	})
}

// GetList retrieves a list stored at key
func GetList[T any](key string) ([]T, error) {
	var result []T

	err := withDB(func(db BadgeDB) error {
		anyList, err := db.GetList(key)
		if err != nil {
			return err
		}

		result = make([]T, len(anyList))
		for i, v := range anyList {
			var typedValue T
			err = unmarshalValue[T](v, &typedValue)
			if err != nil {
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
	var result []T

	err := withDB(func(db BadgeDB) error {
		anyList, err := db.GetListWithTxn(txn, key)
		if err != nil {
			return err
		}

		result = make([]T, len(anyList))
		for i, v := range anyList {
			var typedValue T
			err = unmarshalValue[T](v, &typedValue)
			if err != nil {
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
	return withDB(func(db BadgeDB) error {
		return db.StoreWithCompositeKey(entity, id, columns)
	})
}

func StoreWithCompositeKeyWithTxn(txn *badger.Txn, entity string, id string, columns map[string]string) error {
	return withDB(func(db BadgeDB) error {
		return db.StoreWithCompositeKeyWithTxn(txn, entity, id, columns)
	})
}

// QueryByCompositeKey queries data by composite key
func QueryByCompositeKey[T any](entity string, columns map[string]string) (T, bool, error) {
	var result T
	var exists bool

	err := withDB(func(db BadgeDB) error {
		value, e, err := db.QueryByCompositeKey(entity, columns)
		if err != nil {
			return err
		}

		exists = e
		if !exists {
			return nil
		}

		return unmarshalValue[T](value, &result)
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

	err := withDB(func(db BadgeDB) error {
		value, e, err := db.QueryByCompositeKeyWithTxn(txn, entity, columns)
		if err != nil {
			return err
		}

		exists = e
		if !exists {
			return nil
		}

		return unmarshalValue[T](value, &result)
	})

	if err != nil {
		var zero T
		return zero, exists, err
	}

	return result, exists, nil
}

// BatchQueryByPrefix queries multiple items by prefix
func BatchQueryByPrefix[T any](prefix string) (map[string]T, map[string]error, error) {
	manager, err := getManager()
	if err != nil {
		return nil, nil, err
	}
	defer releaseManager(manager)

	anyMap, errors, err := manager.db.BatchQueryByPrefix(prefix)
	if err != nil {
		return nil, nil, err
	}

	typedResults := make(map[string]T)
	for key, v := range anyMap {
		var typedValue T
		err = unmarshalValue[T](v, &typedValue)
		if err != nil {
			return nil, nil, fmt.Errorf("type assertion to %T failed for element at index %s", typedValue, key)
		}
		typedResults[key] = typedValue
	}
	return typedResults, errors, nil
}

func BatchQueryByPrefixWithTxn[T any](txn *badger.Txn, prefix string) (map[string]T, map[string]error, error) {
	manager, err := getManager()
	if err != nil {
		return nil, nil, err
	}
	defer releaseManager(manager)

	anyMap, errors, err := manager.db.BatchQueryByPrefixWithTxn(txn, prefix)
	if err != nil {
		return nil, nil, err
	}

	typedResults := make(map[string]T)
	for key, v := range anyMap {
		var typedValue T
		err = unmarshalValue[T](v, &typedValue)
		if err != nil {
			return nil, nil, fmt.Errorf("type assertion to %T failed for element at index %s", typedValue, key)
		}
		typedResults[key] = typedValue
	}
	return typedResults, errors, nil
}

// HSet sets a field in a hash stored at key
func HSet[T any](key, field string, value T) error {
	return withDB(func(db BadgeDB) error {
		return db.HSet(key, field, value)
	})
}

// HGet retrieves a field from a hash stored at key
func HGet[T any](key, field string) (T, error) {
	var result T

	err := withDB(func(db BadgeDB) error {
		value, err := db.HGet(key, field)
		if err != nil {
			return nil
		}

		return unmarshalValue[T](value, &result)
	})

	if err != nil {
		var zero T
		return zero, nil
	}

	return result, nil
}

// HDel deletes a field from a hash stored at key
func HDel(key, field string) error {
	return withDB(func(db BadgeDB) error {
		return db.HDel(key, field)
	})
}

func marshalValue(v any) ([]byte, error) {
	return json.Marshal(v)
}

func unmarshalValue[T any](data []byte, v *T) error {
	switch any(v).(type) {
	case *string:
		*v = any(string(data)).(T)
	case *int:
		i, err := strconv.Atoi(string(data))
		if err != nil {
			return err
		}
		*v = any(i).(T)
	case *float64:
		f, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return err
		}
		*v = any(f).(T)
	default:
		if json.Valid(data) {
			return json.Unmarshal(data, v)
		}
		return fmt.Errorf("invalid data format: %s", string(data))
	}
	return nil
}

//

func AddToSet[T any](key string, value T) error {
	return withDB(func(db BadgeDB) error {
		return db.AddToSet(key, value)
	})
}

func AddToSetWithTxn[T any](txn *badger.Txn, key string, value T) error {
	return withDB(func(db BadgeDB) error {
		return db.AddToSetWithTxn(txn, key, value)
	})
}

func GetSet[T any](key string) ([]T, error) {
	var result []T

	err := withDB(func(db BadgeDB) error {
		anyList, err := db.GetSet(key)
		if err != nil {
			return err
		}

		result = make([]T, len(anyList))
		for i, v := range anyList {
			var typedValue T
			err = unmarshalValue[T](v, &typedValue)
			if err != nil {
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
	var result []T

	err := withDB(func(db BadgeDB) error {
		anyList, err := db.GetSetWithTxn(txn, key)
		if err != nil {
			return err
		}

		result = make([]T, len(anyList))
		for i, v := range anyList {
			var typedValue T
			err = unmarshalValue[T](v, &typedValue)
			if err != nil {
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
