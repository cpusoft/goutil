package rsyncutil

import (
	"fmt"
	"strings"
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
	s1 := rq.GetRsyncingCount()
	fmt.Println(s1)
	s2 := rq.GetRsyncUrlsLen()
	fmt.Println(s2)

	s1 = rq.SubRsyncingCount()
	fmt.Println(s1)

	fmt.Println("urls:", rq.GetRsyncUrls())

	s := ".cer"
	s = strings.Replace(s, ".", "", -1)
	fmt.Println(s)

}
