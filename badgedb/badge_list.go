package badgedb

import "github.com/dgraph-io/badger/v4"

func Append[T any](key string, value T) error {
	// 获取当前列表
	var list []T
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			// 如果 key 不存在，则初始化为空列表
			list = []T{}
			return nil
		} else if err != nil {
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// 反序列化为列表
		err = unmarshalValue(val, &list)
		return err
	})

	if err != nil {
		return err
	}

	// 将新的值追加到列表中
	list = append(list, value)

	// 将更新后的列表存储回数据库
	return db.Update(func(txn *badger.Txn) error {
		// 序列化列表
		listBytes, err := marshalValue(list)
		if err != nil {
			return err
		}

		// 存储回数据库
		return txn.Set([]byte(key), listBytes)
	})
}

func AppendWithTxn[T any](txn *badger.Txn, key string, value T) error {
	// 获取当前列表
	var list []T
	item, err := txn.Get([]byte(key))
	if err == badger.ErrKeyNotFound {
		// 如果 key 不存在，则初始化为空列表
		list = []T{}
	} else if err != nil {
		return err
	} else {
		// 获取当前的值
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// 反序列化为列表
		err = unmarshalValue(val, &list)
		if err != nil {
			return err
		}
	}

	// 将新的值追加到列表中
	list = append(list, value)

	// 序列化列表
	listBytes, err := marshalValue(list)
	if err != nil {
		return err
	}

	// 存储更新后的列表回数据库
	return txn.Set([]byte(key), listBytes)
}

func GetList[T any](key string) ([]T, error) {
	var list []T

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			// 如果 key 不存在，返回空列表
			list = []T{}
			return nil
		} else if err != nil {
			return err
		}

		// 获取当前值
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// 反序列化为列表
		err = unmarshalValue(val, &list)
		return err
	})

	if err != nil {
		return nil, err
	}

	return list, nil
}

func GetListWithTxn[T any](txn *badger.Txn, key string) ([]T, error) {
	var list []T

	item, err := txn.Get([]byte(key))
	if err == badger.ErrKeyNotFound {
		// 如果 key 不存在，返回空列表
		return []T{}, nil
	} else if err != nil {
		return nil, err
	}

	// 获取值
	val, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	// 反序列化列表
	err = unmarshalValue(val, &list)
	if err != nil {
		return nil, err
	}

	return list, nil
}
