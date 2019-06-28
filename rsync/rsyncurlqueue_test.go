package rsync

import (
	"fmt"
	"testing"
)

func TestRsyncQueue(t *testing.T) {
	rq := NewQueue()
	le := rq.Enqueue("rsync://apnic.com/1")
	s := rq.Size()
	fmt.Println(le)
	fmt.Println(s)

	rq.Enqueue("rsync://apnic.com/1")
	s = rq.Size()
	fmt.Println(s)

	rq.Enqueue("rsync://apnic.com/2")
	s = rq.Size()
	fmt.Println(s)

	rq.Enqueue("rsync://apnic.com/3")
	s = rq.Size()
	fmt.Println(s)

	url := rq.Dequeue()
	fmt.Println(url)
	s = rq.Size()
	fmt.Println(s)

	urls := rq.Traversal()
	fmt.Println(urls)

}
