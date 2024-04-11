package maputil

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
	c.datas[baseKey] = NewBaseCache(baseCapacity)
}

func (c *Cache) Set(baseKey string, key string, value any) {
	if baseKey == "" || key == "" || value == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.datas[baseKey].Set(key, value)
}
func (c *Cache) Sets(baseKey string, values []any, getKey func(value any) string) {
	if baseKey == "" || len(values) == 0 || getKey == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.datas[baseKey].Sets(values, getKey)
}

func (c *Cache) Get(baseKey string, key string) (value any, exist bool) {
	if baseKey == "" || key == "" {
		return nil, false
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.datas[baseKey].Get(key)
}

func (c *Cache) Gets(baseKey string) map[string]any {
	if baseKey == "" {
		return make(map[string]any, 0)
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.datas[baseKey].Gets()
}

func (c *Cache) GetCount(baseKey string) int {
	m := c.Gets(baseKey)
	return len(m)
}

func (c *Cache) Update(baseKey string, key string, value any) {
	if baseKey == "" || key == "" || value == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.datas[baseKey].Update(key, value)
}

func (c *Cache) Remove(baseKey string, key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.datas[baseKey].Remove(key)
}
func (c *Cache) RemoveAll(baseKey string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
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
