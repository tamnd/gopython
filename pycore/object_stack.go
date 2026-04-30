package pycore

const ObjectStackChunkSize = 254

type ObjectStackChunk[T any] struct {
	prev *ObjectStackChunk[T]
	objs []T
}

type ObjectStack[T any] struct {
	head *ObjectStackChunk[T]
}

func NewObjectStackChunk[T any]() *ObjectStackChunk[T] {
	return &ObjectStackChunk[T]{
		objs: make([]T, 0, ObjectStackChunkSize),
	}
}

func (s *ObjectStack[T]) Push(obj T) {
	buf := s.head
	if buf == nil || len(buf.objs) == ObjectStackChunkSize {
		buf = NewObjectStackChunk[T]()
		buf.prev = s.head
		s.head = buf
	}
	buf.objs = append(buf.objs, obj)
}

func (s *ObjectStack[T]) Pop() (T, bool) {
	var zero T
	buf := s.head
	if buf == nil {
		return zero, false
	}
	n := len(buf.objs)
	obj := buf.objs[n-1]
	buf.objs = buf.objs[:n-1]
	if len(buf.objs) == 0 {
		s.head = buf.prev
	}
	return obj, true
}

func (s *ObjectStack[T]) Size() int {
	size := 0
	for buf := s.head; buf != nil; buf = buf.prev {
		size += len(buf.objs)
	}
	return size
}

func (s *ObjectStack[T]) Clear() {
	for s.head != nil {
		s.head.objs = s.head.objs[:0]
		s.head = s.head.prev
	}
}

func (s *ObjectStack[T]) Merge(src *ObjectStack[T]) {
	if src.head == nil {
		return
	}
	if s.head != nil {
		last := src.head
		for last.prev != nil {
			last = last.prev
		}
		last.prev = s.head
	}
	s.head = src.head
	src.head = nil
}
