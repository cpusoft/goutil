package cacheutil

import (
	"errors"
	"sync"
)

type AdjacentBaseCache struct {
	parentData    map[string]any //key: filenamepath
	childDatas    map[string]any //key: filenamepath
	mutex         sync.RWMutex
	datasCapacity uint64
}

func NewAdjacentBaseCache(datasCapacity uint64) *AdjacentBaseCache {
	c := &AdjacentBaseCache{
		datasCapacity: datasCapacity,
	}
	c.parentData = make(map[string]any, 1)
	c.childDatas = make(map[string]any, datasCapacity)
	return c
}
func (c *AdjacentBaseCache) SetParentData(key string, value any) error {
	if key == "" || value == nil {
		return errors.New("key or value is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.parentData == nil {
		return errors.New("parentData is empty, need call NewAdjacentBaseCache first")
	}
	c.parentData[key] = value
	return nil
}
func (c *AdjacentBaseCache) SetChildData(key string, value any) error {
	if key == "" || value == nil {
		return errors.New("key or value is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.childDatas == nil {
		return errors.New("childDatas is empty, need call NewAdjacentBaseCache first")
	}
	c.childDatas[key] = value
	return nil
}

func (c *AdjacentBaseCache) GetParentData() (map[string]any, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.parentData == nil {
		return nil, errors.New("parentData is empty, need call NewAdjacentBaseCache first")
	}
	return c.parentData, nil
}
func (c *AdjacentBaseCache) GetChildData(key string) (any, bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.childDatas == nil {
		return nil, false, errors.New("childDatas is empty, need call NewAdjacentBaseCache first")
	}
	v, ok := c.childDatas[key]
	return v, ok, nil
}

func (c *AdjacentBaseCache) GetChildDatas() (map[string]any, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.childDatas == nil {
		return nil, errors.New("childDatas is empty, need call NewAdjacentBaseCache first")
	}
	return c.childDatas, nil
}
