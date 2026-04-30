package pycore

import "testing"

func TestArenaMallocAlignmentAndBlocks(t *testing.T) {
	arena := NewArena()
	p := arena.Malloc(1)
	if len(p) != ArenaAlignment {
		t.Fatalf("len = %d", len(p))
	}
	p[0] = 7
	if arena.BlockCount() != 1 {
		t.Fatalf("block count = %d", arena.BlockCount())
	}
	big := arena.Malloc(ArenaDefaultBlockSize + 1)
	if len(big) != sizeRoundUp(ArenaDefaultBlockSize+1, ArenaAlignment) {
		t.Fatalf("big len = %d", len(big))
	}
	if arena.BlockCount() != 2 {
		t.Fatalf("block count after big alloc = %d", arena.BlockCount())
	}
}

func TestArenaAddObjectAndFree(t *testing.T) {
	arena := NewArena()
	arena.AddObject("obj")
	if arena.ObjectCount() != 1 {
		t.Fatalf("object count = %d", arena.ObjectCount())
	}
	arena.Free()
	if arena.ObjectCount() != 0 || arena.BlockCount() != 0 {
		t.Fatalf("free left objects=%d blocks=%d", arena.ObjectCount(), arena.BlockCount())
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic after free")
		}
	}()
	arena.Malloc(1)
}

func TestSizeRoundUp(t *testing.T) {
	tests := map[int]int{
		0:  0,
		1:  8,
		8:  8,
		9:  16,
		15: 16,
		16: 16,
	}
	for input, want := range tests {
		if got := sizeRoundUp(input, ArenaAlignment); got != want {
			t.Fatalf("sizeRoundUp(%d) = %d, want %d", input, got, want)
		}
	}
}
