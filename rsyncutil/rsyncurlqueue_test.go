package rsyncutil

import (
	"fmt"
	"testing"
)

func TestRsyncQueue(t *testing.T) {
	rq := NewQueue()
	tmp1, err := rq.AddNewUrl("rsync://apnic.com/1", "/tmp1")
	fmt.Println(tmp1, err)

	tmp1, err = rq.AddNewUrl("rsync://apnic.com/1", "/tmp1")
	fmt.Println(tmp1, err)

	tmp1, err = rq.AddNewUrl("rsync://apnic.com/1/zz", "/tmp1")
	fmt.Println(tmp1, err)

	tmp2, err := rq.AddNewUrl("rsync://apnic.com/2", "/tmp2")
	fmt.Println(tmp2, err)

	tmp3, err := rq.AddNewUrl("rsync://apnic.com/3", "/tmp3")
	fmt.Println(tmp3, err)

	s := rq.GetRsyncUrlsLen()
	fmt.Println(s)
	fmt.Println("urls:", rq.GetRsyncUrls())

}
