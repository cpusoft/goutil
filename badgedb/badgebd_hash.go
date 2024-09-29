package badgedb

import "github.com/dgraph-io/badger/v4"

func HSet[T any](key, field string, value T) error {
	// 组合 key 和 field 作为 BadgerDB 的键
	hashKey := key + ":" + field

	// 将 value 序列化为字节
	valueBytes, err := marshalValue(value)
	if err != nil {
		return err
	}

	// 存储字段和值到数据库
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(hashKey), valueBytes)
	})
}

func HSetWithTxn[T any](txn *badger.Txn, key, field string, value T) error {
	// 组合 key 和 field 作为 BadgerDB 的键
	hashKey := key + ":" + field

	// 将 value 序列化为字节
	valueBytes, err := marshalValue(value)
	if err != nil {
		return err
	}
	// 存储字段和值到数据库
	return txn.Set([]byte(hashKey), valueBytes)
}

func HGet[T any](key, field string) (T, error) {
	var result T
	hashKey := key + ":" + field

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hashKey))
		if err != nil {
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// 反序列化为类型 T
		return unmarshalValue(val, &result)
	})

	if err != nil {
		return result, err
	}
	return result, nil
}

func HGetWithTxn[T any](txn *badger.Txn, key, field string) (T, error) {
	var result T
	hashKey := key + ":" + field

	item, err := txn.Get([]byte(hashKey))
	if err != nil {
		return result, err
	}

	val, err := item.ValueCopy(nil)
	if err != nil {
		return result, err
	}

	// 反序列化为类型 T
	err = unmarshalValue(val, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func HDel(key, field string) error {
	hashKey := key + ":" + field

	// 删除字段
	return db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(hashKey))
	})
}

func HDelWithTxn(txn *badger.Txn, key, field string) error {
	hashKey := key + ":" + field

	// 删除字段
	return txn.Delete([]byte(hashKey))
}
