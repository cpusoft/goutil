package rsync

import (
	"container/list"
	"sync"
)

type RsyncUrlQueue struct {
	list  *list.List
	mutex *sync.RWMutex
}

func NewQueue() *RsyncUrlQueue {
	ls := list.New()
	m := new(sync.RWMutex)
	return &RsyncUrlQueue{
		list:  ls,
		mutex: m}
}
func (r *RsyncUrlQueue) Size() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.list.Len()
}
func (r *RsyncUrlQueue) Enqueue(url string) *list.Element {

	if len(url) == 0 {
		return nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	e := r.list.Front()

	for e != nil {
		if e.Value.(string) == url {
			return nil
		} else {
			e = e.Next()
		}
	}

	e = r.list.PushBack(url)
	return e
}
func (r *RsyncUrlQueue) Dequeue() string {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	e := r.list.Front()
	if e == nil {
		return ""
	}
	r.list.Remove(e)
	return e.Value.(string)
}

func (r *RsyncUrlQueue) Traversal() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	urls := make([]string, 0)
	e := r.list.Front()
	for e != nil {
		urls = append(urls, e.Value.(string))
		e = e.Next()
	}
	return urls
}
