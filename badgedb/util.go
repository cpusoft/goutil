package badgedb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"
	"unsafe"

	"github.com/dgraph-io/badger/v4"
)

var (
	instance unsafe.Pointer //
)

var ErrDBNotInitialized = errors.New("database not initialized")

type DBManager struct {
	db BadgeDB
}

func NewDBManager(options badger.Options) (*DBManager, error) {
	db := NewBadgeDB()
	if err := db.Init(options); err != nil {
		return nil, err
	}
	return &DBManager{db: db}, nil
}

func GetDB() BadgeDB {
	return (*DBManager)(atomic.LoadPointer(&instance)).db
}

func InitDB(options badger.Options) error {
	manager, err := NewDBManager(options)
	if err != nil {
		return err
	}

	if !atomic.CompareAndSwapPointer(&instance, nil, unsafe.Pointer(manager)) {

		_ = manager.db.Close()
		return errors.New("database already initialized")
	}

	return nil
}

func CloseDB() error {
	manager := (*DBManager)(atomic.LoadPointer(&instance))
	if manager == nil {
		return ErrDBNotInitialized
	}

	err := manager.db.Close()
	if err == nil {
		atomic.StorePointer(&instance, nil)
	}
	return err
}

// Insert inserts a key-value pair into the database
func Insert[T any](key string, value T) error {
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.Insert(key, value)
}

func InsertWithTxn[T any](txn *badger.Txn, key string, value T) error {
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.InsertWithTxn(txn, key, value)
}

// Delete deletes a key-value pair from the database
func Delete(key string) error {
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.Delete(key)
}

func DeleteWithTxn(txn *badger.Txn, key string) error {
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.DeleteWithTxn(txn, key)
}

// BatchInsert inserts multiple key-value pairs into the database
func BatchInsert[T any](data map[string]T) map[string]error {
	db := GetDB()
	if db == nil {
		return map[string]error{"error": ErrDBNotInitialized}
	}

	convertedData := make(map[string]any, len(data))
	for k, v := range data {
		convertedData[k] = v
	}
	return db.BatchInsert(convertedData)
}

// Get retrieves a value from the database by key
func Get[T any](key string) (T, bool, error) {
	db := GetDB()
	if db == nil {
		var zero T
		return zero, false, ErrDBNotInitialized
	}

	value, exists, err := db.Get(key)
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
	db := GetDB()
	if db == nil {
		var zero T
		return zero, false, ErrDBNotInitialized
	}
	value, exists, err := db.GetWithTxn(txn, key)
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
	db := GetDB()
	if db == nil {
		return nil, nil, ErrDBNotInitialized
	}
	results, errors, err := db.MGet(keys)
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
	db := GetDB()
	if db == nil {
		return nil, nil, ErrDBNotInitialized
	}
	results, errors, err := db.MGetWithTxn(txn, keys)
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
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.Append(key, value)
}

func AppendWithTxn[T any](txn *badger.Txn, key string, value T) error {
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.AppendWithTxn(txn, key, value)
}

// GetList retrieves a list stored at key
func GetList[T any](key string) ([]T, error) {
	db := GetDB()
	if db == nil {
		return nil, ErrDBNotInitialized
	}

	anyList, err := db.GetList(key)
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
	db := GetDB()
	if db == nil {
		return nil, ErrDBNotInitialized
	}

	anyList, err := db.GetListWithTxn(txn, key)
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
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.StoreWithCompositeKey(entity, id, columns)
}

func StoreWithCompositeKeyWithTxn(txn *badger.Txn, entity string, id string, columns map[string]string) error {
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.StoreWithCompositeKeyWithTxn(txn, entity, id, columns)
}

// QueryByCompositeKey queries data by composite key
func QueryByCompositeKey[T any](entity string, columns map[string]string) (T, bool, error) {
	db := GetDB()
	if db == nil {
		var zero T
		return zero, false, ErrDBNotInitialized
	}
	value, exists, err := db.QueryByCompositeKey(entity, columns)
	if err != nil {
		var zero T
		return zero, exists, err
	}

	var typedValue T
	err = unmarshalValue[T](value, &typedValue)
	if err != nil {
		var zero T
		return zero, exists, fmt.Errorf("type assertion to %T failed", typedValue)
	}
	return typedValue, exists, nil
}

func QueryByCompositeKeyWithTxn[T any](txn *badger.Txn, entity string, columns map[string]string) (T, bool, error) {
	db := GetDB()
	if db == nil {
		var zero T
		return zero, false, ErrDBNotInitialized
	}
	value, exists, err := db.QueryByCompositeKeyWithTxn(txn, entity, columns)
	if err != nil {
		var zero T
		return zero, exists, err
	}

	var typedValue T
	err = unmarshalValue[T](value, &typedValue)
	if err != nil {
		var zero T
		return zero, exists, fmt.Errorf("type assertion to %T failed", typedValue)
	}
	return typedValue, exists, nil
}

// BatchQueryByPrefix queries multiple items by prefix
func BatchQueryByPrefix[T any](prefix string) (map[string]T, map[string]error, error) {
	db := GetDB()
	if db == nil {
		return nil, nil, ErrDBNotInitialized
	}
	anyMap, errMap, err := db.BatchQueryByPrefix(prefix)
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
	db := GetDB()
	if db == nil {
		return nil, nil, ErrDBNotInitialized
	}
	anyMap, errMap, err := db.BatchQueryByPrefixWithTxn(txn, prefix)
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
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.HSet(key, field, value)
}

// HGet retrieves a field from a hash stored at key
func HGet[T any](key, field string) (T, error) {
	db := GetDB()
	if db == nil {
		var zero T
		return zero, ErrDBNotInitialized
	}

	value, err := db.HGet(key, field)
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
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.HDel(key, field)
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
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.AddToSet(key, value)
}

func AddToSetWithTxn[T any](txn *badger.Txn, key string, value T) error {
	db := GetDB()
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.AddToSetWithTxn(txn, key, value)
}

func GetSet[T any](key string) ([]T, error) {
	db := GetDB()
	if db == nil {
		return nil, ErrDBNotInitialized
	}

	anyList, err := db.GetSet(key)
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

func GetSetWithTxn[T any](txn *badger.Txn, key string) ([]T, error) {
	db := GetDB()
	if db == nil {
		return nil, ErrDBNotInitialized
	}

	anyList, err := db.GetSetWithTxn(txn, key)
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
