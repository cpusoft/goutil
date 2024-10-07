package badgedb

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
)

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
}

type BadgeDBImpl struct {
	db *badger.DB
}

func (b *BadgeDBImpl) BuildColumnIndex(txn *badger.Txn, entityKey string, column string, value any) error {
	valueStr, err := marshalValue(value)
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
		valueStr, err := marshalValue(value)
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
