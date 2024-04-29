package cacheutil

type HorizontalCache DualCache

/*
type HorizontalCache struct {
	horizontalBaseCaches map[string]*HorizontalBaseCache
	mutex                sync.RWMutex
	capacity             uint64
}

func NewHorizontalCache(capacity uint64) *HorizontalCache {
	c := &HorizontalCache{
		horizontalBaseCaches: make(map[string]*HorizontalBaseCache, capacity),
		capacity:             capacity,
	}
	return c
}

func (c *HorizontalCache) AddBaseCache(baseKey string, datasCapacity uint64) {
	if baseKey == "" {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.horizontalBaseCaches[baseKey]; ok {
		// only add once
		return
	}
	c.horizontalBaseCaches[baseKey] = NewHorizontalBaseCache(datasCapacity)
}

func (c *HorizontalCache) Add(baseKey string, value any) error {
	if baseKey == "" || value == nil {
		return errors.New("baseKey, or value is empty")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.horizontalBaseCaches[baseKey]; !ok {
		return errors.New("not found by baseKey, call AddBaseCache first")
	}
	return c.horizontalBaseCaches[baseKey].Add(value)
}

func (c *HorizontalCache) Sets(baseKey string, values []any, getKey func(value any) string) {
	if baseKey == "" || len(values) == 0 || getKey == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.horizontalBaseCaches[baseKey]; !ok {
		return
	}
	c.horizontalBaseCaches[baseKey].Sets(values, getKey)

}

func (c *HorizontalCache) Get(baseKey string, key string) (value any, exist bool) {
	if baseKey == "" || key == "" {
		return nil, false
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	dualBaseCache, ok := c.horizontalBaseCaches[baseKey]
	if !ok {
		return nil, false
	}
	return dualBaseCache.Get(key)
}

func (c *HorizontalCache) Gets(baseKey string) (map[string]any, bool) {
	if baseKey == "" {
		return nil, false
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	dualBaseCache, ok := c.horizontalBaseCaches[baseKey]
	if !ok {
		return nil, false
	}
	values := dualBaseCache.Gets()
	return values, true
}

func (c *HorizontalCache) GetCount(baseKey string) int {
	if baseKey == "" {
		return 0
	}
	d, ok := c.Gets(baseKey)
	if ok {
		return len(d)
	} else {
		return 0
	}
}

func (c *HorizontalCache) Remove(baseKey string, key string) {
	if baseKey == "" || key == "" {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.horizontalBaseCaches[baseKey]
	if !ok {
		return
	}
	c.horizontalBaseCaches[baseKey].Remove(key)
}

func (c *HorizontalCache) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for baseKey := range c.horizontalBaseCaches {
		c.horizontalBaseCaches[baseKey].Reset()
	}
	c.horizontalBaseCaches = make(map[string]*HorizontalBaseCache, c.capacity)
}
*/
