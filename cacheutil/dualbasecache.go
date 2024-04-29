package cacheutil

import (
	"errors"
	"sync"
)

type DualBaseCache struct {
	datas         map[string]any
	mutex         sync.RWMutex
	datasCapacity uint64
}

func NewDualBaseCache(datasCapacity uint64) *DualBaseCache {
	c := &DualBaseCache{
		datasCapacity: datasCapacity,
	}
	c.datas = make(map[string]any, datasCapacity)
	return c
}

func (c *DualBaseCache) Set(key string, value any) error {
	if key == "" || value == nil {
		return errors.New("key or value is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.datas == nil {
		return errors.New("datas is empty, need call NewDualBaseCache first")
	}
	c.datas[key] = value
	return nil
}

func (c *DualBaseCache) Sets(values []any, getKey func(value any) string) error {
	if len(values) == 0 || getKey == nil {
		return errors.New("values or getKey is empty")
	}

	for _, value := range values {
		key := getKey(value)
		if err := c.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}
func (c *DualBaseCache) Get(key string) (any, bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.datas == nil {
		return nil, false, errors.New("datas is empty, need call NewDualBaseCache first")
	}
	v, ok := c.datas[key]
	return v, ok, nil
}
func (c *DualBaseCache) Gets() (map[string]any, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.datas == nil {
		return nil, errors.New("datas is empty, need call NewDualBaseCache first")
	}
	return c.datas, nil
}

func (c *DualBaseCache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.datas, key)
}
func (c *DualBaseCache) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.datas = make(map[string]any, c.datasCapacity)
}
