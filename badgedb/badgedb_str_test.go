package badgedb

import (
	"fmt"
	"log"
	"testing"
)

func TestStoreWithCompositeKey(t *testing.T) {
	Init()
	columns := map[string]string{
		"name": "John",
		"age":  "30",
	}

	err := StoreWithCompositeKey("user", "123", columns)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("User data stored.")
	columnsToQuery := map[string]string{
		"name": "John",
		"age":  "30",
	}

	result, err := QueryByCompositeKey[string]("user", columnsToQuery)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Query result:", string(result))
}
