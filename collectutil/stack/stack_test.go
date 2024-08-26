package stack

import (
	"slices"
	"testing"
)

func TestIntStack(t *testing.T) {
	data := []int{2, 4, 6, 8, 10}

	s := New[int]()
	for i, v := range data {
		s.Push(v)
		if s.Size() != i+1 {
			t.Fatalf("expected stack size of %d, got %d", i+1, s.Size())
		}
	}

	slices.Reverse(data)
	for i, v := range data {
		x, ok := s.Pop()
		if x != v {
			t.Fatalf("expected to pop %d, but got %d (ok=%t)", v, x, ok)
		}
		if s.Size() != len(data)-i-1 {
			t.Fatalf("after popping %d expected size of %d, but got %d", v, len(data)-i-1, s.Size())
		}
	}

	x, ok := s.Pop()
	if x != 0 {
		t.Fatalf("expected popping from an empty stack to return a 0 value, but got %d", x)
	}
	if ok != false {
		t.Fatalf("expected popping from an empty stack to return ok=false, but got ok=%t", ok)
	}
	if s.Size() != 0 {
		t.Fatalf("after popping from an empty stack expected size of 0, but got %d", s.Size())
	}
}

func TestStringStack(t *testing.T) {
	data := []string{"red", "orange", "yellow", "green", "blue", "indigo", "violet"}

	s := New[string]()
	for i, v := range data {
		s.Push(v)
		if s.Size() != i+1 {
			t.Fatalf("expected stack size of %d, got %d", i+1, s.Size())
		}
	}

	slices.Reverse(data)
	for i, v := range data {
		x, ok := s.Pop()
		if x != v {
			t.Fatalf("expected to pop %q, but got %q (ok=%t)", v, x, ok)
		}
		if s.Size() != len(data)-i-1 {
			t.Fatalf("after popping %q expected size of %d, but got %d", v, len(data)-i-1, s.Size())
		}
	}

	x, ok := s.Pop()
	if x != "" {
		t.Fatalf("expected popping from an empty stack to return an empty string, but got %q", x)
	}
	if ok != false {
		t.Fatalf("expected popping from an empty stack to return ok=false, but got ok=%t", ok)
	}
	if s.Size() != 0 {
		t.Fatalf("after popping from an empty stack expected size of 0, but got %d", s.Size())
	}
}

func TestIntPeek(t *testing.T) {
	data := []int{2, 4, 6, 8, 10}

	s := New[int]()
	for _, v := range data {
		s.Push(v)
	}

	slices.Reverse(data)
	for i, v := range data {
		x, ok := s.Peek()
		if x != v {
			t.Fatalf("expected peek #%d to return %d, but got %d", i+1, v, x)
		}
		if ok != true {
			t.Fatalf("expected peek #%d to return ok=true, but got %t", i+1, ok)
		}
		if s.Size() != len(data)-i {
			t.Fatalf("after peeking expected size of %d, but got %d", len(data)-i, s.Size())
		}

		x, ok = s.Pop()
		if x != v {
			t.Fatalf("expected pop #%d to return %d, but got %d", i+1, v, x)
		}
		if ok != true {
			t.Fatalf("expected peek #%d to return ok=true, but got %t", i+1, ok)
		}
	}

	// empty
	x, ok := s.Peek()
	if x != 0 {
		t.Fatalf("expected peek on an empty stack to return 0, but got %d", x)
	}
	if ok != false {
		t.Fatalf("expected peek on an empty stack to return ok=false, but got %t", ok)
	}
	if s.Size() != 0 {
		t.Fatalf("after peeking expected size of 0, but got %d", s.Size())
	}

	x, ok = s.Pop()
	if x != 0 {
		t.Fatalf("expected pop on an empty stack to return 0, but got %d", x)
	}
	if ok != false {
		t.Fatalf("expected pop on an empty stack to return ok=false, but got %t", ok)
	}
}

func TestStringPeek(t *testing.T) {
	data := []string{"red", "orange", "yellow", "green", "blue", "indigo", "violet"}

	s := New[string]()
	for _, v := range data {
		s.Push(v)
	}

	slices.Reverse(data)
	for i, v := range data {
		x, ok := s.Peek()
		if x != v {
			t.Fatalf("expected peek #%d to return %q, but got %q", i+1, v, x)
		}
		if ok != true {
			t.Fatalf("expected peek #%d to return ok=true, but got %t", i+1, ok)
		}
		if s.Size() != len(data)-i {
			t.Fatalf("after peeking expected size of %d, but got %d", len(data)-i, s.Size())
		}

		x, ok = s.Pop()
		if x != v {
			t.Fatalf("expected pop #%d to return %q, but got %q", i+1, v, x)
		}
		if ok != true {
			t.Fatalf("expected peek #%d to return ok=true, but got %t", i+1, ok)
		}
	}

	// empty
	x, ok := s.Peek()
	if x != "" {
		t.Fatalf("expected peek on an empty stack to return an empty string, but got %q", x)
	}
	if ok != false {
		t.Fatalf("expected peek on an empty stack to return ok=false, but got %t", ok)
	}
	if s.Size() != 0 {
		t.Fatalf("after peeking expected size of 0, but got %d", s.Size())
	}

	x, ok = s.Pop()
	if x != "" {
		t.Fatalf("expected pop on an empty stack to return an empty string, but got %q", x)
	}
	if ok != false {
		t.Fatalf("expected pop on an empty stack to return ok=false, but got %t", ok)
	}
}
