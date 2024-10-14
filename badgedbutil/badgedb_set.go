package badgedb

import (
	"encoding/json"
	"fmt"
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
