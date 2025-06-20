package badgerdb

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestBadgeDBImpl_Insert(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")
	bDB := NewBadgeDB()
	err := bDB.Init(opts)

	defer bDB.Close()
	assert.Nil(t, err)

	err = bDB.Insert("name", "张三")
	assert.Nil(t, err)

	name, exist, err := bDB.Get("name")
	assert.Nil(t, err)
	assert.Equal(t, true, exist)
	assert.Equal(t, "张三", strings.Trim(string(name), "\""))
	fmt.Println("名字：", string(name))

	err = bDB.Delete("name")
	assert.Nil(t, err)

	_, exist, err = bDB.Get("name")
	assert.Nil(t, err)
	assert.Equal(t, false, exist)

}

func TestBadgeDBImpl_InsertWithTxn(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")
	bDB := NewBadgeDB()
	err := bDB.Init(opts)

	defer bDB.Close()
	assert.Nil(t, err)

	bDB.RunWithTxn(func(txn *badger.Txn) error {

		err := bDB.InsertWithTxn(txn, "name", "张三")
		assert.Nil(t, err)

		name, exist, err := bDB.GetWithTxn(txn, "name")
		assert.Nil(t, err)
		assert.Equal(t, true, exist)
		assert.Equal(t, "张三", strings.Trim(string(name), "\""))

		err = bDB.DeleteWithTxn(txn, "name")
		assert.Nil(t, err)

		_, exist, err = bDB.GetWithTxn(txn, "name")
		assert.Nil(t, err)
		assert.Equal(t, false, exist)

		return nil
	})

}

func TestStoreWithCompositeKey(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")
	bDB := NewBadgeDB()
	err := bDB.Init(opts)
	assert.Nil(t, err)

	defer bDB.Close()

	columns := map[string]string{
		"name": "John",
		"age":  "30",
	}
	entity := "user"

	err = bDB.StoreWithCompositeKey(entity, "123", columns)
	assert.Nil(t, err)

	result, exist, err := bDB.QueryByCompositeKey(entity, columns)
	assert.Nil(t, err)
	assert.Equal(t, true, exist)
	assert.Equal(t, "123:user", string(result))

}
