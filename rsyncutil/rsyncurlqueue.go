package rsyncutil

import (
	"container/list"
	"errors"
	"strings"
	"sync"

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
	m := new(sync.Mutex)
	rsyncUrlChan := make(chan RsyncUrl, 90000)
	return &RsyncUrlQueue{
		Mutex:         m,
		RsyncUrls:     list.New(),
		RsyncingCount: 0,
		RsyncUrlChan:  rsyncUrlChan}
}
func (r *RsyncUrlQueue) GetRsyncUrlsLen() int {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	belogs.Debug("GetRsyncUrlsLen():r.Mutex.Lock(): len:", r.RsyncUrls.Len())

	return r.RsyncUrls.Len()
}
func (r *RsyncUrlQueue) GetRsyncUrls() []RsyncUrl {
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

func (r *RsyncUrlQueue) GetRsyncingCount() int64 {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	belogs.Debug("GetRsyncingCount():RsyncingCount:", r.RsyncingCount)
	return r.RsyncingCount

}
func (r *RsyncUrlQueue) AddRsyncingCount() int64 {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.RsyncingCount = r.RsyncingCount + 1
	belogs.Debug("AddRsyncingCount():RsyncingCount:", r.RsyncingCount)
	return r.RsyncingCount
}
func (r *RsyncUrlQueue) SubRsyncingCount() int64 {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.RsyncingCount = r.RsyncingCount - 1
	belogs.Debug("SubRsyncingCount():RsyncingCount:", r.RsyncingCount)
	return r.RsyncingCount
}

func (r *RsyncUrlQueue) AddNewUrl(url string, dest string) (RsyncUrl, error) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	belogs.Debug("AddNewUrl():url", url, "    dest:", dest)
	if len(url) == 0 || len(dest) == 0 {
		return RsyncUrl{}, errors.New("rsync url or dest is emtpy")
	}

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
	r.RsyncingCount = r.RsyncingCount + 1
	belogs.Debug("AddNewUrl():RsyncingCount:", r.RsyncingCount)
	r.RsyncUrlChan <- rsync

	return rsync, nil
}
