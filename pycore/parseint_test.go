package pycore

import (
	"errors"
	"testing"
)

func TestPyOSStrtoulBaseDetection(t *testing.T) {
	tests := []struct {
		s     string
		base  int
		value uint64
		end   int
	}{
		{"  123xyz", 0, 123, 5},
		{"0x10z", 0, 16, 4},
		{"0X10z", 16, 16, 4},
		{"0o10z", 0, 8, 4},
		{"0O10z", 8, 8, 4},
		{"0b10z", 0, 2, 4},
		{"0B10z", 2, 2, 4},
		{"0123", 0, 0, 1},
		{"000  x", 0, 0, 5},
		{"0xg", 0, 0, 1},
		{"0og", 8, 0, 1},
		{"0b2", 2, 0, 1},
		{"123", 1, 0, 0},
		{"123", 37, 0, 0},
	}
	for _, tt := range tests {
		value, end, err := PyOSStrtoul(tt.s, tt.base)
		if err != nil {
			t.Fatalf("%q base %d err = %v", tt.s, tt.base, err)
		}
		if value != tt.value || end != tt.end {
			t.Fatalf(
				"%q base %d = (%d, %d), want (%d, %d)",
				tt.s,
				tt.base,
				value,
				end,
				tt.value,
				tt.end,
			)
		}
	}
}

func TestPyOSStrtoulOverflow(t *testing.T) {
	value, end, err := PyOSStrtoul("18446744073709551616x", 10)
	if !errors.Is(err, ErrRange) {
		t.Fatalf("err = %v", err)
	}
	if value != ulongMax {
		t.Fatalf("value = %d", value)
	}
	if end != 20 {
		t.Fatalf("end = %d", end)
	}
}

func TestPyOSStrtol(t *testing.T) {
	tests := []struct {
		s     string
		base  int
		value int64
		end   int
	}{
		{" -10x", 10, -10, 4},
		{"+0x10z", 0, 16, 5},
		{"9223372036854775807x", 10, longMax, 19},
		{"-9223372036854775808x", 10, longMin, 20},
	}
	for _, tt := range tests {
		value, end, err := PyOSStrtol(tt.s, tt.base)
		if err != nil {
			t.Fatalf("%q err = %v", tt.s, err)
		}
		if value != tt.value || end != tt.end {
			t.Fatalf("%q = (%d, %d), want (%d, %d)", tt.s, value, end, tt.value, tt.end)
		}
	}
}

func TestPyOSStrtolOverflow(t *testing.T) {
	value, end, err := PyOSStrtol("9223372036854775808x", 10)
	if !errors.Is(err, ErrRange) {
		t.Fatalf("err = %v", err)
	}
	if value != longMax || end != 19 {
		t.Fatalf("got (%d, %d)", value, end)
	}
}
