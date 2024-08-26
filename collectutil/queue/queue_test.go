package queue

import (
	"testing"
)

func TestIntQueue(t *testing.T) {
	data := []int{2, 4, 6, 8, 10}

	q := New[int]()
	for i, v := range data {
		q.Enqueue(v)
		if q.Size() != i+1 {
			t.Fatalf("expected queue size of %d, got %d", i+1, q.Size())
		}
	}

	for i, v := range data {
		x, ok := q.Dequeue()
		if x != v {
			t.Fatalf("expected to dequeue %d, but got %d (ok=%t)", v, x, ok)
		}
		if q.Size() != len(data)-i-1 {
			t.Fatalf("after dequeuing %d expected size of %d, but got %d", v, len(data)-i-1, q.Size())
		}
	}

	x, ok := q.Dequeue()
	if x != 0 {
		t.Fatalf("expected dequeining from an empty queue to return 0, but got %d", x)
	}
	if ok != false {
		t.Fatalf("expected dequeing from an empty queue to return ok=false, but got ok=%t", ok)
	}
	if q.Size() != 0 {
		t.Fatalf("after dequeuing from an empty queue expected size of 0, but got %d", q.Size())
	}
}

func TestStringQueue(t *testing.T) {
	data := []string{"red", "orange", "yellow", "green", "blue", "indigo", "violet"}

	q := New[string]()
	for i, v := range data {
		q.Enqueue(v)
		if q.Size() != i+1 {
			t.Fatalf("expected queue size of %d, got %d", i+1, q.Size())
		}
	}

	for i, v := range data {
		x, ok := q.Dequeue()
		if x != v {
			t.Fatalf("expected to dequeue %q, but got %q (ok=%t)", v, x, ok)
		}
		if q.Size() != len(data)-i-1 {
			t.Fatalf("after dequeuing %q expected size of %d, but got %d", v, len(data)-i-1, q.Size())
		}
	}

	x, ok := q.Dequeue()
	if x != "" {
		t.Fatalf("expected dequeuing from an empty queue to return an empty string, but got %q", x)
	}
	if ok != false {
		t.Fatalf("expected dequeuing from an empty queue to return ok=false, but got ok=%t", ok)
	}
	if q.Size() != 0 {
		t.Fatalf("after dequeuing from an empty queue expected size of 0, but got %d", q.Size())
	}
}

func TestIntPeek(t *testing.T) {
	data := []int{2, 4, 6, 8, 10}

	q := New[int]()
	for _, v := range data {
		q.Enqueue(v)
	}

	for i, v := range data {
		x, ok := q.Peek()
		if x != v {
			t.Fatalf("expected peek #%d to return %d, but got %d", i+1, v, x)
		}
		if ok != true {
			t.Fatalf("expected peek #%d to return ok=true, but got %t", i+1, ok)
		}
		if q.Size() != len(data)-i {
			t.Fatalf("after peeking expected size of %d, but got %d", len(data)-i, q.Size())
		}

		x, ok = q.Dequeue()
		if x != v {
			t.Fatalf("expected dequeue #%d to return %d, but got %d", i+1, v, x)
		}
		if ok != true {
			t.Fatalf("expected dequeue #%d to return ok=true, but got %t", i+1, ok)
		}
	}

	// empty
	x, ok := q.Peek()
	if x != 0 {
		t.Fatalf("expected peek on an empty queue to return 0, but got %d", x)
	}
	if ok != false {
		t.Fatalf("expected peek on an empty queue to return ok=false, but got %t", ok)
	}
	if q.Size() != 0 {
		t.Fatalf("after peeking expected size of 0, but got %d", q.Size())
	}

	x, ok = q.Dequeue()
	if x != 0 {
		t.Fatalf("expected dequeue on an empty queue to return 0, but got %d", x)
	}
	if ok != false {
		t.Fatalf("expected dequeue on an empty queue to return ok=false, but got %t", ok)
	}
}

func TestStringPeek(t *testing.T) {
	data := []string{"red", "orange", "yellow", "green", "blue", "indigo", "violet"}

	q := New[string]()
	for _, v := range data {
		q.Enqueue(v)
	}

	for i, v := range data {
		x, ok := q.Peek()
		if x != v {
			t.Fatalf("expected peek #%d to return %q, but got %q", i+1, v, x)
		}
		if ok != true {
			t.Fatalf("expected peek #%d to return ok=true, but got %t", i+1, ok)
		}
		if q.Size() != len(data)-i {
			t.Fatalf("after peeking expected size of %d, but got %d", len(data)-i, q.Size())
		}

		x, ok = q.Dequeue()
		if x != v {
			t.Fatalf("expected dequeue #%d to return %q, but got %q", i+1, v, x)
		}
		if ok != true {
			t.Fatalf("expected dequeue #%d to return ok=true, but got %t", i+1, ok)
		}
	}

	// empty
	x, ok := q.Peek()
	if x != "" {
		t.Fatalf("expected peek on an empty queue to return an empty string, but got %q", x)
	}
	if ok != false {
		t.Fatalf("expected peek on an empty queue to return ok=false, but got %t", ok)
	}
	if q.Size() != 0 {
		t.Fatalf("after peeking expected size of 0, but got %d", q.Size())
	}

	x, ok = q.Dequeue()
	if x != "" {
		t.Fatalf("expected dequeue on an empty queue to return an empty string, but got %q", x)
	}
	if ok != false {
		t.Fatalf("expected dequeue on an empty queue to return ok=false, but got %t", ok)
	}
}
