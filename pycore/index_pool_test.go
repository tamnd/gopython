package pycore

import "testing"

func TestIndexPoolAllocFree(t *testing.T) {
	var pool IndexPool
	if got := pool.AllocIndex(); got != 0 {
		t.Fatalf("first index = %d", got)
	}
	if got := pool.AllocIndex(); got != 1 {
		t.Fatalf("second index = %d", got)
	}
	pool.FreeIndex(1)
	pool.FreeIndex(0)
	if got := pool.AllocIndex(); got != 0 {
		t.Fatalf("reused first = %d", got)
	}
	if got := pool.AllocIndex(); got != 1 {
		t.Fatalf("reused second = %d", got)
	}
	if pool.TLBCGeneration != 6 {
		t.Fatalf("generation = %d", pool.TLBCGeneration)
	}
}

func TestIndexPoolFini(t *testing.T) {
	var pool IndexPool
	pool.AllocIndex()
	pool.FreeIndex(0)
	pool.Fini()
	if pool.TLBCGeneration != 0 {
		t.Fatalf("generation = %d", pool.TLBCGeneration)
	}
	if got := pool.AllocIndex(); got != 0 {
		t.Fatalf("index after fini = %d", got)
	}
}
