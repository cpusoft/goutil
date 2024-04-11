package maputil

import (
	"sync"
)

type BaseCache struct {
	baseDatas    map[string]any
	mutex        sync.RWMutex
	baseCapacity uint64
}

func NewBaseMap(baseCapacity uint64) *BaseCache {
	c := &BaseCache{
		baseCapacity: baseCapacity,
		baseDatas:    make(map[string]any, baseCapacity),
	}
	return c
}

func (c *BaseCache) Set(key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.baseDatas[key] = value
}
func (c *BaseCache) Sets(values []any, getKey func(value any) string) {
	if len(values) == 0 || getKey == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for i := range values {
		key := getKey(i)
		c.baseDatas[key] = values[i]
	}
}
func (c *BaseCache) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.baseDatas[key]
	return v, ok
}
func (c *BaseCache) Gets() map[string]any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.baseDatas
}
func (c *BaseCache) Update(key string, value any) {
	if key == "" || value == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.baseDatas[key] = value
}

func (c *BaseCache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.baseDatas, key)
}
func (c *BaseCache) RemoveAll() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.baseDatas = make(map[string]any, c.baseCapacity)
}
