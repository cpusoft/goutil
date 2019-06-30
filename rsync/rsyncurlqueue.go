package rsync

import (
	"container/list"
	"strings"
	"sync"
)

type RsyncUrlQueue struct {
	curUrls   *list.List
	usedUrls  *list.List
	curMutex  *sync.RWMutex
	usedMutex *sync.RWMutex

	Msg chan string
}

func NewQueue() *RsyncUrlQueue {
	cl := list.New()
	ul := list.New()
	cm := new(sync.RWMutex)
	um := new(sync.RWMutex)
	m := make(chan string)
	return &RsyncUrlQueue{
		curUrls:   cl,
		usedUrls:  ul,
		curMutex:  cm,
		usedMutex: um,
		Msg:       m}
}
func (r *RsyncUrlQueue) CurUrlsSize() int {
	r.curMutex.RLock()
	defer r.curMutex.RUnlock()
	return r.curUrls.Len()
}
func (r *RsyncUrlQueue) UsedUrlsSize() int {
	r.usedMutex.RLock()
	defer r.usedMutex.RUnlock()
	return r.usedUrls.Len()
}

func (r *RsyncUrlQueue) AddNewUrl(url string) *list.Element {

	if len(url) == 0 {
		return nil
	}

	r.curMutex.Lock()
	r.usedMutex.Lock()
	defer r.curMutex.Unlock()
	defer r.usedMutex.Unlock()

	e := r.curUrls.Front()
	for e != nil {
		if strings.Contains(e.Value.(string), url) {
			return nil
		} else {
			e = e.Next()
		}
	}
	e = r.usedUrls.Front()
	for e != nil {
		if strings.Contains(e.Value.(string), url) {
			return nil
		} else {
			e = e.Next()
		}
	}
	e = r.curUrls.PushBack(url)
	r.Msg <- "add"
	return e
}
func (r *RsyncUrlQueue) GetNextUrl() string {
	r.curMutex.Lock()
	r.usedMutex.Lock()
	defer r.curMutex.Unlock()
	defer r.usedMutex.Unlock()

	e := r.curUrls.Front()
	if e == nil {
		return ""
	}
	r.curUrls.Remove(e)
	r.usedUrls.PushBack(e.Value.(string))

	return e.Value.(string)
}

func (r *RsyncUrlQueue) GetCurUrls() []string {
	r.curMutex.RLock()
	defer r.curMutex.RUnlock()

	urls := make([]string, 0)
	e := r.curUrls.Front()
	for e != nil {
		urls = append(urls, e.Value.(string))
		e = e.Next()
	}
	return urls
}
func (r *RsyncUrlQueue) GetUsedUrls() []string {
	r.usedMutex.RLock()
	defer r.usedMutex.RUnlock()

	urls := make([]string, 0)
	e := r.usedUrls.Front()
	for e != nil {
		urls = append(urls, e.Value.(string))
		e = e.Next()
	}
	return urls
}
