package pycore

import "testing"

func TestObjectStackPushPopSize(t *testing.T) {
	var stack ObjectStack[int]
	if _, ok := stack.Pop(); ok {
		t.Fatal("empty pop succeeded")
	}
	for i := 0; i < ObjectStackChunkSize+3; i++ {
		stack.Push(i)
	}
	if stack.Size() != ObjectStackChunkSize+3 {
		t.Fatalf("size = %d", stack.Size())
	}
	for i := ObjectStackChunkSize + 2; i >= 0; i-- {
		got, ok := stack.Pop()
		if !ok || got != i {
			t.Fatalf("pop = (%d, %v), want (%d, true)", got, ok, i)
		}
	}
	if stack.Size() != 0 {
		t.Fatalf("size after pop = %d", stack.Size())
	}
}

func TestObjectStackMergeAndClear(t *testing.T) {
	var dst ObjectStack[int]
	var src ObjectStack[int]
	dst.Push(1)
	dst.Push(2)
	src.Push(3)
	src.Push(4)
	dst.Merge(&src)
	if src.Size() != 0 {
		t.Fatalf("src size = %d", src.Size())
	}
	for _, want := range []int{4, 3, 2, 1} {
		got, ok := dst.Pop()
		if !ok || got != want {
			t.Fatalf("pop = (%d, %v), want (%d, true)", got, ok, want)
		}
	}
	dst.Push(5)
	dst.Clear()
	if dst.Size() != 0 {
		t.Fatalf("clear size = %d", dst.Size())
	}
}
