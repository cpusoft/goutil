package badgerdb

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/dgraph-io/badger/v4"
)

func (b *BadgeDBImpl) AddToSet(key string, value any) error {
	valueBytes, err := b.marshalValueForSet(value)
	if err != nil {
		return err
	}

	return b.db.Update(func(txn *badger.Txn) error {
		setKey := fmt.Sprintf("%s:set:%x", key, valueBytes)
		_, err := txn.Get([]byte(setKey))
		if err == badger.ErrKeyNotFound {

			return txn.Set([]byte(setKey), valueBytes)
		}

		return nil
	})
}

func (b *BadgeDBImpl) AddToSetWithTxn(txn *badger.Txn, key string, value any) error {
	valueBytes, err := b.marshalValueForSet(value)
	if err != nil {
		return err
	}
	setKey := fmt.Sprintf("%s:set:%x", key, valueBytes)
	_, err = txn.Get([]byte(setKey))
	if err == badger.ErrKeyNotFound {
		return txn.Set([]byte(setKey), valueBytes)
	}

	return nil
}

func (b *BadgeDBImpl) RemoveFromSet(key string, value any) error {
	valueBytes, err := b.marshalValueForSet(value)
	if err != nil {
		return err
	}

	return b.db.Update(func(txn *badger.Txn) error {
		setKey := fmt.Sprintf("%s:set:%x", key, valueBytes)

		return txn.Delete([]byte(setKey))
	})
}

func (b *BadgeDBImpl) RemoveFromSetWithTxn(txn *badger.Txn, key string, value any) error {
	valueBytes, err := b.marshalValueForSet(value)
	if err != nil {
		return err
	}

	setKey := fmt.Sprintf("%s:set:%x", key, valueBytes)

	return txn.Delete([]byte(setKey))
}

func (b *BadgeDBImpl) ContainsInSet(key string, value any) (bool, error) {
	valueBytes, err := b.marshalValueForSet(value)
	if err != nil {
		return false, err
	}

	var found bool
	err = b.db.View(func(txn *badger.Txn) error {
		setKey := fmt.Sprintf("%s:set:%x", key, valueBytes)
		_, err := txn.Get([]byte(setKey))
		if err == nil {
			found = true
		} else if err == badger.ErrKeyNotFound {
			found = false
		} else {
			return err
		}
		return nil
	})
	return found, err
}

func (b *BadgeDBImpl) ContainsInSetWithTxn(txn *badger.Txn, key string, value any) (bool, error) {
	valueBytes, err := b.marshalValueForSet(value)
	if err != nil {
		return false, err
	}

	var found bool

	setKey := fmt.Sprintf("%s:set:%x", key, valueBytes)
	_, err = txn.Get([]byte(setKey))
	if err == nil {
		found = true
	} else if err == badger.ErrKeyNotFound {
		found = false
	} else {
		return false, err
	}
	return found, err
}

func (b *BadgeDBImpl) GetSet(key string) ([][]byte, error) {
	var result [][]byte

	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(key + ":set:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			result = append(result, val)
		}
		return nil
	})

	return result, err
}

func (b *BadgeDBImpl) GetSetWithTxn(txn *badger.Txn, key string) ([][]byte, error) {
	var result [][]byte

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := []byte(key + ":set:")
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		val, err := item.ValueCopy(nil)
		if err != nil {
			return result, err
		}
		result = append(result, val)
	}

	return result, nil
}

func (b *BadgeDBImpl) ClearSet(key string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(key + ":set:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := txn.Delete(item.KeyCopy(nil))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *BadgeDBImpl) ClearSetWithTxn(txn *badger.Txn, key string) error {

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := []byte(key + ":set:")
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		err := txn.Delete(item.KeyCopy(nil))
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BadgeDBImpl) marshalValueForSet(value any) ([]byte, error) {
	switch v := value.(type) {
	case string:
		return []byte(v), nil
	case int, int64, float64, bool:
		return []byte(fmt.Sprintf("%v", v)), nil
	default:
		return json.Marshal(v)
	}
}

// BatchAddToSet adds multiple values to a set in a single transaction
func (b *BadgeDBImpl) BatchAddToSet(key string, values []any) error {
	return b.db.Update(func(txn *badger.Txn) error {
		for _, value := range values {
			valueBytes, err := b.marshalValueForSet(value)
			if err != nil {
				return err
			}
			setKey := fmt.Sprintf("%s:set:%x", key, valueBytes)
			if err := txn.Set([]byte(setKey), valueBytes); err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchAddToSetWithTxn adds multiple values to a set in an existing transaction
func (b *BadgeDBImpl) BatchAddToSetWithTxn(txn *badger.Txn, key string, values []any) error {
	for _, value := range values {
		valueBytes, err := b.marshalValueForSet(value)
		if err != nil {
			return err
		}
		setKey := fmt.Sprintf("%s:set:%x", key, valueBytes)
		if err := txn.Set([]byte(setKey), valueBytes); err != nil {
			return err
		}
	}
	return nil
}

// BatchRemoveFromSet removes multiple values from a set in a single transaction
func (b *BadgeDBImpl) BatchRemoveFromSet(key string, values []any) error {
	return b.db.Update(func(txn *badger.Txn) error {
		for _, value := range values {
			valueBytes, err := b.marshalValueForSet(value)
			if err != nil {
				return err
			}
			setKey := fmt.Sprintf("%s:set:%x", key, valueBytes)
			if err := txn.Delete([]byte(setKey)); err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchRemoveFromSetWithTxn removes multiple values from a set in an existing transaction
func (b *BadgeDBImpl) BatchRemoveFromSetWithTxn(txn *badger.Txn, key string, values []any) error {
	for _, value := range values {
		valueBytes, err := b.marshalValueForSet(value)
		if err != nil {
			return err
		}
		setKey := fmt.Sprintf("%s:set:%x", key, valueBytes)
		if err := txn.Delete([]byte(setKey)); err != nil {
			return err
		}
	}
	return nil
}

// GetSetWithPrefix retrieves all sets with a given prefix
func (b *BadgeDBImpl) GetSetWithPrefix(prefix string) (map[string][][]byte, error) {
	result := make(map[string][][]byte)
	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		fullPrefix := []byte(prefix + ":set:")
		for it.Seek(fullPrefix); it.ValidForPrefix(fullPrefix); it.Next() {
			item := it.Item()
			key := string(item.Key())
			setKey := strings.TrimPrefix(key, prefix+":set:")

			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			if _, exists := result[setKey]; !exists {
				result[setKey] = make([][]byte, 0)
			}
			result[setKey] = append(result[setKey], val)
		}
		return nil
	})
	return result, err
}

// GetSetWithPrefixTxn retrieves all sets with a given prefix using an existing transaction
func (b *BadgeDBImpl) GetSetWithPrefixTxn(txn *badger.Txn, prefix string) (map[string][][]byte, error) {
	result := make(map[string][][]byte)
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = true
	it := txn.NewIterator(opts)
	defer it.Close()

	fullPrefix := []byte(prefix + ":set:")
	for it.Seek(fullPrefix); it.ValidForPrefix(fullPrefix); it.Next() {
		item := it.Item()
		key := string(item.Key())
		setKey := strings.TrimPrefix(key, prefix+":set:")

		val, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}

		if _, exists := result[setKey]; !exists {
			result[setKey] = make([][]byte, 0)
		}
		result[setKey] = append(result[setKey], val)
	}
	return result, nil
}

// ClearSetWithPrefixBatched 分批删除指定前缀的集合
// batchSize 参数控制每个事务中处理的最大键数
func (b *BadgeDBImpl) ClearSetWithPrefixBatched(prefix string, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 10000 // 增加默认批处理大小
	}

	// 键计数器
	var totalDeleted int
	startTime := time.Now()

	for {
		// 在一个更大的事务中执行更多删除操作
		var keysDeleted int
		err := b.db.Update(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false // 不需要获取值
			it := txn.NewIterator(opts)
			defer it.Close()

			keysDeleted = 0
			fullPrefix := []byte(prefix + ":set:")

			// 直接在事务中进行迭代和删除
			for it.Seek(fullPrefix); it.ValidForPrefix(fullPrefix) && keysDeleted < batchSize; it.Next() {
				key := it.Item().KeyCopy(nil)
				if err := txn.Delete(key); err != nil {
					return err
				}
				keysDeleted++
			}
			return nil
		})

		if err != nil {
			return err
		}

		totalDeleted += keysDeleted

		// 每删除一定数量的键记录一次日志
		if totalDeleted%50000 == 0 && totalDeleted > 0 {
			elapsed := time.Since(startTime)
			rate := float64(totalDeleted) / elapsed.Seconds()
			belogs.Info("ClearSetWithPrefixBatched: deleted", totalDeleted, "keys so far, rate:",
				fmt.Sprintf("%.2f keys/sec", rate), "elapsed:", elapsed)
		}

		// 如果没有删除任何键，表示已完成
		if keysDeleted == 0 {
			break
		}
	}

	elapsed := time.Since(startTime)
	rate := float64(totalDeleted) / elapsed.Seconds()
	belogs.Info("ClearSetWithPrefixBatched: completed, deleted", totalDeleted, "keys, rate:",
		fmt.Sprintf("%.2f keys/sec", rate), "total time:", elapsed)

	return nil
}

// ClearSetWithPrefixFast 使用DropPrefix方法快速删除指定前缀的所有键
// 这是最暴力也是最快的方法，适用于需要清空大量数据的场景
func (b *BadgeDBImpl) ClearSetWithPrefixFast(prefix string) error {
	fullPrefix := []byte(prefix + ":set:")

	// 使用DropPrefix直接删除所有前缀匹配的键
	return b.db.DropPrefix(fullPrefix)
}
