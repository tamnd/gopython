package pycore

import (
	"math"
	"testing"
)

func TestHashDouble(t *testing.T) {
	tests := []struct {
		value float64
		want  int64
	}{
		{0, 0},
		{1, 1},
		{-1, -2},
		{1.5, 1152921504606846977},
		{math.Inf(1), PyHashInf},
		{math.Inf(-1), -PyHashInf},
	}
	for _, tt := range tests {
		if got := HashDouble(tt.value, 123); got != tt.want {
			t.Fatalf("HashDouble(%v) = %d, want %d", tt.value, got, tt.want)
		}
	}
	if got := HashDouble(math.NaN(), 123); got != 123 {
		t.Fatalf("nan hash = %d", got)
	}
}

func TestHashPointer(t *testing.T) {
	if got := HashPointer(0x100); got != HashPointerRaw(0x100) {
		t.Fatalf("pointer hash = %d", got)
	}
	if got := HashPointer(^uintptr(0)); got == -1 {
		t.Fatal("HashPointer returned reserved -1")
	}
}

func TestHashBufferAndKeyedHash(t *testing.T) {
	old := PyHashSecret
	defer func() { PyHashSecret = old }()
	PyHashSecret = HashSecret{K0: 0, K1: 0}

	if got := HashBuffer(nil); got != 0 {
		t.Fatalf("empty hash = %d", got)
	}
	if got := HashBuffer([]byte("abc")); got != -4594863902769663758 {
		t.Fatalf("abc hash = %d", got)
	}
	if got := KeyedHash(42, []byte("abc")); got != 8294136573837871780 {
		t.Fatalf("keyed hash = %d", got)
	}
}

func TestHashGetFuncDef(t *testing.T) {
	def := HashGetFuncDef()
	if def.Name != "siphash13" || def.HashBits != 64 || def.SeedBits != 128 {
		t.Fatalf("func def = %#v", def)
	}
}
