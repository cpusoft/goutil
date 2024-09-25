package badgedb

import (
	"encoding/json"
	"fmt"
	"github.com/cpusoft/goutil/belogs"
	"github.com/dgraph-io/badger/v4"
	"log"
	"sort"
	"strings"
)

// Insert 泛型函数用于插入键值对到数据库
func Insert[T any](key string, value T) error {
	// 将 value 转换为字节切片
	valueBytes, err := marshalValue(value)
	if err != nil {
		return err
	}

	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), valueBytes)
	})

	if err != nil {
		log.Fatal(err)
	}
	return err
}

func InsertWithTxn[T any](txn *badger.Txn, key string, value T) error {
	// 将 value 转换为字节切片
	valueBytes, err := marshalValue(value)
	if err != nil {
		return err
	}
	err = txn.Set([]byte(key), valueBytes)

	if err != nil {
		log.Fatal(err)
	}
	return err
}
func Delete(key string) error {
	err := db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	if err != nil {
		log.Fatal(err)
	}
	return err
}

func DeleteWithTxn(txn *badger.Txn, key string) error {
	err := txn.Delete([]byte(key))
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func BatchInsert[T any](data map[string]T) map[string]error {
	if len(data) <= 0 {
		return nil
	}
	errs := make(map[string]error)
	wb := db.NewWriteBatch() // 创建一个批量写入对象
	defer wb.Cancel()        // 在出错或完成时取消或提交

	for key, value := range data {
		// 将 value 转换为字节切片
		valueBytes, err := marshalValue(value)
		if err != nil {
			errs[key] = err
			belogs.Error("BatchInsert, marshalValue failed, but continue, key:", key)
			continue
		}
		err = wb.Set([]byte(key), valueBytes) // 添加要写入的 key-value 对
		if err != nil {
			errs[key] = err
			belogs.Error("BatchInsert, Set failed, but continue, key:", key)
			continue
		}
	}

	// 提交批量写入
	err := wb.Flush()
	if err != nil {
		for key := range data {
			errs[key] = err // 标记所有未成功写入的键
		}
		return errs
	}

	return nil
}

func Get[T any](key string) (T, error) {
	var result T

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// 调用辅助函数将 []byte 转换为泛型类型 T
		result, err = convertToType[T](val)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
		return result, err
	}

	return result, nil
}

func GetWithTxn[T any](txn *badger.Txn, key string) (T, error) {
	var result T

	item, err := txn.Get([]byte(key))
	if err != nil {
		return result, err
	}

	val, err := item.ValueCopy(nil)
	if err != nil {
		return result, err
	}

	// 调用辅助函数将 []byte 转换为泛型类型 T
	result, err = convertToType[T](val)
	if err != nil {
		return result, err
	}

	if err != nil {
		log.Fatal(err)
		return result, err
	}

	return result, nil
}

func MGet[T any](keys []string) (map[string]T, error) {
	results := make(map[string]T)
	err := db.View(func(txn *badger.Txn) error {

		for _, key := range keys {

			var result T
			item, err := txn.Get([]byte(key))
			if err != nil {
				return err
			}

			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			// 调用辅助函数将 []byte 转换为泛型类型 T
			result, err = convertToType[T](val)
			if err != nil {
				return err
			}
			results[key] = result
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
		return results, err
	}

	return results, nil
}

func MGetWithTxn[T any](txn *badger.Txn, keys []string) (map[string]T, error) {
	results := make(map[string]T)
	for _, key := range keys {

		var result T
		item, err := txn.Get([]byte(key))
		if err != nil {
			return results, err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return results, err
		}

		// 调用辅助函数将 []byte 转换为泛型类型 T
		result, err = convertToType[T](val)
		if err != nil {
			return results, err
		}
		results[key] = result
	}
	return results, nil
}

// convertToType 根据目标类型 T 将 []byte 转换为 T
func convertToType[T any](data []byte) (T, error) {
	var result T

	// 通过类型断言处理几种常见类型
	switch any(result).(type) {
	case string:
		// 如果 T 是 string 类型，直接转换
		return any(string(data)).(T), nil
	case int, int32, int64, float32, float64, bool:
		// 如果是基本类型，尝试 JSON 解析
		err := json.Unmarshal(data, &result)
		if err != nil {
			return result, err
		}
		return result, nil
	default:
		// 处理复杂类型（例如结构体）
		err := json.Unmarshal(data, &result)
		if err != nil {
			return result, err
		}
		return result, nil
	}
}

// 构建索列
func BuildColumnIndex[T any](txn *badger.Txn, entityKey string, column string, value T) error {
	// 将列名和值组合成索引键
	valueStr, err := marshalValue(value)
	if err != nil {
		return err
	}

	// 生成列索引键，形式为 "column:value:entityKey"
	indexKey := fmt.Sprintf("%s:%s:%s", column, string(valueStr), entityKey)

	// 存储索引
	err = txn.Set([]byte(indexKey), []byte(entityKey))
	if err != nil {
		return err
	}
	return nil
}

func BuildMultipleColumnIndexes(txn *badger.Txn, entityKey string, columns map[string]interface{}) error {
	for column, value := range columns {
		valueStr, err := marshalValue(value)
		if err != nil {
			return err
		}

		// 生成列索引键
		indexKey := fmt.Sprintf("%s:%s:%s", column, string(valueStr), entityKey)

		// 存储索引
		err = txn.Set([]byte(indexKey), []byte(entityKey))
		if err != nil {
			return err
		}
	}
	return nil
}

// --------构造复合键  ----------
func buildCompositeKey(columns ...string) string {
	// 对传入的列进行字典序排序
	sort.Strings(columns)

	return strings.Join(columns, ":")
}

// 存储数据和构造复合键索引
func StoreWithCompositeKey(entity string, id string, columns map[string]string) error {
	// 构造复合键
	compositeKey := buildCompositeKey(entity, id)

	// 将数据序列化
	userData, err := marshalValue(columns)
	if err != nil {
		return err
	}

	err = db.Update(func(txn *badger.Txn) error {
		// 存储数据
		err := txn.Set([]byte(compositeKey), userData)
		if err != nil {
			return err
		}

		// 存储列索引，基于复合键
		for col, val := range columns {
			indexKey := buildCompositeKey(entity, col, val)
			err := txn.Set([]byte(indexKey), []byte(compositeKey))
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func StoreWithCompositeKeyWithTxn(txn *badger.Txn, entity string, id string, columns map[string]string) error {
	// 构造复合键
	compositeKey := buildCompositeKey(entity, id)

	// 将数据序列化
	userData, err := marshalValue(columns)
	if err != nil {
		return err
	}

	// 存储数据
	err = txn.Set([]byte(compositeKey), userData)
	if err != nil {
		return err
	}

	// 存储列索引，基于复合键
	for col, val := range columns {
		indexKey := buildCompositeKey(entity, col, val)
		err := txn.Set([]byte(indexKey), []byte(compositeKey))
		if err != nil {
			return err
		}
	}

	return nil

}

func QueryByCompositeKey[T any](entity string, columns map[string]string) (T, error) {
	var result T
	var compositeKey string

	for col, val := range columns {
		compositeKey = buildCompositeKey(entity, col, val)
	}

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(compositeKey))
		if err != nil {
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return unmarshalValue(val, &result)
	})

	return result, err
}

func QueryByCompositeKeyWithTxn[T any](txn *badger.Txn, entity string, columns map[string]string) (T, error) {
	var result T
	var compositeKey string

	for col, val := range columns {
		compositeKey = buildCompositeKey(entity, col, val)
	}

	item, err := txn.Get([]byte(compositeKey))
	if err != nil {
		return result, err
	}

	val, err := item.ValueCopy(nil)
	if err != nil {
		return result, err
	}

	err = unmarshalValue(val, &result)
	return result, err
}

func BatchQueryByPrefix[T any](prefix string) (map[string]T, error) {
	results := make(map[string]T)

	err := db.View(func(txn *badger.Txn) error {
		// 设置迭代器选项，按前缀查询
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		// 迭代查询
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			key := string(item.Key())

			// 获取值并解析为 T
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			var parsedValue T
			err = unmarshalValue(val, &parsedValue)
			if err != nil {
				return err
			}

			// 存储到结果 map 中
			results[key] = parsedValue
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}
