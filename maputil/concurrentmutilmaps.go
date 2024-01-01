package maputil

type ConcurrentMutilMaps struct {
	concurrentMaps map[string]ConcurrentMap[string, any]
}

func NewConcurrentMutilMaps(mapKeys []string, capacity int) *ConcurrentMutilMaps {
	c := &ConcurrentMutilMaps{}
	c.concurrentMaps = make(map[string]ConcurrentMap[string, any])
	for _, key := range mapKeys {
		c.concurrentMaps[key] = New[any](capacity)
	}
	return c
}

func (c *ConcurrentMutilMaps) AddKeyMap(mapKey string, capacity int) {
	c.concurrentMaps[mapKey] = New[any](capacity)
}
func (c *ConcurrentMutilMaps) Clear(mapKey string) {
	c.concurrentMaps[mapKey] = New[any](0)
}
func (c *ConcurrentMutilMaps) ClearAll() {
	for mapKey, _ := range c.concurrentMaps {
		c.concurrentMaps[mapKey] = New[any](0)
	}
	c.concurrentMaps = make(map[string]ConcurrentMap[string, any])
}

func (c *ConcurrentMutilMaps) Set(mapKey string, key string, value any) {
	c.concurrentMaps[mapKey].Set(key, value)
}

func (c *ConcurrentMutilMaps) Get(mapKey string, key string) (value any, exist bool) {
	return c.concurrentMaps[mapKey].Get(key)
}

func (c *ConcurrentMutilMaps) Remove(mapKey string, key string) {
	c.concurrentMaps[mapKey].Remove(key)
}

// return temp map
func (c *ConcurrentMutilMaps) GetMap(mapKey string) map[string]any {
	return c.concurrentMaps[mapKey].Items()
}
