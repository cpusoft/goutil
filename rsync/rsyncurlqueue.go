package rsync

import (
	"container/list"
	"strings"
	"sync"

	belogs "github.com/astaxie/beego/logs"
)

type RsyncUrl struct {
	Url  string `json:"url"`
	Dest string `jsong:"dest"`
}

// queue for rsync url
type RsyncUrlQueue struct {
	curUrls   *list.List
	usedUrls  *list.List
	curMutex  *sync.RWMutex
	usedMutex *sync.RWMutex

	Msg chan string // will trigger rsync
}

func NewQueue() *RsyncUrlQueue {
	cl := list.New()
	ul := list.New()
	cm := new(sync.RWMutex)
	um := new(sync.RWMutex)
	m := make(chan string, 10)
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

func (r *RsyncUrlQueue) AddNewUrl(url string, dest string) *list.Element {
	belogs.Debug("AddNewUrl():url", url, "    dest:", dest)
	if len(url) == 0 || len(dest) == 0 {
		return nil
	}

	r.curMutex.Lock()
	r.usedMutex.Lock()
	defer r.curMutex.Unlock()
	defer r.usedMutex.Unlock()

	e := r.curUrls.Front()
	for e != nil {
		if strings.Contains(e.Value.(RsyncUrl).Url, url) {
			return nil
		} else {
			e = e.Next()
		}
	}
	e = r.usedUrls.Front()
	for e != nil {
		if strings.Contains(e.Value.(RsyncUrl).Url, url) {
			return nil
		} else {
			e = e.Next()
		}
	}
	rsync := RsyncUrl{Url: url, Dest: dest}
	e = r.curUrls.PushBack(rsync)
	r.Msg <- "add"
	return e
}
func (r *RsyncUrlQueue) GetNextUrl() RsyncUrl {
	r.curMutex.Lock()
	r.usedMutex.Lock()
	defer r.curMutex.Unlock()
	defer r.usedMutex.Unlock()

	e := r.curUrls.Front()
	if e == nil {
		return RsyncUrl{}
	}
	r.curUrls.Remove(e)
	r.usedUrls.PushBack(e.Value.(RsyncUrl))

	return e.Value.(RsyncUrl)
}

func (r *RsyncUrlQueue) GetNextUrls() []RsyncUrl {
	r.curMutex.Lock()
	r.usedMutex.Lock()
	defer r.curMutex.Unlock()
	defer r.usedMutex.Unlock()

	urls := make([]RsyncUrl, 0)
	var next *list.Element
	for e := r.curUrls.Front(); e != nil; e = next {
		next = e.Next()
		urls = append(urls, e.Value.(RsyncUrl))
		r.usedUrls.PushBack(e.Value.(RsyncUrl))
		r.curUrls.Remove(e)

	}

	return urls
}

func (r *RsyncUrlQueue) GetCurUrls() []RsyncUrl {
	r.curMutex.RLock()
	defer r.curMutex.RUnlock()

	urls := make([]RsyncUrl, 0)
	e := r.curUrls.Front()
	for e != nil {
		urls = append(urls, e.Value.(RsyncUrl))
		e = e.Next()
	}
	return urls
}
func (r *RsyncUrlQueue) GetUsedUrls() []RsyncUrl {
	r.usedMutex.RLock()
	defer r.usedMutex.RUnlock()

	urls := make([]RsyncUrl, 0)
	e := r.usedUrls.Front()
	for e != nil {
		urls = append(urls, e.Value.(RsyncUrl))
		e = e.Next()
	}
	return urls
}
