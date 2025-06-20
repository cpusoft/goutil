package badgerdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dgraph-io/badger/v4"
)

// QueryOptions 查询选项
type QueryOptions struct {
	PrefetchValues bool // 是否预取值
	PrefetchSize   int  // 预取的数量
	Reverse        bool // 是否反向遍历
}

// Iterator 迭代器接口
type Iterator interface {
	Next() bool
	Key() []byte
	Value() ([]byte, error)
	Close() error
	Valid() bool
}

// BadgeDB interface defines all available database operations
type BadgeDB interface {
	// Initialization methods
	Init(options badger.Options) error
	Close() error

	// General operations
	Insert(key string, value any) error
	InsertWithTxn(txn *badger.Txn, key string, value any) error
	Delete(key string) error
	DeleteWithTxn(txn *badger.Txn, key string) error
	BatchInsert(data map[string]any) map[string]error
	BatchInsertWithTxn(txn *badger.Txn, data map[string]any) error
	Get(key string) ([]byte, bool, error)
	GetWithTxn(txn *badger.Txn, key string) ([]byte, bool, error)
	MGet(keys []string) (map[string][]byte, map[string]error, error)
	MGetWithTxn(txn *badger.Txn, keys []string) (map[string][]byte, map[string]error, error)

	// List operations
	Append(key string, value any) error
	AppendWithTxn(txn *badger.Txn, key string, value any) error
	GetList(key string) ([][]byte, error)
	GetListWithTxn(txn *badger.Txn, key string) ([][]byte, error)
	ClearList(key string) error
	ClearListWithTxn(txn *badger.Txn, key string) error

	// Set operations
	AddToSet(key string, value any) error
	AddToSetWithTxn(txn *badger.Txn, key string, value any) error
	RemoveFromSet(key string, value any) error
	RemoveFromSetWithTxn(txn *badger.Txn, key string, value any) error
	ContainsInSet(key string, value any) (bool, error)
	ContainsInSetWithTxn(txn *badger.Txn, key string, value any) (bool, error)
	GetSet(key string) ([][]byte, error)
	GetSetWithTxn(txn *badger.Txn, key string) ([][]byte, error)
	ClearSet(key string) error
	ClearSetWithTxn(txn *badger.Txn, key string) error

	// New batch set operations
	BatchAddToSet(key string, values []any) error
	BatchAddToSetWithTxn(txn *badger.Txn, key string, values []any) error
	BatchRemoveFromSet(key string, values []any) error
	BatchRemoveFromSetWithTxn(txn *badger.Txn, key string, values []any) error
	GetSetWithPrefix(prefix string) (map[string][][]byte, error)
	GetSetWithPrefixTxn(txn *badger.Txn, prefix string) (map[string][][]byte, error)
	//ClearSetWithPrefix(prefix string) error
	//ClearSetWithPrefixTxn(txn *badger.Txn, prefix string) error
	ClearSetWithPrefixBatched(prefix string, batchSize int) error

	ClearSetWithPrefixFast(prefix string) error

	// Index operations
	BuildColumnIndex(txn *badger.Txn, entityKey string, column string, value any) error
	BuildMultipleColumnIndexes(txn *badger.Txn, entityKey string, columns map[string]interface{}) error
	StoreWithCompositeKey(entity string, id string, columns map[string]string) error
	StoreWithCompositeKeyWithTxn(txn *badger.Txn, entity string, id string, columns map[string]string) error
	QueryByCompositeKey(entity string, columns map[string]string) ([]byte, bool, error)
	QueryByCompositeKeyWithTxn(txn *badger.Txn, entity string, columns map[string]string) ([]byte, bool, error)
	BatchQueryByPrefix(prefix string) (map[string][]byte, map[string]error, error)
	BatchQueryByPrefixWithTxn(txn *badger.Txn, prefix string) (map[string][]byte, map[string]error, error)

	// Hash operations
	HSet(key, field string, value any) error
	HSetWithTxn(txn *badger.Txn, key, field string, value any) error
	HGet(key, field string) ([]byte, error)
	HGetWithTxn(txn *badger.Txn, key, field string) ([]byte, error)
	HDel(key, field string) error
	HDelWithTxn(txn *badger.Txn, key, field string) error

	RunWithTxn(txnFunc func(txn *badger.Txn) error) error

	// 新增的方法
	BatchQueryByPrefixWithOptions(prefix string, opts QueryOptions) (Iterator, error)
	BatchQueryByPrefixWithTxnAndOptions(txn *badger.Txn, prefix string, opts QueryOptions) (Iterator, error)
}

// BadgeDBImpl 是 BadgeDB 接口的实现
type BadgeDBImpl struct {
	db *badger.DB
}

// badgerIterator 是 Iterator 接口的实现
type badgerIterator struct {
	iterator *badger.Iterator
	prefix   []byte
}

func (it *badgerIterator) Next() bool {
	it.iterator.Next()
	return it.iterator.Valid()
}

func (it *badgerIterator) Key() []byte {
	return it.iterator.Item().KeyCopy(nil)
}

func (it *badgerIterator) Value() ([]byte, error) {
	return it.iterator.Item().ValueCopy(nil)
}

func (it *badgerIterator) Close() error {
	it.iterator.Close()
	return nil
}

func (it *badgerIterator) Valid() bool {
	if !it.iterator.Valid() {
		return false
	}

	// 检查当前键是否以指定前缀开头
	key := it.iterator.Item().Key()
	return bytes.HasPrefix(key, it.prefix)
}

// Init initializes the database connection
func (b *BadgeDBImpl) Init(options badger.Options) error {
	db, err := badger.Open(options)
	if err != nil {
		return err
	}
	b.db = db
	return nil
}

// Close closes the database connection
func (b *BadgeDBImpl) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

func (b *BadgeDBImpl) BuildColumnIndex(txn *badger.Txn, entityKey string, column string, value any) error {
	valueStr, err := MarshalValue(value)
	if err != nil {
		return err
	}

	indexKey := fmt.Sprintf("%s:%s:%s", column, string(valueStr), entityKey)

	err = txn.Set([]byte(indexKey), []byte(entityKey))
	if err != nil {
		return err
	}
	return nil
}

func (b *BadgeDBImpl) BuildMultipleColumnIndexes(txn *badger.Txn, entityKey string, columns map[string]interface{}) error {
	for column, value := range columns {
		valueStr, err := MarshalValue(value)
		if err != nil {
			return err
		}
		indexKey := fmt.Sprintf("%s:%s:%s", column, string(valueStr), entityKey)

		// 存储索引
		err = txn.Set([]byte(indexKey), []byte(entityKey))
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BadgeDBImpl) StoreWithCompositeKey(entity string, id string, columns map[string]string) error {

	baseKey := b.buildBaseKey(entity, id)
	err := b.db.Update(func(txn *badger.Txn) error {

		indexKey := b.buildCompositeKey(columns)
		err := txn.Set([]byte(indexKey), []byte(baseKey))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (b *BadgeDBImpl) StoreWithCompositeKeyWithTxn(txn *badger.Txn, entity string, id string, columns map[string]string) error {
	indexKey := b.buildCompositeKey(columns)
	baseKey := b.buildBaseKey(entity, id)
	err := txn.Set([]byte(indexKey), []byte(baseKey))
	if err != nil {
		return err
	}
	return nil
}

// NewBadgeDB creates a new BadgeDBImpl instance
func NewBadgeDBWithDB(db *badger.DB) BadgeDB {
	return &BadgeDBImpl{db: db}
}

func NewBadgeDB() BadgeDB {
	return &BadgeDBImpl{}
}

//	func IterateBadgerDB() error {
//		err := db.View(func(txn *badger.Txn) error {
//			opts := badger.DefaultIteratorOptions
//			//opts.PrefetchValues = true // Whether to prefetch values to improve iteration efficiency
//			it := txn.NewIterator(opts)
//			defer it.Close()
//
//			for it.Rewind(); it.Valid(); it.Next() {
//				item := it.Item()
//				key := item.Key()
//
//				err := item.Value(func(val []byte) error {
//					fmt.Printf("Key: %s, Value: %s\n", key, val)
//					return nil
//				})
//
//				if err != nil {
//					return err
//				}
//			}
//			return nil
//		})
//		return err
//	}
func (b *BadgeDBImpl) RunWithTxn(txnFunc func(txn *badger.Txn) error) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txnFunc(txn)
	})
}

func (b *BadgeDBImpl) BatchQueryByPrefixWithOptions(prefix string, opts QueryOptions) (Iterator, error) {
	iterOpts := badger.DefaultIteratorOptions
	iterOpts.PrefetchValues = opts.PrefetchValues
	iterOpts.PrefetchSize = opts.PrefetchSize
	iterOpts.Reverse = opts.Reverse

	txn := b.db.NewTransaction(false)
	it := txn.NewIterator(iterOpts)

	it.Seek([]byte(prefix))
	iterator := &badgerIterator{
		iterator: it,
		prefix:   []byte(prefix),
	}

	return iterator, nil
}

func (b *BadgeDBImpl) BatchQueryByPrefixWithTxnAndOptions(txn *badger.Txn, prefix string, opts QueryOptions) (Iterator, error) {
	iterOpts := badger.DefaultIteratorOptions
	iterOpts.PrefetchValues = opts.PrefetchValues
	iterOpts.PrefetchSize = opts.PrefetchSize
	iterOpts.Reverse = opts.Reverse

	it := txn.NewIterator(iterOpts)
	it.Seek([]byte(prefix))
	iterator := &badgerIterator{
		iterator: it,
		prefix:   []byte(prefix),
	}

	return iterator, nil
}

func MarshalValue(v any) ([]byte, error) {
	return json.Marshal(v)
}

func UnMarshalValue[T any](data []byte, v *T) error {
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
