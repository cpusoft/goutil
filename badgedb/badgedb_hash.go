package badgedb

import "github.com/dgraph-io/badger/v4"

func (b *BadgeDBImpl) HSet(key, field string, value any) error {
	// Combine key and field as BadgerDB key
	hashKey := key + ":" + field

	// Serialize value to bytes
	valueBytes, err := marshalValue(value)
	if err != nil {
		return err
	}

	// Store field and value to database
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(hashKey), valueBytes)
	})
}

func (b *BadgeDBImpl) HSetWithTxn(txn *badger.Txn, key, field string, value any) error {
	// Combine key and field as BadgerDB key
	hashKey := key + ":" + field

	// Serialize value to bytes
	valueBytes, err := marshalValue(value)
	if err != nil {
		return err
	}
	// Store field and value to database
	return txn.Set([]byte(hashKey), valueBytes)
}

func (b *BadgeDBImpl) HGet(key, field string) ([]byte, error) {
	var result []byte
	hashKey := key + ":" + field

	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hashKey))
		if err != nil {
			return err
		}

		result, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// Deserialize to type T
		return nil
	})

	if err != nil {
		return result, err
	}
	return result, nil
}

func (b *BadgeDBImpl) HGetWithTxn(txn *badger.Txn, key, field string) ([]byte, error) {
	var result []byte
	hashKey := key + ":" + field

	item, err := txn.Get([]byte(hashKey))
	if err != nil {
		return result, err
	}

	result, err = item.ValueCopy(nil)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (b *BadgeDBImpl) HDel(key, field string) error {
	hashKey := key + ":" + field

	// Delete field
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(hashKey))
	})
}

func (b *BadgeDBImpl) HDelWithTxn(txn *badger.Txn, key, field string) error {
	hashKey := key + ":" + field

	// Delete field
	return txn.Delete([]byte(hashKey))
}
