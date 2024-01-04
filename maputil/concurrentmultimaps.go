package maputil

import (
	"sync"
)

type ConcurrentMultiMap struct {
	concurrentMap map[string]any
	mutex         sync.RWMutex
}

func NewConcurrentMultiMap(capacity int) *ConcurrentMultiMap {
	c := &ConcurrentMultiMap{}
	c.concurrentMap = make(map[string]any, capacity)
	return c
}
func (c *ConcurrentMultiMap) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.concurrentMap = make(map[string]any, 0)
}
func (c *ConcurrentMultiMap) Set(key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.concurrentMap[key] = value
}
func (c *ConcurrentMultiMap) Sets(keys []string, values []any) {
	if len(keys) != len(values) || len(keys) == 0 {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for i := range values {
		c.concurrentMap[keys[i]] = values[i]
	}
}
func (c *ConcurrentMultiMap) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.concurrentMap[key]
	return v, ok
}

func (c *ConcurrentMultiMap) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.concurrentMap, key)
}
func (c *ConcurrentMultiMap) GetMap() map[string]any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.concurrentMap
}

type ConcurrentMultiMaps struct {
	concurrentMaps map[string]*ConcurrentMultiMap
	mutex          sync.RWMutex
}

func NewConcurrentMultiMaps(mapKeys []string, capacity int) *ConcurrentMultiMaps {
	c := &ConcurrentMultiMaps{}
	c.concurrentMaps = make(map[string]*ConcurrentMultiMap, capacity)
	for _, mapKey := range mapKeys {
		c.concurrentMaps[mapKey] = NewConcurrentMultiMap(capacity)
	}
	return c
}

func (c *ConcurrentMultiMaps) AddKeyMap(mapKey string, capacity int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.concurrentMaps[mapKey] = NewConcurrentMultiMap(capacity)
}
func (c *ConcurrentMultiMaps) Clear(mapKey string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.concurrentMaps[mapKey].Clear()
}

func (c *ConcurrentMultiMaps) ClearAll() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for mapKey, _ := range c.concurrentMaps {
		c.concurrentMaps[mapKey].Clear()
	}
	c.concurrentMaps = make(map[string]*ConcurrentMultiMap, 0)
}

func (c *ConcurrentMultiMaps) Set(mapKey string, key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.concurrentMaps[mapKey].Set(key, value)
}
func (c *ConcurrentMultiMaps) Sets(mapKey string, keys []string, values []any) {
	if len(keys) != len(values) || len(keys) == 0 {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.concurrentMaps[mapKey].Sets(keys, values)
}

func (c *ConcurrentMultiMaps) Get(mapKey string, key string) (value any, exist bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.concurrentMaps[mapKey].Get(key)
}

func (c *ConcurrentMultiMaps) Remove(mapKey string, key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.concurrentMaps[mapKey].Remove(key)
}

// not change map contents
func (c *ConcurrentMultiMaps) GetMap(mapKey string) map[string]any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.concurrentMaps[mapKey].GetMap()
}
