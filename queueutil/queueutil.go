package queueutil

import (
	"container/list"
	"sync"
)

type Queue struct {
	lock sync.RWMutex
	list *list.List
}

func NewQueue() *Queue {
	list := list.New()
	return &Queue{list: list}
}

func (c *Queue) PushFront(value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.list.PushFront(value)
}
func (c *Queue) PushBack(value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.list.PushBack(value)
}
func (c *Queue) PopFront() interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	e := c.list.Front()
	if e != nil {
		c.list.Remove(e)
		return e.Value
	}
	return nil
}

func (c *Queue) PopBack() interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	e := c.list.Back()
	if e != nil {
		c.list.Remove(e)
		return e.Value
	}
	return nil
}
func (c *Queue) PeakFront() interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	e := c.list.Front()
	if e != nil {
		return e.Value
	}
	return nil
}
func (c *Queue) PeakBack() interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	e := c.list.Back()
	if e != nil {
		return e.Value
	}
	return nil
}

func (c *Queue) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.list.Len()
}

func (c *Queue) Empty() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.list.Len() == 0
}
