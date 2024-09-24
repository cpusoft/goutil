package set

type Set[T comparable] struct {
	itemMap map[T]struct{}
}

func New[T comparable](elems ...T) *Set[T] {
	s := &Set[T]{}
	s.itemMap = make(map[T]struct{})
	s.Add(elems...)
	return s
}

func (s *Set[T]) Size() int {
	return len(s.itemMap)
}

func (s *Set[T]) Items() []T {
	keys := make([]T, len(s.itemMap))
	i := 0
	for k := range s.itemMap {
		keys[i] = k
		i++
	}
	return keys
}

func (s *Set[T]) Add(elems ...T) {
	for _, elem := range elems {
		s.itemMap[elem] = struct{}{}
	}
}

func (s *Set[T]) Has(elem T) bool {
	_, ok := s.itemMap[elem]
	return ok
}

func (s *Set[T]) Remove(elem T) bool {
	_, ok := s.itemMap[elem]
	if !ok {
		return false
	}
	delete(s.itemMap, elem)
	return true
}

func (s *Set[T]) Clear() {
	s.itemMap = make(map[T]struct{})
}
