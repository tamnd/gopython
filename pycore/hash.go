package pycore

import (
	"encoding/binary"
	"math"
	"math/bits"
)

const (
	PyHashMultiplier uint64 = 1000003
	PyHashBits              = 61
	PyHashModulus    uint64 = (uint64(1) << PyHashBits) - 1
	PyHashInf        int64  = 314159
	PyHashImag       uint64 = PyHashMultiplier
)

type HashSecret struct {
	K0 uint64
	K1 uint64
}

type HashFuncDef struct {
	Name     string
	HashBits int
	SeedBits int
}

var PyHashSecret HashSecret

func HashDouble(v float64, nanHash int64) int64 {
	if math.IsInf(v, 1) {
		return PyHashInf
	}
	if math.IsInf(v, -1) {
		return -PyHashInf
	}
	if math.IsNaN(v) {
		return nanHash
	}

	m, e := math.Frexp(v)
	sign := uint64(1)
	if m < 0 {
		sign = ^uint64(0)
		m = -m
	}

	x := uint64(0)
	for m != 0 {
		x = ((x << 28) & PyHashModulus) | (x >> (PyHashBits - 28))
		m *= 268435456.0
		e -= 28
		y := uint64(m)
		m -= float64(y)
		x += y
		if x >= PyHashModulus {
			x -= PyHashModulus
		}
	}

	if e >= 0 {
		e %= PyHashBits
	} else {
		e = PyHashBits - 1 - ((-1 - e) % PyHashBits)
	}
	x = ((x << e) & PyHashModulus) | (x >> (PyHashBits - e))

	if sign == ^uint64(0) {
		x = -x
	}
	if x == ^uint64(0) {
		x = ^uint64(1)
	}
	return int64(x)
}

func HashPointerRaw(ptr uintptr) int64 {
	x := bits.RotateLeft64(uint64(ptr), -4)
	return int64(x)
}

func HashPointer(ptr uintptr) int64 {
	hash := HashPointerRaw(ptr)
	if hash == -1 {
		return -2
	}
	return hash
}

func HashBuffer(data []byte) int64 {
	if len(data) == 0 {
		return 0
	}
	x := int64(sipHash13(PyHashSecret.K0, PyHashSecret.K1, data))
	if x == -1 {
		return -2
	}
	return x
}

func KeyedHash(key uint64, data []byte) uint64 {
	return sipHash13(key, 0, data)
}

func HashGetFuncDef() HashFuncDef {
	return HashFuncDef{
		Name:     "siphash13",
		HashBits: 64,
		SeedBits: 128,
	}
}

func sipHash13(k0, k1 uint64, src []byte) uint64 {
	b := uint64(len(src)) << 56
	in := src

	v0 := k0 ^ 0x736f6d6570736575
	v1 := k1 ^ 0x646f72616e646f6d
	v2 := k0 ^ 0x6c7967656e657261
	v3 := k1 ^ 0x7465646279746573

	for len(in) >= 8 {
		mi := binary.LittleEndian.Uint64(in[:8])
		in = in[8:]
		v3 ^= mi
		singleRound(&v0, &v1, &v2, &v3)
		v0 ^= mi
	}

	t := uint64(0)
	for i, c := range in {
		t |= uint64(c) << (8 * i)
	}
	b |= t

	v3 ^= b
	singleRound(&v0, &v1, &v2, &v3)
	v0 ^= b
	v2 ^= 0xff
	singleRound(&v0, &v1, &v2, &v3)
	singleRound(&v0, &v1, &v2, &v3)
	singleRound(&v0, &v1, &v2, &v3)

	return (v0 ^ v1) ^ (v2 ^ v3)
}

func singleRound(v0, v1, v2, v3 *uint64) {
	halfRound(v0, v1, v2, v3, 13, 16)
	halfRound(v2, v1, v0, v3, 17, 21)
}

func halfRound(a, b, c, d *uint64, s, t int) {
	*a += *b
	*c += *d
	*b = bits.RotateLeft64(*b, s) ^ *a
	*d = bits.RotateLeft64(*d, t) ^ *c
	*a = bits.RotateLeft64(*a, 32)
}
