package pycore

const (
	HashTableMinSize      = 16
	hashTableHighLoad     = 0.50
	hashTableLowLoad      = 0.10
	hashTableRehashFactor = 2.0 / (hashTableLowLoad + hashTableHighLoad)
)

type HashTableHashFunc func(key any) uint64
type HashTableCompareFunc func(key1, key2 any) bool
type HashTableDestroyFunc func(value any)
type HashTableForeachFunc func(ht *HashTable, key, value, userData any) int

type hashTableEntry struct {
	next    *hashTableEntry
	keyHash uint64
	key     any
	value   any
}

type HashTable struct {
	nentries         int
	nbuckets         int
	buckets          []*hashTableEntry
	hashFunc         HashTableHashFunc
	compareFunc      HashTableCompareFunc
	keyDestroyFunc   HashTableDestroyFunc
	valueDestroyFunc HashTableDestroyFunc
}

func NewHashTable(hashFunc HashTableHashFunc, compareFunc HashTableCompareFunc) *HashTable {
	return NewHashTableFull(hashFunc, compareFunc, nil, nil)
}

func NewHashTableFull(
	hashFunc HashTableHashFunc,
	compareFunc HashTableCompareFunc,
	keyDestroyFunc HashTableDestroyFunc,
	valueDestroyFunc HashTableDestroyFunc,
) *HashTable {
	return &HashTable{
		nbuckets:         HashTableMinSize,
		buckets:          make([]*hashTableEntry, HashTableMinSize),
		hashFunc:         hashFunc,
		compareFunc:      compareFunc,
		keyDestroyFunc:   keyDestroyFunc,
		valueDestroyFunc: valueDestroyFunc,
	}
}

func HashTableHashPointer(key any) uint64 {
	if ptr, ok := key.(uintptr); ok {
		return uint64(HashPointerRaw(ptr))
	}
	panic("HashTableHashPointer requires uintptr key")
}

func HashTableCompareDirect(key1, key2 any) bool {
	return key1 == key2
}

func HashTableHashString(key any) uint64 {
	return uint64(HashBuffer([]byte(key.(string))))
}

func HashTableCompareString(key1, key2 any) bool {
	return key1.(string) == key2.(string)
}

func (ht *HashTable) Len() int {
	return ht.nentries
}

func (ht *HashTable) BucketCount() int {
	return ht.nbuckets
}

func (ht *HashTable) SizeEstimate() int {
	return ht.nbuckets + ht.nentries
}

func (ht *HashTable) Set(key, value any) {
	if ht.getEntry(key) != nil {
		panic("duplicate hashtable key")
	}
	entry := &hashTableEntry{
		keyHash: ht.hashFunc(key),
		key:     key,
		value:   value,
	}
	ht.nentries++
	if float64(ht.nentries)/float64(ht.nbuckets) > hashTableHighLoad {
		ht.rehash()
	}
	index := int(entry.keyHash & uint64(ht.nbuckets-1))
	entry.next = ht.buckets[index]
	ht.buckets[index] = entry
}

func (ht *HashTable) Get(key any) (any, bool) {
	entry := ht.getEntry(key)
	if entry == nil {
		return nil, false
	}
	return entry.value, true
}

func (ht *HashTable) Steal(key any) (any, bool) {
	keyHash := ht.hashFunc(key)
	index := int(keyHash & uint64(ht.nbuckets-1))
	var previous *hashTableEntry
	entry := ht.buckets[index]
	for entry != nil {
		if entry.keyHash == keyHash && ht.compareFunc(key, entry.key) {
			break
		}
		previous = entry
		entry = entry.next
	}
	if entry == nil {
		return nil, false
	}
	if previous != nil {
		previous.next = entry.next
	} else {
		ht.buckets[index] = entry.next
	}
	ht.nentries--
	if float64(ht.nentries)/float64(ht.nbuckets) < hashTableLowLoad {
		ht.rehash()
	}
	return entry.value, true
}

func (ht *HashTable) Foreach(fn HashTableForeachFunc, userData any) int {
	for _, bucket := range ht.buckets {
		for entry := bucket; entry != nil; entry = entry.next {
			if res := fn(ht, entry.key, entry.value, userData); res != 0 {
				return res
			}
		}
	}
	return 0
}

func (ht *HashTable) Clear() {
	for i, bucket := range ht.buckets {
		for entry := bucket; entry != nil; entry = entry.next {
			ht.destroyEntry(entry)
		}
		ht.buckets[i] = nil
	}
	ht.nentries = 0
	ht.rehash()
}

func (ht *HashTable) Destroy() {
	ht.Clear()
	ht.buckets = nil
	ht.nbuckets = 0
}

func (ht *HashTable) getEntry(key any) *hashTableEntry {
	keyHash := ht.hashFunc(key)
	index := int(keyHash & uint64(ht.nbuckets-1))
	for entry := ht.buckets[index]; entry != nil; entry = entry.next {
		if entry.keyHash == keyHash && ht.compareFunc(key, entry.key) {
			return entry
		}
	}
	return nil
}

func (ht *HashTable) rehash() {
	newSize := roundHashTableSize(int(float64(ht.nentries) * hashTableRehashFactor))
	if newSize == ht.nbuckets {
		return
	}
	newBuckets := make([]*hashTableEntry, newSize)
	for _, bucket := range ht.buckets {
		for entry := bucket; entry != nil; {
			next := entry.next
			index := int(entry.keyHash & uint64(newSize-1))
			entry.next = newBuckets[index]
			newBuckets[index] = entry
			entry = next
		}
	}
	ht.nbuckets = newSize
	ht.buckets = newBuckets
}

func (ht *HashTable) destroyEntry(entry *hashTableEntry) {
	if ht.keyDestroyFunc != nil {
		ht.keyDestroyFunc(entry.key)
	}
	if ht.valueDestroyFunc != nil {
		ht.valueDestroyFunc(entry.value)
	}
}

func roundHashTableSize(size int) int {
	if size < HashTableMinSize {
		return HashTableMinSize
	}
	i := 1
	for i < size {
		i <<= 1
	}
	return i
}
