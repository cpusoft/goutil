package cacheutil

/*
type Cache struct {
	oneCaches   map[string]*OneCache
	listCaches  map[string]*ListCache
	useOneCache bool
	mutex       sync.RWMutex
	capacity    uint64
}

func NewCacheUseOne(capacity uint64) *Cache {
	c := &Cache{
		oneCaches:   make(map[string]*OneCache, capacity),
		capacity:    capacity,
		useOneCache: true,
	}
	return c
}
func NewCacheUseList(capacity uint64) *Cache {
	c := &Cache{
		listCaches:  make(map[string]*ListCache, capacity),
		capacity:    capacity,
		useOneCache: false,
	}
	return c
}
func (c *Cache) AddCache(baseKey string, mapCapacity uint64) {
	if baseKey == "" {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.useOneCache {
		if _, ok := c.oneCaches[baseKey]; ok {
			// only add once
			return
		}
		c.oneCaches[baseKey] = NewOneCache(mapCapacity)
	} else {
		if _, ok := c.listCaches[baseKey]; ok {
			// only add once
			return
		}
		c.listCaches[baseKey] = NewListCache(mapCapacity)
	}
}

func (c *Cache) Set(baseKey string, key string, value any) {
	if baseKey == "" || key == "" || value == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.useOneCache {
		if _, ok := c.oneCaches[baseKey]; !ok {
			return
		}
		c.oneCaches[baseKey].Set(key, value)
	} else {
		if _, ok := c.listCaches[baseKey]; !ok {
			return
		}
		c.listCaches[baseKey].Set(key, value)
	}
}

func (c *Cache) Sets(baseKey string, values []any, getKey func(value any) string) {
	if baseKey == "" || len(values) == 0 || getKey == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.useOneCache {
		if _, ok := c.oneCaches[baseKey]; !ok {
			return
		}
		c.oneCaches[baseKey].Sets(values, getKey)
	} else {
		if _, ok := c.listCaches[baseKey]; !ok {
			return
		}
		c.listCaches[baseKey].Sets(values, getKey)
	}
}

func (c *Cache) Get(baseKey string, key string) (value any, values []any, exist bool) {
	if baseKey == "" || key == "" {
		return nil, false
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.useOneCache {
		oneCache, ok := c.oneCaches[baseKey]
		if !ok {
			return nil, nil, false
		}
		value, exist = oneCache.Get(key)
		return value, nil, exist
	} else {
		listCache, ok := c.listCaches[baseKey]
		if !ok {
			return nil, nil, false
		}
		values, exist = listCache.Get(key)
		return nil, values, exist
	}
}

func (c *Cache) Gets(baseKey string) (map[string]any, map[string]ListData, bool) {
	if baseKey == "" {
		return nil, false
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.useOneCache {
		oneCache, ok := c.oneCaches[baseKey]
		if !ok {
			return nil, nil, false
		}
		values := oneCache.Gets()
		return values, nil, true
	} else {
		listCache, ok := c.listCaches[baseKey]
		if !ok {
			return nil, nil, false
		}
		values := listCache.Gets()
		return nil, values, true
	}
}

func (c *Cache) GetCount(baseKey string) int {
	if baseKey == "" {
		return 0
	}
	d, ls, ok := c.Gets(baseKey)
	if ok {
		if c.useOneCache {
			return len(d)
		} else {
			return len(ls)
		}
	} else {
		return 0
	}
}

func (c *Cache) Remove(baseKey string, key string) {
	if baseKey == "" || key == "" {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.datas[baseKey]; !ok {
		return
	}
	c.datas[baseKey].Remove(key)
}

func (c *Cache) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for baseKey := range c.datas {
		c.datas[baseKey].Reset()
	}
	c.datas = make(map[string]*OneCache, c.capacity)
}
*/
