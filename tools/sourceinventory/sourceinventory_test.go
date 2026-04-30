package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseInventory(t *testing.T) {
	got := parseInventory("b.c\n\n# comment\na.c\n")
	want := []string{"a.c", "b.c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseInventory() = %#v, want %#v", got, want)
	}
}

func TestCompare(t *testing.T) {
	missing, extra := compare([]string{"a.c", "b.c"}, []string{"b.c", "c.c"})
	if !reflect.DeepEqual(missing, []string{"a.c"}) {
		t.Fatalf("missing = %#v", missing)
	}
	if !reflect.DeepEqual(extra, []string{"c.c"}) {
		t.Fatalf("extra = %#v", extra)
	}
}

func TestTrackedInventoryHasExpectedCount(t *testing.T) {
	names, err := loadTrackedInventory()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 118 {
		t.Fatalf("tracked file count = %d, want 118", len(names))
	}
	if !sortCheck(names) {
		t.Fatalf("tracked inventory is not sorted")
	}
}

func TestListLocalFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"b.c", "a.c"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	got, err := listLocalFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.c", "b.c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("listLocalFiles() = %#v, want %#v", got, want)
	}
}

func sortCheck(names []string) bool {
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			return false
		}
	}
	return true
}
