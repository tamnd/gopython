package pycore

import "testing"

func TestUniqueIDPoolAssignReleaseReuse(t *testing.T) {
	var pool UniqueIDPool
	a := pool.Assign("a")
	b := pool.Assign("b")
	if a != 1 || b != 2 {
		t.Fatalf("ids = %d, %d", a, b)
	}
	obj, ok := pool.Object(a)
	if !ok || obj != "a" {
		t.Fatalf("object = (%v, %v)", obj, ok)
	}
	pool.Release(a)
	if _, ok := pool.Object(a); ok {
		t.Fatal("released id still has object")
	}
	c := pool.Assign("c")
	if c != a {
		t.Fatalf("reused id = %d, want %d", c, a)
	}
}

func TestUniqueIDPoolResizeAndFinalize(t *testing.T) {
	var pool UniqueIDPool
	for i := 0; i < uniqueIDPoolMinSize+1; i++ {
		if got := pool.Assign(i); got != i+1 {
			t.Fatalf("id %d = %d", i, got)
		}
	}
	if len(pool.table) != uniqueIDPoolMinSize*2 {
		t.Fatalf("table size = %d", len(pool.table))
	}
	pool.Finalize()
	if len(pool.table) != 0 || pool.freelist != 0 {
		t.Fatalf("finalized table=%d freelist=%d", len(pool.table), pool.freelist)
	}
}
