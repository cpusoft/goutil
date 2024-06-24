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

	if baseKey == "" || element == "" {
		return 0, errors.New("baseKey or element is empty")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
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
	if baseKey == "" {
		return nil, errors.New("baseKey is empty")
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()
	m, ok := c.dualBaseCaches[baseKey]
	if !ok {
		return nil, errors.New("baseKey not in dualBaseCaches")
	}
	elements := make([]string, 0, len(m.datas))
	for k := range m.datas {
		elements = append(elements, k)
	}
	return elements, nil
}
