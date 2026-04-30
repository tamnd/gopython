package pycore

import (
	"bytes"
	"errors"
	"testing"
)

func TestStrHex(t *testing.T) {
	data := []byte{0x00, 0x01, 0x0f, 0x10, 0xff}
	if got := StrHex(data); got != "00010f10ff" {
		t.Fatalf("StrHex() = %q", got)
	}
	if got := StrHexBytes(data); !bytes.Equal(got, []byte("00010f10ff")) {
		t.Fatalf("StrHexBytes() = %q", got)
	}
}

func TestStrHexWithSep(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	tests := []struct {
		group int
		want  string
	}{
		{0, "0102030405"},
		{9, "0102030405"},
		{1, "01:02:03:04:05"},
		{2, "01:0203:0405"},
		{-1, "01:02:03:04:05"},
		{-2, "0102:0304:05"},
	}
	for _, tt := range tests {
		if got := StrHexWithSep(data, ':', tt.group); got != tt.want {
			t.Fatalf("group %d = %q, want %q", tt.group, got, tt.want)
		}
	}
}

func TestStrHexWithStringSep(t *testing.T) {
	got, err := StrHexWithStringSep([]byte{1, 2}, "-", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got != "01-02" {
		t.Fatalf("got %q", got)
	}
	if _, err := StrHexWithStringSep([]byte{1}, "", 1); !errors.Is(
		err,
		ErrInvalidHexSeparator,
	) {
		t.Fatalf("empty separator err = %v", err)
	}
}
