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

type UrlsMutex struct {
	Urls  *list.List
	Mutex *sync.RWMutex
}

func NewUrlsMutex() *UrlsMutex {
	l := list.New()
	m := new(sync.RWMutex)
	return &UrlsMutex{
		Urls:  l,
		Mutex: m}
}
func (r *UrlsMutex) UrlsSize() int {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()
	return r.Urls.Len()
}

// queue for rsync url
type RsyncUrlQueue struct {
	WaitUrls *UrlsMutex
	CurUrls  *UrlsMutex
	UsedUrls *UrlsMutex

	Msg chan string // will trigger rsync
}

func NewQueue() *RsyncUrlQueue {
	m := make(chan string, 100)
	return &RsyncUrlQueue{
		WaitUrls: NewUrlsMutex(),
		CurUrls:  NewUrlsMutex(),
		UsedUrls: NewUrlsMutex(),
		Msg:      m}
}
func (r *RsyncUrlQueue) WaitUrlsSize() int {
	return r.WaitUrls.UrlsSize()
}
func (r *RsyncUrlQueue) CurUrlsSize() int {
	return r.CurUrls.UrlsSize()
}
func (r *RsyncUrlQueue) UsedUrlsSize() int {
	return r.UsedUrls.UrlsSize()
}
func (r *RsyncUrlQueue) AddNewUrl(url string, dest string) *list.Element {
	belogs.Debug("AddNewUrl():url", url, "    dest:", dest)
	if len(url) == 0 || len(dest) == 0 {
		return nil
	}
	r.WaitUrls.Mutex.Lock()
	r.CurUrls.Mutex.Lock()
	r.UsedUrls.Mutex.Lock()
	defer r.WaitUrls.Mutex.Unlock()
	defer r.CurUrls.Mutex.Unlock()
	defer r.UsedUrls.Mutex.Unlock()

	e := r.WaitUrls.Urls.Front()
	for e != nil {
		if strings.Contains(e.Value.(RsyncUrl).Url, url) {
			return nil
		} else {
			e = e.Next()
		}
	}
	e = r.CurUrls.Urls.Front()
	for e != nil {
		if strings.Contains(e.Value.(RsyncUrl).Url, url) {
			return nil
		} else {
			e = e.Next()
		}
	}
	e = r.UsedUrls.Urls.Front()
	for e != nil {
		if strings.Contains(e.Value.(RsyncUrl).Url, url) {
			return nil
		} else {
			e = e.Next()
		}
	}
	rsync := RsyncUrl{Url: url, Dest: dest}
	e = r.WaitUrls.Urls.PushBack(rsync)
	r.Msg <- "add"
	return e
}

func (r *RsyncUrlQueue) GetNextWaitUrls() []RsyncUrl {
	r.WaitUrls.Mutex.Lock()
	r.CurUrls.Mutex.Lock()
	defer r.WaitUrls.Mutex.Unlock()
	defer r.CurUrls.Mutex.Unlock()

	urls := make([]RsyncUrl, 0)
	var next *list.Element
	for e := r.WaitUrls.Urls.Front(); e != nil; e = next {
		next = e.Next()
		urls = append(urls, e.Value.(RsyncUrl))
		r.CurUrls.Urls.PushBack(e.Value.(RsyncUrl))
		r.WaitUrls.Urls.Remove(e)
	}

	return urls
}

func (r *RsyncUrlQueue) CurUrlsRsyncEnd(rsyncUrl RsyncUrl) {

	r.CurUrls.Mutex.Lock()
	r.UsedUrls.Mutex.Lock()
	defer r.CurUrls.Mutex.Unlock()
	defer r.UsedUrls.Mutex.Unlock()

	var next *list.Element
	for e := r.CurUrls.Urls.Front(); e != nil; e = next {
		next = e.Next()
		rsyncUrlCur := e.Value.(RsyncUrl)
		belogs.Debug("CurUrlsRsyncEnd():rsyncUrlCur", rsyncUrlCur, "    rsyncUrl:", rsyncUrl)

		if rsyncUrlCur.Url == rsyncUrl.Url {
			r.UsedUrls.Urls.PushBack(e.Value.(RsyncUrl))
			r.CurUrls.Urls.Remove(e)
			break
		}
	}
}

func (r *RsyncUrlQueue) GetWaitUrls() []RsyncUrl {
	r.WaitUrls.Mutex.Lock()
	defer r.WaitUrls.Mutex.Unlock()

	urls := make([]RsyncUrl, 0)
	e := r.WaitUrls.Urls.Front()
	for e != nil {
		urls = append(urls, e.Value.(RsyncUrl))
		e = e.Next()
	}
	return urls
}
func (r *RsyncUrlQueue) GetCurUrls() []RsyncUrl {
	r.CurUrls.Mutex.Lock()
	defer r.CurUrls.Mutex.Unlock()

	urls := make([]RsyncUrl, 0)
	e := r.CurUrls.Urls.Front()
	for e != nil {
		urls = append(urls, e.Value.(RsyncUrl))
		e = e.Next()
	}
	return urls
}
func (r *RsyncUrlQueue) GetUsedUrls() []RsyncUrl {
	r.UsedUrls.Mutex.RLock()
	defer r.UsedUrls.Mutex.RUnlock()

	urls := make([]RsyncUrl, 0)
	e := r.UsedUrls.Urls.Front()
	for e != nil {
		urls = append(urls, e.Value.(RsyncUrl))
		e = e.Next()
	}
	return urls
}
