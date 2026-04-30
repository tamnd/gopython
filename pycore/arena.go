package pycore

const (
	ArenaDefaultBlockSize = 8192
	ArenaAlignment        = 8
)

type arenaBlock struct {
	size   int
	offset int
	next   *arenaBlock
	mem    []byte
}

type Arena struct {
	head    *arenaBlock
	cur     *arenaBlock
	objects []any
	freed   bool
}

func NewArena() *Arena {
	head := newArenaBlock(ArenaDefaultBlockSize)
	return &Arena{
		head: head,
		cur:  head,
	}
}

func (a *Arena) Free() {
	a.head = nil
	a.cur = nil
	a.objects = nil
	a.freed = true
}

func (a *Arena) Malloc(size int) []byte {
	if a.freed {
		panic("arena is freed")
	}
	p := blockAlloc(a.cur, size)
	if a.cur.next != nil {
		a.cur = a.cur.next
	}
	return p
}

func (a *Arena) AddObject(obj any) {
	if a.freed {
		panic("arena is freed")
	}
	a.objects = append(a.objects, obj)
}

func (a *Arena) ObjectCount() int {
	return len(a.objects)
}

func (a *Arena) BlockCount() int {
	count := 0
	for block := a.head; block != nil; block = block.next {
		count++
	}
	return count
}

func newArenaBlock(size int) *arenaBlock {
	return &arenaBlock{
		size: size,
		mem:  make([]byte, size),
	}
}

func blockAlloc(block *arenaBlock, size int) []byte {
	size = sizeRoundUp(size, ArenaAlignment)
	if block.offset+size > block.size {
		newSize := ArenaDefaultBlockSize
		if size > newSize {
			newSize = size
		}
		block.next = newArenaBlock(newSize)
		block = block.next
	}
	p := block.mem[block.offset : block.offset+size]
	block.offset += size
	return p
}

func sizeRoundUp(size, alignment int) int {
	return (size + alignment - 1) & ^(alignment - 1)
}
