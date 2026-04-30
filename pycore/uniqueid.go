package pycore

const InvalidUniqueID = 0
const uniqueIDPoolMinSize = 8

type uniqueIDEntry struct {
	obj  any
	next int
}

type UniqueIDPool struct {
	table    []uniqueIDEntry
	freelist int
}

func (p *UniqueIDPool) Assign(obj any) int {
	if p.freelist == 0 {
		p.resize()
	}
	idx := p.freelist - 1
	entry := &p.table[idx]
	p.freelist = entry.next
	entry.next = 0
	entry.obj = obj
	return idx + 1
}

func (p *UniqueIDPool) Release(uniqueID int) {
	if uniqueID <= 0 || uniqueID > len(p.table) {
		panic("invalid unique id")
	}
	idx := uniqueID - 1
	p.table[idx].obj = nil
	p.table[idx].next = p.freelist
	p.freelist = uniqueID
}

func (p *UniqueIDPool) Object(uniqueID int) (any, bool) {
	if uniqueID <= 0 || uniqueID > len(p.table) {
		return nil, false
	}
	obj := p.table[uniqueID-1].obj
	return obj, obj != nil
}

func (p *UniqueIDPool) Finalize() {
	for i := range p.table {
		p.table[i] = uniqueIDEntry{}
	}
	p.table = nil
	p.freelist = 0
}

func (p *UniqueIDPool) resize() {
	oldSize := len(p.table)
	newSize := oldSize * 2
	if newSize < uniqueIDPoolMinSize {
		newSize = uniqueIDPoolMinSize
	}
	table := make([]uniqueIDEntry, newSize)
	copy(table, p.table)
	for i := oldSize; i < newSize-1; i++ {
		table[i].next = i + 2
	}
	table[newSize-1].next = 0
	p.table = table
	p.freelist = oldSize + 1
}
