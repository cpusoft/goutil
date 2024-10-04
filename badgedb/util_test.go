package badgedb

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestInsert(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")

	InitDB(opts)

	defer CloseDB()
	// *************** 1.insert  str **************
	err := Insert("name", "张三")
	assert.Nil(t, err)
	// *************** 2. Get  **************
	name, exist, err := Get[string]("name")
	assert.Nil(t, err)
	assert.Equal(t, true, exist)

	assert.Equal(t, "张三", strings.Trim(name, "\""))
	// *************** 3. Delete  **************
	err = Delete("name")
	assert.Nil(t, err)
	// *************** 4. Get  **************
	_, exist, err = Get[string]("name")
	assert.Nil(t, err)
	assert.Equal(t, false, exist)
	// *************** 4. Get  **************
	// *************** 4. Get  **************
	// *************** 4. Get  **************
	// *************** 4. Get  **************
	type Student struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Age  int64  `json:"age"`
	}
	student := Student{
		ID:   "123456789",
		Name: "王五",
		Age:  int64(13),
	}
	err = Insert[Student](student.ID, student)
	assert.Nil(t, err)

	tStu, exist, err := Get[Student](student.ID)
	assert.Nil(t, err)
	assert.Equal(t, true, exist)
	assert.Equal(t, "王五", tStu.Name)

}

func TestInsertWithTxn(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")

	InitDB(opts)

	defer CloseDB()

	RunWithTxn(func(txn *badger.Txn) error {
		// *************** 1.insert  **************
		err := InsertWithTxn(txn, "name", "李四")
		assert.Nil(t, err)
		// *************** 2. Get  **************
		name, exist, err := GetWithTxn[string](txn, "name")
		assert.Nil(t, err)
		assert.Equal(t, true, exist)

		assert.Equal(t, "李四", strings.Trim(name, "\""))

		return nil
	})

	RunWithTxn(func(txn *badger.Txn) error {

		// *************** 3. Delete  **************
		err := DeleteWithTxn(txn, "name")
		assert.Nil(t, err)
		// *************** 4. Get  **************
		_, exist, err := GetWithTxn[string](txn, "name")
		assert.Nil(t, err)
		assert.Equal(t, false, exist)

		return nil
	})

}

func TestUtilStoreWithCompositeKey(t *testing.T) {

	opts := badger.DefaultOptions("./badgerdb")

	InitDB(opts)

	defer CloseDB()

	columns := map[string]string{
		"name": "zhangsan",
		"age":  "30",
	}

	err := StoreWithCompositeKey("user", "123", columns)
	assert.Nil(t, err)

	columnsToQuery := map[string]string{
		"name": "zhangsan",
		"age":  "30",
	}

	result, exist, err := QueryByCompositeKey[string]("user", columnsToQuery)
	assert.Nil(t, err)
	assert.Equal(t, true, exist)
	assert.Equal(t, "123:user", result)
}
