package badgedb

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
	"testing"
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
