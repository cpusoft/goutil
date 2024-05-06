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

func getName(value any) string {

	t := value.(*TestModel)
	return t.Name
}
func getAki(value any) string {

	t := value.(*TestModel)
	return t.Aki
}
func getSki(value any) string {

	t := value.(*TestModel)
	return t.Ski
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
	cache.Sets("test", ts, getName)
	m, ok, err := cache.GetsClone("test")
	fmt.Println(m, ok, err)
}
func TestNewAdjacentCache(t *testing.T) {
	cache := NewAdjacentCache(100)

	p := &TestModel{Name: "parent", Ski: "ski1"}
	c1 := &TestModel{Name: "c1", Aki: "ski1", Ski: "skic1"}
	c2 := &TestModel{Name: "c2", Aki: "ski1", Ski: "skic2"}
	c3 := &TestModel{Name: "c3", Aki: "ski1", Ski: "skic3"}
	c4 := &TestModel{Name: "c4", Aki: "ski1", Ski: "skic4"}
	anys := make([]any, 0)
	anys = append(anys, p)
	cache.AddParentData(getSki, anys, getName)

	anys = make([]any, 0)
	anys = append(anys, c1)
	anys = append(anys, c2)
	anys = append(anys, c3)
	anys = append(anys, c4)
	cache.AddChildData(getAki, anys, getName)

	c, _, _ := cache.GetBaseCache("ski1")
	pd, _ := c.GetParentData()
	cds, _ := c.GetChildDatas()
	fmt.Println(jsonutil.MarshalJson(pd))
	fmt.Println(jsonutil.MarshalJson(cds))
}

func TestNewHorizontalCache(t *testing.T) {
	cache := NewHorizontalCache(100)

	c1 := &TestModel{Name: "c1", Aki: "ski1"}
	c2 := &TestModel{Name: "c2", Aki: "ski1"}
	c3 := &TestModel{Name: "c3", Aki: "ski1"}
	c4 := &TestModel{Name: "c4", Aki: "ski1"}
	anys := make([]any, 0)
	anys = append(anys, c1)
	anys = append(anys, c2)
	anys = append(anys, c3)
	anys = append(anys, c4)

	cache.Sets(getAki, anys, getName)
	c, ok, err := cache.GetsClone("ski1")
	fmt.Println(c, ok, err)
}
