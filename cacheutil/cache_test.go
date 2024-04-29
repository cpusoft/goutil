package cacheutil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

type TestModel struct {
	Name    string
	Address string

	Ski string
	Aki string
}

func getKey(value any) string {
	fmt.Println(value)
	t := value.(*TestModel)
	return t.Name
}

func TestDualCache(t *testing.T) {
	cache := NewDualCache(100)
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
	m, ok, err := cache.Gets("test")
	fmt.Println(m, ok, err)
}
func TestNewAdjacentCache(t *testing.T) {
	cache := NewAdjacentCache(100)

	p := &TestModel{Name: "parent", Ski: "ski1"}
	c1 := &TestModel{Name: "c1", Aki: "ski1"}
	c2 := &TestModel{Name: "c2", Aki: "ski1"}

	cache.AddAdjacentBaseCacheByParentData("ski1", "p", p)
	cache.AddAdjacentBaseCacheByChildData("ski1", "c1", c1)
	cache.AddAdjacentBaseCacheByChildData("ski1", "c2", c2)
	c, _ := cache.GetAdjacentBaseCache("ski1")
	pd, _ := c.GetParentData()
	cds, _ := c.GetChildDatas()
	fmt.Println(jsonutil.MarshalJson(pd))
	fmt.Println(jsonutil.MarshalJson(cds))
}
