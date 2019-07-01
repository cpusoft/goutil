package rsync

import (
	"fmt"
	"testing"
)

func TestRsyncQueue(t *testing.T) {
	rq := NewQueue()
	rq.AddNewUrl("rsync://apnic.com/1", "/tmp1")
	rq.AddNewUrl("rsync://apnic.com/1", "/tmp1")
	rq.AddNewUrl("rsync://apnic.com/2", "/tmp2")
	rq.AddNewUrl("rsync://apnic.com/3", "/tmp3")
	s := rq.CurUrlsSize()
	fmt.Println(s)
	fmt.Println("wait:", rq.WaitUrlsSize(), "    cur:", rq.CurUrlsSize(), "   used:", rq.UsedUrlsSize())

	urls := rq.GetNextWaitUrls()
	fmt.Println(urls)
	fmt.Println("wait:", rq.WaitUrlsSize(), "    cur:", rq.CurUrlsSize(), "   used:", rq.UsedUrlsSize())

	rq.AddNewUrl("rsync://apnic.com/new1", "/new1")
	rq.AddNewUrl("rsync://apnic.com/new2", "/new2")

	rq.CurUrlsRsyncEnd(RsyncUrl{Url: "rsync://apnic.com/2", Dest: "/tmp2"})
	fmt.Println("wait:", rq.WaitUrlsSize(), "    cur:", rq.CurUrlsSize(), "   used:", rq.UsedUrlsSize())
	fmt.Println("wait:", rq.GetWaitUrls(), "    cur:", rq.GetCurUrls(), "   used:", rq.GetUsedUrls())
}
