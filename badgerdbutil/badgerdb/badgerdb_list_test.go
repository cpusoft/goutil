package badgerdb

import (
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestBadgeDBImpl_Append(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")
	bDB := NewBadgeDB()
	err := bDB.Init(opts)
	assert.Nil(t, err)

	defer bDB.Close()

	baseKey := "baseKey"
	l1 := "l1"
	l2 := "l2"
	l3 := "l3"
	err = bDB.ClearList(baseKey)
	assert.Nil(t, err)

	err = bDB.Append(baseKey, l1)
	err = bDB.Append(baseKey, l2)
	err = bDB.Append(baseKey, l3)

	assert.Nil(t, err)

	v, _ := bDB.GetList(baseKey)
	assert.Equal(t, 3, len(v))

}

func TestBadgeDBImpl_AppendWithTnx(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")
	bDB := NewBadgeDB()
	err := bDB.Init(opts)
	assert.Nil(t, err)

	defer bDB.Close()

	baseKey := "baseKey"
	l1 := "l1"
	l2 := "l2"
	l3 := "l3"

	bDB.RunWithTxn(func(txn *badger.Txn) error {

		err = bDB.ClearListWithTxn(txn, baseKey)
		assert.Nil(t, err)

		err = bDB.AppendWithTxn(txn, baseKey, l1)
		err = bDB.AppendWithTxn(txn, baseKey, l2)
		err = bDB.AppendWithTxn(txn, baseKey, l3)

		assert.Nil(t, err)

		v, _ := bDB.GetListWithTxn(txn, baseKey)
		assert.Equal(t, 3, len(v))

		return nil
	})

}
