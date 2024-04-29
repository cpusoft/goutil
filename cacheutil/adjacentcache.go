package cacheutil

import (
	"errors"
	"sync"
)

type AdjacentCache struct {
	datas         map[string]*AdjacentBaseCache // baseKey is parent's ski, as child's aki
	mutex         sync.RWMutex
	datasCapacity uint64
}

func NewAdjacentCache(datasCapacity uint64) *AdjacentCache {
	c := &AdjacentCache{
		datasCapacity: datasCapacity,
	}
	c.datas = make(map[string]*AdjacentBaseCache, datasCapacity)
	return c
}

// baseKey is parent's ski, as child's aki
func (c *AdjacentCache) GetAdjacentBaseCache(baseKey string) (*AdjacentBaseCache, error) {
	if baseKey == "" {
		return nil, errors.New("baseKey is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.datas[baseKey], nil
}

func (c *AdjacentCache) AddAdjacentBaseCacheByParentData(baseKey string, key string, parentData any) error {
	if baseKey == "" || key == "" || parentData == nil {
		return errors.New("baseKey, key or parentData is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.datas[baseKey]
	if !ok {
		c.datas[baseKey] = NewAdjacentBaseCache(5)
	}
	c.datas[baseKey].SetParentData(key, parentData)
	return nil
}
func (c *AdjacentCache) AddAdjacentBaseCacheByChildData(baseKey string, key string, childData any) error {
	if baseKey == "" || key == "" || childData == nil {
		return errors.New("baseKey, key or childData is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.datas[baseKey]
	if !ok {
		c.datas[baseKey] = NewAdjacentBaseCache(5)
	}
	c.datas[baseKey].SetChildData(key, childData)
	return nil
}
