package cacheutil

import (
	"errors"
	"sync"
)

type DualCache struct {
	dualBaseCaches map[string]*DualBaseCache
	mutex          sync.RWMutex
	capacity       uint64
}

func NewDualCache(capacity uint64) *DualCache {
	c := &DualCache{
		dualBaseCaches: make(map[string]*DualBaseCache, capacity),
		capacity:       capacity,
	}
	return c
}

func (c *DualCache) AddBaseCache(baseKey string, mapCapacity uint64) error {
	if baseKey == "" {
		return errors.New("baseKey is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.dualBaseCaches == nil {
		return errors.New("dualBaseCaches is empty, need call NewDualCache first")
	}
	if _, ok := c.dualBaseCaches[baseKey]; ok {
		// only add once
		return nil
	}
	c.dualBaseCaches[baseKey] = NewDualBaseCache(mapCapacity)
	return nil
}

func (c *DualCache) Set(baseKey string, key string, value any) error {
	if baseKey == "" || key == "" || value == nil {
		return errors.New("baseKey, key or value is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.dualBaseCaches == nil {
		return errors.New("dualBaseCaches is empty, need call NewDualCache first")
	}
	if _, ok := c.dualBaseCaches[baseKey]; !ok {
		return errors.New("not found by baseKey, call AddBaseCache first")
	}
	return c.dualBaseCaches[baseKey].Set(key, value)
}

func (c *DualCache) Sets(baseKey string, values []any,
	getKey func(value any) string) error {
	if baseKey == "" || len(values) == 0 || getKey == nil {
		return errors.New("baseKey, values or getKey is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.dualBaseCaches == nil {
		return errors.New("dualBaseCaches is empty, need call NewDualCache first")
	}
	if _, ok := c.dualBaseCaches[baseKey]; !ok {
		return errors.New("not found by baseKey, call AddBaseCache first")
	}
	return c.dualBaseCaches[baseKey].Sets(values, getKey)
}

func (c *DualCache) Get(baseKey string, key string) (any, bool, error) {
	if baseKey == "" || key == "" {
		return nil, false, errors.New("baseKey, or key is empty")
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.dualBaseCaches == nil {
		return nil, false, errors.New("dualBaseCaches is empty, need call NewDualCache first")
	}
	dualBaseCache, ok := c.dualBaseCaches[baseKey]
	if !ok {
		return nil, false, errors.New("not found by baseKey, call AddBaseCache first")
	}
	return dualBaseCache.Get(key)
}

func (c *DualCache) Gets(baseKey string) (map[string]any, bool, error) {
	if baseKey == "" {
		return nil, false, errors.New("baseKey is empty")
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.dualBaseCaches == nil {
		return nil, false, errors.New("dualBaseCaches is empty, need call NewDualCache first")
	}
	dualBaseCache, ok := c.dualBaseCaches[baseKey]
	if !ok {
		return nil, false, errors.New("not found by baseKey, call AddBaseCache first")
	}
	values, err := dualBaseCache.Gets()
	return values, true, err
}

func (c *DualCache) GetCount(baseKey string) int {
	if baseKey == "" {
		return 0
	}
	d, ok, err := c.Gets(baseKey)
	if err != nil {
		return 0
	}
	if ok {
		return len(d)
	}
	return 0

}

func (c *DualCache) Remove(baseKey string, key string) error {
	if baseKey == "" || key == "" {
		return errors.New("baseKey or key is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.dualBaseCaches == nil {
		return errors.New("dualBaseCaches is empty, need call NewDualCache first")
	}
	_, ok := c.dualBaseCaches[baseKey]
	if !ok {
		return errors.New("not found by baseKey, call AddBaseCache first")
	}
	c.dualBaseCaches[baseKey].Remove(key)
	return nil
}

func (c *DualCache) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.dualBaseCaches != nil {
		for baseKey := range c.dualBaseCaches {
			c.dualBaseCaches[baseKey].Reset()
		}
	}
	c.dualBaseCaches = make(map[string]*DualBaseCache, c.capacity)
}
