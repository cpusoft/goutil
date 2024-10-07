package badgedb

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"strconv"
)

func (b *BadgeDBImpl) Append(key string, value any) error {

	var valueBytes []byte
	switch v := value.(type) {
	case string:
		valueBytes = []byte(v)
	case []byte:
		valueBytes = v
	default:

		var err error
		valueBytes, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %v", err)
		}
	}

	return b.db.Update(func(txn *badger.Txn) error {
		// 获取当前尾部索引
		tailKey := key + ":tail"
		item, err := txn.Get([]byte(tailKey))
		var tailIndex int64
		if err == badger.ErrKeyNotFound {
			tailIndex = 0
		} else {
			val, _ := item.ValueCopy(nil)
			tailIndex, _ = strconv.ParseInt(string(val), 10, 64)
		}

		// 存储新元素到尾部
		newKey := fmt.Sprintf("%s:%d", key, tailIndex)
		err = txn.Set([]byte(newKey), valueBytes)
		if err != nil {
			return err
		}

		// 更新尾部索引
		tailIndex++
		return txn.Set([]byte(tailKey), []byte(strconv.FormatInt(tailIndex, 10)))
	})
}

func (b *BadgeDBImpl) AppendWithTxn(txn *badger.Txn, key string, value any) error {
	var valueBytes []byte
	switch v := value.(type) {
	case string:
		valueBytes = []byte(v)
	case []byte:
		valueBytes = v
	default:
		var err error
		valueBytes, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %v", err)
		}
	}

	tailKey := key + ":tail"
	item, err := txn.Get([]byte(tailKey))
	var tailIndex int64
	if err == badger.ErrKeyNotFound {
		tailIndex = 0
	} else {
		val, _ := item.ValueCopy(nil)
		tailIndex, _ = strconv.ParseInt(string(val), 10, 64)
	}
	newKey := fmt.Sprintf("%s:%d", key, tailIndex)
	err = txn.Set([]byte(newKey), valueBytes)
	if err != nil {
		return err
	}

	tailIndex++
	return txn.Set([]byte(tailKey), []byte(strconv.FormatInt(tailIndex, 10)))

}

func (b *BadgeDBImpl) GetList(key string) ([][]byte, error) {
	var result [][]byte
	err := b.db.View(func(txn *badger.Txn) error {
		tailKey := key + ":tail"
		item, err := txn.Get([]byte(tailKey))
		if err == badger.ErrKeyNotFound {
			return nil //
		}

		val, _ := item.ValueCopy(nil)
		tailIndex, _ := strconv.ParseInt(string(val), 10, 64)

		for i := int64(0); i < tailIndex; i++ {
			elementKey := fmt.Sprintf("%s:%d", key, i)
			item, err := txn.Get([]byte(elementKey))
			if err == badger.ErrKeyNotFound {
				continue
			}
			val, _ := item.ValueCopy(nil)
			result = append(result, val)
		}

		return nil
	})
	return result, err
}

func (b *BadgeDBImpl) GetListWithTxn(txn *badger.Txn, key string) ([][]byte, error) {
	var result [][]byte

	tailKey := key + ":tail"
	item, err := txn.Get([]byte(tailKey))
	if err == badger.ErrKeyNotFound {
		return result, nil
	}

	val, _ := item.ValueCopy(nil)
	tailIndex, _ := strconv.ParseInt(string(val), 10, 64)

	for i := int64(0); i < tailIndex; i++ {
		elementKey := fmt.Sprintf("%s:%d", key, i)
		item, err := txn.Get([]byte(elementKey))
		if err == badger.ErrKeyNotFound {
			continue
		}
		val, _ := item.ValueCopy(nil)
		result = append(result, val)
	}

	return result, err
}

func (b *BadgeDBImpl) ClearList(key string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		tailKey := key + ":tail"
		item, err := txn.Get([]byte(tailKey))
		if err == badger.ErrKeyNotFound {
			return nil
		}

		val, _ := item.ValueCopy(nil)
		tailIndex, _ := strconv.ParseInt(string(val), 10, 64)

		for i := int64(0); i < tailIndex; i++ {
			elementKey := fmt.Sprintf("%s:%d", key, i)
			err := txn.Delete([]byte(elementKey))
			if err != nil {
				return err
			}
		}

		err = txn.Delete([]byte(key + ":head"))
		if err != nil {
			return err
		}
		err = txn.Delete([]byte(tailKey))
		return err
	})
}

func (b *BadgeDBImpl) ClearListWithTxn(txn *badger.Txn, key string) error {

	tailKey := key + ":tail"
	item, err := txn.Get([]byte(tailKey))
	if err == badger.ErrKeyNotFound {
		return nil
	}

	val, _ := item.ValueCopy(nil)
	tailIndex, _ := strconv.ParseInt(string(val), 10, 64)

	for i := int64(0); i < tailIndex; i++ {
		elementKey := fmt.Sprintf("%s:%d", key, i)
		err := txn.Delete([]byte(elementKey))
		if err != nil {
			return err
		}
	}

	err = txn.Delete([]byte(key + ":head"))
	if err != nil {
		return err
	}
	err = txn.Delete([]byte(tailKey))
	return err

}

func (b *BadgeDBImpl) marshalValueForList(value any) ([]byte, error) {
	switch v := value.(type) {
	case string:
		return []byte(v), nil
	case int, int64, float64, bool:
		return []byte(fmt.Sprintf("%v", v)), nil
	default:
		return json.Marshal(v)
	}
}

func (b *BadgeDBImpl) unmarshalValueForList(data []byte, v any) error {
	if json.Valid(data) {
		return json.Unmarshal(data, v)
	} else {
		return fmt.Errorf("invalid data format: %s", string(data))
	}
}
