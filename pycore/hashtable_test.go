package pycore

import (
	"slices"
	"testing"
)

func TestHashTableSetGetSteal(t *testing.T) {
	ht := NewHashTable(HashTableHashString, HashTableCompareString)
	ht.Set("a", 1)
	ht.Set("b", 2)
	if ht.Len() != 2 {
		t.Fatalf("len = %d", ht.Len())
	}
	value, ok := ht.Get("a")
	if !ok || value != 1 {
		t.Fatalf("get a = (%v, %v)", value, ok)
	}
	value, ok = ht.Steal("a")
	if !ok || value != 1 {
		t.Fatalf("steal a = (%v, %v)", value, ok)
	}
	if _, ok := ht.Get("a"); ok {
		t.Fatal("stolen key still present")
	}
	if ht.Len() != 1 {
		t.Fatalf("len after steal = %d", ht.Len())
	}
}

func TestHashTableDuplicatePanics(t *testing.T) {
	ht := NewHashTable(HashTableHashString, HashTableCompareString)
	ht.Set("a", 1)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	ht.Set("a", 2)
}

func TestHashTableRehashAndForeach(t *testing.T) {
	ht := NewHashTable(HashTableHashString, HashTableCompareString)
	for i := range 20 {
		ht.Set(string(rune('a'+i)), i)
	}
	if ht.BucketCount() <= HashTableMinSize {
		t.Fatalf("bucket count = %d", ht.BucketCount())
	}
	var keys []string
	res := ht.Foreach(func(_ *HashTable, key, _ any, _ any) int {
		keys = append(keys, key.(string))
		return 0
	}, nil)
	if res != 0 {
		t.Fatalf("foreach result = %d", res)
	}
	slices.Sort(keys)
	if len(keys) != 20 || keys[0] != "a" || keys[19] != "t" {
		t.Fatalf("keys = %#v", keys)
	}
	if stop := ht.Foreach(func(_ *HashTable, _, _ any, _ any) int {
		return 7
	}, nil); stop != 7 {
		t.Fatalf("stop = %d", stop)
	}
}

func TestHashTableClearDestroyCallbacks(t *testing.T) {
	var destroyed []any
	ht := NewHashTableFull(
		HashTableHashString,
		HashTableCompareString,
		func(value any) { destroyed = append(destroyed, "k:"+value.(string)) },
		func(value any) { destroyed = append(destroyed, value) },
	)
	ht.Set("a", 1)
	ht.Set("b", 2)
	ht.Clear()
	if ht.Len() != 0 {
		t.Fatalf("len = %d", ht.Len())
	}
	if ht.BucketCount() != HashTableMinSize {
		t.Fatalf("bucket count = %d", ht.BucketCount())
	}
	if len(destroyed) != 4 {
		t.Fatalf("destroyed = %#v", destroyed)
	}
	ht.Destroy()
	if ht.BucketCount() != 0 {
		t.Fatalf("destroy bucket count = %d", ht.BucketCount())
	}
}

func TestRoundHashTableSize(t *testing.T) {
	tests := map[int]int{
		0:  HashTableMinSize,
		1:  HashTableMinSize,
		16: HashTableMinSize,
		17: 32,
		33: 64,
	}
	for input, want := range tests {
		if got := roundHashTableSize(input); got != want {
			t.Fatalf("roundHashTableSize(%d) = %d, want %d", input, got, want)
		}
	}
}
