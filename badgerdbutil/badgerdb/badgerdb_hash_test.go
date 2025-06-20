package badgerdb

import (
	"strings"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestBadgeDBImpl_HSet(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")
	bDB := NewBadgeDB()
	err := bDB.Init(opts)
	assert.Nil(t, err)

	defer bDB.Close()

	baseKey := "zhangsan"

	err = bDB.HSet(baseKey, "name", "张三")
	assert.Nil(t, err)
	bDB.HSet(baseKey, "age", "30")
	assert.Nil(t, err)

	nameByte, err := bDB.HGet(baseKey, "name")
	assert.Nil(t, err)
	assert.Equal(t, "张三", strings.Trim(string(nameByte), "\""))

}
