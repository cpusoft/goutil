package cayleyutils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	_ "github.com/cayleygraph/cayley/graph/kv/badger"
	"github.com/cayleygraph/cayley/quad"
)

func TestCayleyStore_QuerySubject(t *testing.T) {
	boltDbPath := "./boltdb"
	store := NewCayleyStore(true, boltDbPath)

	start := time.Now()
	i := 0
	for {

		if i > 3000 {
			break
		}

		// 添加一些关系
		err := store.AddQuad("Alice", "knows", "Bob", "")
		assert.Nil(t, err)
		err = store.AddQuad("Bob", "knows", "Charlie", "")
		assert.Nil(t, err)

		i += 1
	}

	// 查询 Alice 的关系
	results := store.QuerySubject("Alice")
	for _, q := range results {
		fmt.Println(q)
	}

	path := store.BuildPath("Alice", "knows")
	store.IteratePath(path, func(v quad.Value) {
		fmt.Println("Alice knows:", quad.NativeOf(v))
	})

	fmt.Println("存储时长:", time.Since(start))

	store.Close()
}

func TestCayleyStore_QuerySubject2(t *testing.T) {
	boltDbPath := "./boltdb"
	store := NewCayleyStore(true, boltDbPath)
	start := time.Now()
	// 添加一些关系
	err := store.AddQuad("张三", "knows", "李四", "")
	assert.Nil(t, err)
	err = store.AddQuad("李四", "knows", "王五", "")
	assert.Nil(t, err)
	err = store.AddQuad("李四", "knows", "bob", "")
	assert.Nil(t, err)
	err = store.AddQuad("王五", "knows", "马六", "")
	assert.Nil(t, err)

	path := store.BuildRecursivePath("张三", "knows", 10)

	store.IteratePath(path, func(v quad.Value) {
		fmt.Println("张三 knows:", quad.NativeOf(v))
	})

	result := store.BFS("张三", "knows")
	fmt.Println(result)

	result2 := store.BFSByLevelWithMap("张三", "knows")
	fmt.Println(result2)
	fmt.Println("存储时长:", time.Since(start))
	// 关闭存储
	store.Close()
}
