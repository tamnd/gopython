package pycore

import "testing"

func TestCTypeClasses(t *testing.T) {
	for _, c := range []byte{'\t', '\n', '\v', '\f', '\r', ' '} {
		if !IsSpace(int(c)) {
			t.Fatalf("%q should be space", c)
		}
	}
	if IsSpace('x') {
		t.Fatal("x should not be space")
	}
	for c := byte('0'); c <= '9'; c++ {
		if !IsDigit(int(c)) || !IsXDigit(int(c)) || !IsAlnum(int(c)) {
			t.Fatalf("%q digit flags mismatch", c)
		}
	}
	for c := byte('A'); c <= 'Z'; c++ {
		if !IsUpper(int(c)) || !IsAlpha(int(c)) || !IsAlnum(int(c)) {
			t.Fatalf("%q upper flags mismatch", c)
		}
	}
	for c := byte('a'); c <= 'z'; c++ {
		if !IsLower(int(c)) || !IsAlpha(int(c)) || !IsAlnum(int(c)) {
			t.Fatalf("%q lower flags mismatch", c)
		}
	}
	if IsAlpha('_') || IsAlnum('_') || IsXDigit('g') || IsXDigit('G') {
		t.Fatal("punctuation or non-hex classification mismatch")
	}
}

func TestCTypeCaseTablesAndCharmask(t *testing.T) {
	if Charmask(-1) != 0xff {
		t.Fatalf("Charmask(-1) = %#x", Charmask(-1))
	}
	if ToLower('A') != 'a' || ToLower('Z') != 'z' || ToLower(0xc0) != 0xc0 {
		t.Fatal("tolower mismatch")
	}
	if ToUpper('a') != 'A' || ToUpper('z') != 'Z' || ToUpper(0xe0) != 0xe0 {
		t.Fatal("toupper mismatch")
	}
}
