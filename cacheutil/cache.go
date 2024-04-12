package cacheutil

import (
	"sync"
)

type Cache struct {
	datas    map[string]*BaseCache
	mutex    sync.RWMutex
	capacity uint64
}

func NewCache(capacity uint64) *Cache {
	c := &Cache{
		datas:    make(map[string]*BaseCache, capacity),
		capacity: capacity,
	}
	return c
}

func (c *Cache) AddBaseCache(baseKey string, baseCapacity uint64) {
	if baseKey == "" {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.datas[baseKey]; ok {
		// only add once
		return
	}
	c.datas[baseKey] = NewBaseCache(baseCapacity)
}

func (c *Cache) Set(baseKey string, key string, value any) {
	if baseKey == "" || key == "" || value == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.datas[baseKey]; !ok {
		return
	}
	c.datas[baseKey].Set(key, value)
}

func (c *Cache) Sets(baseKey string, values []any, getKey func(value any) string) {
	if baseKey == "" || len(values) == 0 || getKey == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.datas[baseKey]; !ok {
		return
	}
	c.datas[baseKey].Sets(values, getKey)
}

func (c *Cache) Get(baseKey string, key string) (value any, exist bool) {
	if baseKey == "" || key == "" {
		return nil, false
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if _, ok := c.datas[baseKey]; !ok {
		return
	}
	return c.datas[baseKey].Get(key)
}

func (c *Cache) Gets(baseKey string) (values map[string]any, exist bool) {
	if baseKey == "" {
		return nil, false
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if _, ok := c.datas[baseKey]; !ok {
		return nil, false
	}
	return c.datas[baseKey].Gets(), true
}

func (c *Cache) GetCount(baseKey string) int {
	m, ok := c.Gets(baseKey)
	if ok {
		return len(m)
	} else {
		return 0
	}
}

func (c *Cache) Update(baseKey string, key string, value any) {
	if baseKey == "" || key == "" || value == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.datas[baseKey]; !ok {
		return
	}
	c.datas[baseKey].Update(key, value)
}

func (c *Cache) Remove(baseKey string, key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.datas[baseKey]; !ok {
		return
	}
	c.datas[baseKey].Remove(key)
}
func (c *Cache) RemoveAll(baseKey string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.datas[baseKey]; !ok {
		return
	}
	c.datas[baseKey].RemoveAll()
}

func (c *Cache) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for baseKey := range c.datas {
		c.datas[baseKey].RemoveAll()
	}
	c.datas = make(map[string]*BaseCache, c.capacity)
}
