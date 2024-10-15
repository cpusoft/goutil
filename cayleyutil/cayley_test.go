package cayleyutils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	_ "github.com/cayleygraph/cayley/graph/kv/badger"
	"github.com/cayleygraph/cayley/quad"
)

func TestCayleyStore_QuerySubject(t *testing.T) {
	boltDbPath := "./boltdb"
	store := NewCayleyStore(boltDbPath)

	// 添加一些关系
	err := store.AddQuad("Alice", "knows", "Bob", "")
	assert.Nil(t, err)
	err = store.AddQuad("Bob", "knows", "Charlie", "")
	assert.Nil(t, err)
	// 查询 Alice 的关系
	results := store.QuerySubject("Alice")
	for _, q := range results {
		fmt.Println(q)
	}

	path := store.BuildPath("Alice", "knows")
	store.IteratePath(path, func(v quad.Value) {
		fmt.Println("Alice knows:", quad.NativeOf(v))
	})

	// 关闭存储
	store.Close()
}
