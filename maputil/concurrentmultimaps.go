package maputil

type ConcurrentMultiMaps struct {
	concurrentMaps map[string]ConcurrentMap[string, any]
}

func NewConcurrentMultiMaps(mapKeys []string, capacity int) *ConcurrentMultiMaps {
	c := &ConcurrentMultiMaps{}
	c.concurrentMaps = make(map[string]ConcurrentMap[string, any])
	for _, key := range mapKeys {
		c.concurrentMaps[key] = New[any](capacity)
	}
	return c
}

func (c *ConcurrentMultiMaps) AddKeyMap(mapKey string, capacity int) {
	c.concurrentMaps[mapKey] = New[any](capacity)
}
func (c *ConcurrentMultiMaps) Clear(mapKey string) {
	c.concurrentMaps[mapKey] = New[any](0)
}
func (c *ConcurrentMultiMaps) ClearAll() {
	for mapKey, _ := range c.concurrentMaps {
		c.concurrentMaps[mapKey] = New[any](0)
	}
	c.concurrentMaps = make(map[string]ConcurrentMap[string, any])
}

func (c *ConcurrentMultiMaps) Set(mapKey string, key string, value any) {
	c.concurrentMaps[mapKey].Set(key, value)
}

func (c *ConcurrentMultiMaps) Get(mapKey string, key string) (value any, exist bool) {
	return c.concurrentMaps[mapKey].Get(key)
}

func (c *ConcurrentMultiMaps) Remove(mapKey string, key string) {
	c.concurrentMaps[mapKey].Remove(key)
}

// return temp map
func (c *ConcurrentMultiMaps) GetMap(mapKey string) map[string]any {
	return c.concurrentMaps[mapKey].Items()
}
