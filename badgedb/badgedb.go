package badgedb

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/dgraph-io/badger/v4"
)

var db *badger.DB

func Init() {
	opts := badger.DefaultOptions("./badgerdb")
	tdb, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	db = tdb

}
func Close() {
	db.Close()
}

func BadgerDB() {

	Init()

	Insert("name", "baozhuo")
	get, exists, err := Get[string]("name")
	fmt.Println(err)
	fmt.Println(get)
	fmt.Println(exists)

}

func marshalValue(v any) ([]byte, error) {
	// 使用JSON序列化保持类型信息
	return json.Marshal(v)
}
func unmarshalValue[T any](data []byte, v *T) error {
	switch any(v).(type) {
	case *string:
		*v = any(string(data)).(T) // 直接处理为字符串
	case *int:
		// 尝试将字节数据转换为整数
		i, err := strconv.Atoi(string(data))
		if err != nil {
			return err
		}
		*v = any(i).(T)
	case *float64:
		// 尝试将字节数据转换为浮点数
		f, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return err
		}
		*v = any(f).(T)
	default:
		// 其他类型，假设是结构化数据，使用 JSON 反序列化
		return json.Unmarshal(data, v)
	}
	return nil
}
func IterateBadgerDB() error {
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		//opts.PrefetchValues = true // 是否提前预取值，提高遍历效率
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()

			err := item.Value(func(val []byte) error {
				fmt.Printf("Key: %s, Value: %s\n", key, val)
				return nil
			})

			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
func RunWithTxn(txnFunc func(txn *badger.Txn) error) error {
	return db.Update(func(txn *badger.Txn) error {
		return txnFunc(txn)
	})
}
