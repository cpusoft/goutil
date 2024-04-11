package cacheutil

import (
	"fmt"
	"testing"
)

type TestModel struct {
	Name    string
	Address string
}

func getKey(value any) string {
	fmt.Println(value)
	t := value.(*TestModel)
	return t.Name
}

func TestCache(t *testing.T) {
	cache := NewCache(100)
	cache.AddBaseCache("test", 100)

	t1 := &TestModel{Name: "Name1", Address: "Address1"}
	t2 := &TestModel{Name: "Name2", Address: "Address2"}
	t3 := &TestModel{Name: "Name3", Address: "Address3"}
	ts := make([]any, 0)
	ts = append(ts, t1)
	ts = append(ts, t2)
	ts = append(ts, t3)
	fmt.Println(ts)
	cache.Sets("test", ts, getKey)
	m := cache.Gets("test")
	fmt.Println(m)
}
