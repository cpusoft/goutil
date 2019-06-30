package rsync

import (
	"fmt"
	"testing"
)

func TestRsyncQueue(t *testing.T) {
	rq := NewQueue()
	le := rq.AddNewUrl("rsync://apnic.com/1")
	s := rq.CurUrlsSize()
	fmt.Println(le)
	fmt.Println(s)

	rq.AddNewUrl("rsync://apnic.com/1")
	s = rq.CurUrlsSize()
	fmt.Println(s)

	rq.AddNewUrl("rsync://apnic.com/2")
	s = rq.CurUrlsSize()
	fmt.Println(s)

	rq.AddNewUrl("rsync://apnic.com/3")
	s = rq.CurUrlsSize()
	fmt.Println(s)

	url := rq.GetNextUrl()
	fmt.Println(url)
	s = rq.CurUrlsSize()
	fmt.Println(s)

	s = rq.UsedUrlsSize()
	fmt.Println(s)

	urls := rq.GetCurUrls()
	fmt.Println(urls)

	urls = rq.GetUsedUrls()
	fmt.Println(urls)

}
