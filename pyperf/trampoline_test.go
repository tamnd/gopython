package pyperf

import (
	"errors"
	"testing"
)

func TestRoundUpAndPerfMapSymbol(t *testing.T) {
	if got := RoundUp(33, 32); got != 64 {
		t.Fatalf("RoundUp = %d, want 64", got)
	}
	co := CodeObject{QualName: "pkg.fn", Filename: "mod.py"}
	if got := PerfMapSymbol(co); got != "py::pkg.fn:mod.py" {
		t.Fatalf("PerfMapSymbol = %q", got)
	}
}

func TestTrampolineManagerCompileAndCallbacks(t *testing.T) {
	wrote := 0
	freed := 0
	manager := NewTrampolineManager([]byte{1, 2, 3, 4}, 0, Callbacks{
		InitState: func() any { return "state" },
		WriteState: func(state any, codeAddr uint64, codeSize uint32, co CodeObject) {
			wrote++
			if state != "state" {
				t.Fatalf("unexpected callback state: %v", state)
			}
			if codeSize != 4 {
				t.Fatalf("unexpected code size: %d", codeSize)
			}
		},
		FreeState: func(state any) error {
			freed++
			if state != "state" {
				t.Fatalf("unexpected free state: %v", state)
			}
			return nil
		},
	})
	if err := manager.Init(TrampolineMap); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	addr1, size1, err := manager.CompileTrampoline(CodeObject{QualName: "a", Filename: "b.py"})
	if err != nil {
		t.Fatalf("CompileTrampoline returned error: %v", err)
	}
	addr2, size2, err := manager.CompileTrampoline(CodeObject{QualName: "c", Filename: "d.py"})
	if err != nil {
		t.Fatalf("CompileTrampoline returned error: %v", err)
	}
	if size1 != 4 || size2 != 4 || addr2 <= addr1 {
		t.Fatalf("unexpected trampoline allocation: %x/%d %x/%d", addr1, size1, addr2, size2)
	}
	if wrote != 2 {
		t.Fatalf("write callbacks = %d, want 2", wrote)
	}
	if err := manager.Fini(); err != nil {
		t.Fatalf("Fini returned error: %v", err)
	}
	if freed != 1 {
		t.Fatalf("free callbacks = %d, want 1", freed)
	}
}

func TestTrampolineManagerInitError(t *testing.T) {
	manager := NewTrampolineManager([]byte{}, 0, Callbacks{})
	manager.ArenaSize = 1
	if err := manager.Init(TrampolineMap); err == nil {
		t.Fatal("Init should fail with invalid arena sizing")
	}
}

func TestTrampolineManagerFiniPropagatesError(t *testing.T) {
	manager := NewTrampolineManager([]byte{1}, 0, Callbacks{
		InitState: func() any { return "state" },
		FreeState: func(state any) error { return errors.New("boom") },
	})
	if err := manager.Init(TrampolineMap); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if err := manager.Fini(); err == nil {
		t.Fatal("Fini should propagate callback error")
	}
}
