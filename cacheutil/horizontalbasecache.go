package cacheutil

/*
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

func (c *HorizontalBaseCache) Add(key string, value any) error {
	if key == "" || value == nil {
		return errors.New("key or value is emtpy")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.datas == nil {
		return errors.New("datas is emtpy, call NewHorizontalBaseCache first")
	}
	c.datas[key] = value
	return nil
}

func (c *HorizontalBaseCache) Adds(values []any) error {
	if len(values) == 0 {
		return errors.New("values is emtpy")
	}

	for _, value := range values {
		if err := c.Add(value); err != nil {
			return err
		}
	}
	return nil
}
func (c *HorizontalBaseCache) Get(key string) (ListData, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.datas[key]
	return v, ok
}
func (c *HorizontalBaseCache) Gets() map[string]ListData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.datas
}

func (c *HorizontalBaseCache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.datas, key)
}
func (c *HorizontalBaseCache) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.datas = make(map[string]ListData, c.datasCapacity)
}
*/
