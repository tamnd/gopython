package pycore

type indexHeap struct {
	values []int32
}

type IndexPool struct {
	freeIndices    indexHeap
	nextIndex      int32
	TLBCGeneration uint64
}

func (p *IndexPool) AllocIndex() int32 {
	var index int32
	if p.freeIndices.size() == 0 {
		p.freeIndices.ensureCapacity(int(p.nextIndex) + 1)
		index = p.nextIndex
		p.nextIndex++
	} else {
		index = p.freeIndices.pop()
	}
	p.TLBCGeneration++
	return index
}

func (p *IndexPool) FreeIndex(index int32) {
	p.TLBCGeneration++
	p.freeIndices.add(index)
}

func (p *IndexPool) Fini() {
	p.freeIndices.values = nil
	p.nextIndex = 0
	p.TLBCGeneration = 0
}

func (h *indexHeap) size() int {
	return len(h.values)
}

func (h *indexHeap) ensureCapacity(limit int) {
	if cap(h.values) > limit {
		return
	}
	newCapacity := cap(h.values)
	if newCapacity == 0 {
		newCapacity = 1024
	}
	for newCapacity < limit {
		newCapacity <<= 1
	}
	values := make([]int32, len(h.values), newCapacity)
	copy(values, h.values)
	h.values = values
}

func (h *indexHeap) add(val int32) {
	h.values = append(h.values, val)
	for cur := h.size() - 1; cur > 0; cur = parent(cur) {
		if !h.trySwap(cur, parent(cur)) {
			break
		}
	}
}

func (h *indexHeap) pop() int32 {
	result := h.values[0]
	last := h.values[h.size()-1]
	h.values[0] = last
	h.values = h.values[:h.size()-1]
	for cur := 0; cur < h.size(); {
		minChild := h.minChild(cur)
		if minChild > -1 && h.trySwap(cur, minChild) {
			cur = minChild
		} else {
			break
		}
	}
	return result
}

func (h *indexHeap) trySwap(i, j int) bool {
	if i < 0 || i >= h.size() || j < 0 || j >= h.size() {
		return false
	}
	if i <= j {
		if h.values[i] <= h.values[j] {
			return false
		}
	} else if h.values[j] <= h.values[i] {
		return false
	}
	h.values[i], h.values[j] = h.values[j], h.values[i]
	return true
}

func (h *indexHeap) minChild(i int) int {
	left := leftChild(i)
	right := rightChild(i)
	if left < h.size() {
		if right < h.size() {
			if h.values[left] < h.values[right] {
				return left
			}
			return right
		}
		return left
	}
	if right < h.size() {
		return right
	}
	return -1
}

func parent(i int) int {
	return (i - 1) / 2
}

func leftChild(i int) int {
	return 2*i + 1
}

func rightChild(i int) int {
	return 2*i + 2
}
