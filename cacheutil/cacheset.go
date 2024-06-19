package cacheutil

import (
	"errors"
	"sync"
)

type CacheSet struct {
	dualBaseCaches map[string]*DualBaseCache
	mutex          sync.RWMutex
	capacity       uint64
}

func NewCacheSet(capacity uint64) *CacheSet {
	c := &CacheSet{
		dualBaseCaches: make(map[string]*DualBaseCache, capacity),
		capacity:       capacity,
	}
	return c
}

func (c *CacheSet) Append(baseKey string, element string) (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if baseKey == "" {
		return 0, errors.New("baseKey is empty")
	}

	m, ok := c.dualBaseCaches[baseKey]
	if !ok {
		c.dualBaseCaches[baseKey] = NewDualBaseCache(c.capacity)
		m = c.dualBaseCaches[baseKey]
	}
	m.Set(element, true)
	count, _ := m.Count()
	return int64(count), nil
}

func (c *CacheSet) ListElements(baseKey string) ([]string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	elements := make([]string, 0)
	if baseKey == "" {
		return elements, errors.New("baseKey is empty")
	}

	m, ok := c.dualBaseCaches[baseKey]
	if !ok {
		return elements, errors.New("baseKey not in dualBaseCaches")
	}

	for k := range m.datas {
		elements = append(elements, k)
	}
	return elements, nil
}
