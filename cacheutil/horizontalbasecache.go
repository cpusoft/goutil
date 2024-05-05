package cacheutil

import (
	"errors"
	"sync"
)

// key: filepathnname
type HorizontalBaseCache struct {
	datas         map[string]any
	mutex         sync.RWMutex
	datasCapacity uint64
}

func NewHorizontalBaseCache(datasCapacity uint64) *HorizontalBaseCache {
	c := &HorizontalBaseCache{
		datasCapacity: datasCapacity,
	}
	c.datas = make(map[string]any, datasCapacity)
	return c
}
func (c *HorizontalBaseCache) Set(key string, value any) error {
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

// unsafe, be careful
func (c *HorizontalBaseCache) GetsUnsafe() (map[string]any, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.datas == nil {
		return nil, errors.New("datas is empty, need call NewDualBaseCache first")
	}
	return c.datas, nil
}
func (c *HorizontalBaseCache) GetsClone() (map[string]any, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.datas == nil {
		return nil, errors.New("datas is empty, need call NewDualBaseCache first")
	}
	n := make(map[string]any, len(c.datas))
	for k, v := range c.datas {
		n[k] = v
	}
	return n, nil
}

func (c *HorizontalBaseCache) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.datas = make(map[string]any, c.datasCapacity)
}
