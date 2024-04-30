package cacheutil

import (
	"errors"
	"sync"
)

type AdjacentCache struct {
	adjacentBaseCaches map[string]*AdjacentBaseCache // baseKey is parent's ski, as child's aki
	mutex              sync.RWMutex
	datasCapacity      uint64
}

func NewAdjacentCache(datasCapacity uint64) *AdjacentCache {
	c := &AdjacentCache{
		datasCapacity: datasCapacity,
	}
	c.adjacentBaseCaches = make(map[string]*AdjacentBaseCache, datasCapacity)
	return c
}

func (c *AdjacentCache) AddParentData(getBaseKey func(value any) string,
	values []any, getKey func(value any) string) error {
	if getBaseKey == nil || len(values) == 0 || getKey == nil {
		return errors.New("getBaseKey, values, or getKey is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.adjacentBaseCaches == nil {
		return errors.New("adjacentBaseCaches is empty, need call NewAdjacentCache first")
	}

	for _, value := range values {
		baseKey := getBaseKey(value)
		_, ok := c.adjacentBaseCaches[baseKey]
		if !ok {
			c.adjacentBaseCaches[baseKey] = NewAdjacentBaseCache(5)
		}
		key := getKey(value)
		c.adjacentBaseCaches[baseKey].SetParentData(key, value)
	}
	return nil
}

func (c *AdjacentCache) AddChildData(getBaseKey func(value any) string,
	values []any, getKey func(value any) string) error {
	if getBaseKey == nil || len(values) == 0 || getKey == nil {
		return errors.New("getBaseKey, values, or getKey is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.adjacentBaseCaches == nil {
		return errors.New("adjacentBaseCaches is empty, need call NewAdjacentCache first")
	}

	for _, value := range values {
		baseKey := getBaseKey(value)
		_, ok := c.adjacentBaseCaches[baseKey]
		if !ok {
			c.adjacentBaseCaches[baseKey] = NewAdjacentBaseCache(5)
		}
		key := getKey(value)
		c.adjacentBaseCaches[baseKey].SetChildData(key, value)
	}
	return nil
}

// baseKey is parent's ski, as child's aki
func (c *AdjacentCache) GetBaseCache(baseKey string) (*AdjacentBaseCache, bool, error) {
	if baseKey == "" {
		return nil, false, errors.New("baseKey is empty")
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.adjacentBaseCaches == nil {
		return nil, false, errors.New("adjacentBaseCaches is empty, need call NewAdjacentCache first")
	}
	v, ok := c.adjacentBaseCaches[baseKey]
	return v, ok, nil
}

func (c *AdjacentCache) GetCounts() int {
	return len(c.adjacentBaseCaches)
}
