package badgedb

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v4"
	"github.com/goccy/go-json"
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
	get, err := Get[string]("name")
	fmt.Println(err)
	fmt.Println(get)

}

// marshalValue 将任意类型的数据转换为字节切片
func marshalValue(v any) ([]byte, error) {
	return []byte(fmt.Sprintf("%v", v)), nil
}
func unmarshalValue[T any](data []byte, v *T) error {
	return json.Unmarshal(data, v)
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
