package stack

type Stack[T any] struct {
	values []T
}

func New[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Push(val T) {
	s.values = append(s.values, val)
}

func (s *Stack[T]) Size() int {
	return len(s.values)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.values) == 0 {
		var zero T
		return zero, false
	}
	top := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return top, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if len(s.values) == 0 {
		var zero T
		return zero, false
	}
	top := s.values[len(s.values)-1]
	return top, true
}
