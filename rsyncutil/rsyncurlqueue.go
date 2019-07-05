package rsyncutil

import (
	"container/list"
	"errors"
	"strings"
	"sync"
	"sync/atomic"

	belogs "github.com/astaxie/beego/logs"
)

// rsync info
type RsyncUrl struct {
	Url  string `json:"url"`
	Dest string `jsong:"dest"`
}

// queue for rsync url
type RsyncUrlQueue struct {
	Mutex *sync.Mutex
	// all rsync urls: including rsyncing and rsynced
	RsyncUrls *list.List
	// rsyncing count
	RsyncingCount int64
	// will add to RsyncUrls to rsync
	RsyncUrlChan chan RsyncUrl
}

func NewQueue() *RsyncUrlQueue {
	belogs.Debug("RsyncUrlQueue():")
	m := new(sync.Mutex)
	rsyncUrlChan := make(chan RsyncUrl, 10000)
	return &RsyncUrlQueue{
		Mutex:         m,
		RsyncUrls:     list.New(),
		RsyncingCount: 0,
		RsyncUrlChan:  rsyncUrlChan}
}
func (r *RsyncUrlQueue) GetRsyncUrlsLen() int {
	belogs.Debug("GetRsyncUrlsLen():r.Mutex.Lock()")

	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	belogs.Debug("GetRsyncUrlsLen():r.Mutex.Lock()")

	return r.RsyncUrls.Len()
}
func (r *RsyncUrlQueue) GetRsyncUrls() []RsyncUrl {
	belogs.Debug("GetRsyncUrls():r.Mutex.Lock()")

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	belogs.Debug("GetRsyncUrls():get r.Mutex.Lock(): len:", r.RsyncUrls.Len())
	urls := make([]RsyncUrl, 0)
	e := r.RsyncUrls.Front()
	for e != nil {
		urls = append(urls, e.Value.(RsyncUrl))
		e = e.Next()
	}
	return urls
}
func (r *RsyncUrlQueue) AddRsyncingCount() int64 {
	return atomic.AddInt64(&r.RsyncingCount, 1)
}
func (r *RsyncUrlQueue) SubRsyncingCount() int64 {
	return atomic.AddInt64(&r.RsyncingCount, -1)
}

func (r *RsyncUrlQueue) AddNewUrl(url string, dest string) (RsyncUrl, error) {
	belogs.Debug("AddNewUrl():url", url, "    dest:", dest)
	if len(url) == 0 || len(dest) == 0 {
		return RsyncUrl{}, errors.New("rsync url or dest is emtpy")
	}
	belogs.Debug("AddNewUrl():r.Mutex.Lock() ", url)
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	belogs.Debug("AddNewUrl():get r.Mutex.Lock() ", url)

	e := r.RsyncUrls.Front()
	for e != nil {
		if strings.Contains(url, e.Value.(RsyncUrl).Url) {
			belogs.Debug("AddNewUrl():have existed:", url, " in ", e.Value.(RsyncUrl).Url)
			return RsyncUrl{}, errors.New(url + " have existed")
		} else {
			e = e.Next()
		}
	}

	rsync := RsyncUrl{Url: url, Dest: dest}
	e = r.RsyncUrls.PushBack(rsync)
	r.RsyncUrlChan <- rsync
	atomic.AddInt64(&r.RsyncingCount, 1)
	return rsync, nil
}
