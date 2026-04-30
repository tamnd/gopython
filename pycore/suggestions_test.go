package pycore

import (
	"strings"
	"testing"
)

func TestUTF8EditCost(t *testing.T) {
	tests := []struct {
		a       string
		b       string
		maxCost int
		want    int
	}{
		{"spam", "spam", -1, 0},
		{"spam", "Spam", -1, 1},
		{"spam", "span", -1, 2},
		{"spam", "spams", -1, 2},
		{"abc", "xyz", 2, 3},
		{strings.Repeat("a", 41), "b", -1, 83},
	}
	for _, tt := range tests {
		if got := UTF8EditCost(tt.a, tt.b, tt.maxCost); got != tt.want {
			t.Fatalf(
				"UTF8EditCost(%q, %q, %d) = %d, want %d",
				tt.a,
				tt.b,
				tt.maxCost,
				got,
				tt.want,
			)
		}
	}
}

func TestCalculateSuggestion(t *testing.T) {
	candidates := []string{"append", "extend", "clear"}
	got, ok := CalculateSuggestion(candidates, "apend")
	if !ok {
		t.Fatal("expected suggestion")
	}
	if got != "append" {
		t.Fatalf("suggestion = %q", got)
	}

	if got, ok := CalculateSuggestion(candidates, "append"); ok || got != "" {
		t.Fatalf("exact match should be skipped, got %q", got)
	}
}

func TestCalculateSuggestionLimits(t *testing.T) {
	candidates := make([]string, MaxCandidateItems)
	if got, ok := CalculateSuggestion(candidates, "x"); ok || got != "" {
		t.Fatalf("large candidate list should fail, got %q", got)
	}
	if got, ok := CalculateSuggestion([]string{"abcdefghijklmnop"}, "x"); ok || got != "" {
		t.Fatalf("far match should fail, got %q", got)
	}
}
