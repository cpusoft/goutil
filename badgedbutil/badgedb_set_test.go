package badgedb

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBadgeDBImpl_AddToSet(t *testing.T) {
	opts := badger.DefaultOptions("./badgerdb")
	bDB := NewBadgeDB()
	err := bDB.Init(opts)
	assert.Nil(t, err)

	baseKey := "baseKey"
	l1 := "l1"
	l2 := "l2"
	l3 := "l3"
	l4 := "l3"
	err = bDB.ClearSet(baseKey)
	assert.Nil(t, err)

	err = bDB.AddToSet(baseKey, l1)
	assert.Nil(t, err)

	err = bDB.AddToSet(baseKey, l2)
	assert.Nil(t, err)

	err = bDB.AddToSet(baseKey, l3)
	assert.Nil(t, err)
	err = bDB.AddToSet(baseKey, l4)
	assert.Nil(t, err)

	l, err := bDB.GetSet(baseKey)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(l))

}

func TestBadgeDBImpl_AddToSet1(t *testing.T) {
	opts := badger.DefaultOptions("./badgerdb")
	bDB := NewBadgeDB()
	err := bDB.Init(opts)
	assert.Nil(t, err)

	baseKey := "baseKey"

	err = bDB.ClearSet(baseKey)
	assert.Nil(t, err)

	err = bDB.ClearList(baseKey)
	assert.Nil(t, err)

	for i := 0; i < 100000; i++ {
		bDB.Append(baseKey, i)
	}
	t1 := time.Now()
	r, err := bDB.GetList(baseKey)
	assert.Equal(t, 100000, len(r))
	fmt.Println(len(r), time.Now().Sub(t1))

	for i := 0; i < 100000; i++ {
		bDB.AddToSet(baseKey, i)
	}
	t2 := time.Now()
	r, err = bDB.GetSet(baseKey)
	assert.Equal(t, 100000, len(r))
	fmt.Println(len(r), time.Now().Sub(t2))

	//err = bDB.AddToSet(baseKey, l1)
	//assert.Nil(t, err)
	//
	//err = bDB.AddToSet(baseKey, l2)
	//assert.Nil(t, err)
	//
	//err = bDB.AddToSet(baseKey, l3)
	//assert.Nil(t, err)
	//err = bDB.AddToSet(baseKey, l4)
	//assert.Nil(t, err)
	//
	//l, err := bDB.GetSet(baseKey)
	//assert.Nil(t, err)
	//assert.Equal(t, 3, len(l))

}
