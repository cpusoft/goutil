package queue

type Queue[T any] struct {
	values []T
}

func New[T any]() *Queue[T] {
	return &Queue[T]{}
}

func (s *Queue[T]) Size() int {
	return len(s.values)
}

func (s *Queue[T]) Enqueue(val T) {
	s.values = append(s.values, val)
}

func (s *Queue[T]) Dequeue() (T, bool) {
	if len(s.values) == 0 {
		var zero T
		return zero, false
	}
	first := s.values[0]
	s.values = s.values[1:]
	return first, true
}

func (s *Queue[T]) Peek() (T, bool) {
	if len(s.values) == 0 {
		var zero T
		return zero, false
	}
	first := s.values[0]
	return first, true
}
