package pycore

import "testing"

func TestMyStrICmp(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		{"abc", "ABC", 0},
		{"abc", "abd", -1},
		{"abd", "abc", 1},
		{"abc", "abcd", -int('d')},
		{"abc\x00zzz", "ABC\x00yyy", 0},
	}
	for _, tt := range tests {
		if got := MyStrICmp(tt.a, tt.b); got != tt.want {
			t.Fatalf("MyStrICmp(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMyStrNICmp(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		size int
		want int
	}{
		{"abc", "ABD", 0, 0},
		{"abc", "ABD", 2, 0},
		{"abc", "ABD", 3, -1},
		{"abc\x00x", "ABC\x00y", 8, 0},
	}
	for _, tt := range tests {
		if got := MyStrNICmp(tt.a, tt.b, tt.size); got != tt.want {
			t.Fatalf(
				"MyStrNICmp(%q, %q, %d) = %d, want %d",
				tt.a,
				tt.b,
				tt.size,
				got,
				tt.want,
			)
		}
	}
}
