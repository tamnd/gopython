package pycrossinterp

import (
	"errors"
	"testing"
)

func TestSyncModuleLifecycle(t *testing.T) {
	current := "current"
	main := &SyncModule{Filename: "main.py"}
	hooks := MainModuleHooks{
		GetMainModule: func() (any, error) { return current, nil },
		LoadFromPath: func(filename string, modname string) (any, error) {
			return filename + ":" + modname, nil
		},
		SetMainModule: func(module any) error {
			current = module.(string)
			return nil
		},
	}
	if err := EnsureIsolatedMain(main, hooks); err != nil {
		t.Fatalf("EnsureIsolatedMain returned error: %v", err)
	}
	if err := ApplyIsolatedMain(main, hooks); err != nil {
		t.Fatalf("ApplyIsolatedMain returned error: %v", err)
	}
	if current != "main.py:<fake __main__>" {
		t.Fatalf("current = %q", current)
	}
	if err := RestoreMain(main, hooks); err != nil {
		t.Fatalf("RestoreMain returned error: %v", err)
	}
	if current != "current" {
		t.Fatalf("current = %q", current)
	}
}

func TestSyncModuleFailureCaching(t *testing.T) {
	main := &SyncModule{Filename: "main.py"}
	want := errors.New("boom")
	err := EnsureIsolatedMain(main, MainModuleHooks{
		GetMainModule: func() (any, error) { return "main", nil },
		LoadFromPath:  func(string, string) (any, error) { return nil, want },
	})
	if err != want || main.Cached.Failed != want {
		t.Fatalf("cached failure = %v, err = %v", main.Cached.Failed, err)
	}
	if err := ApplyIsolatedMain(main, MainModuleHooks{}); err != want {
		t.Fatalf("ApplyIsolatedMain error = %v, want %v", err, want)
	}
}

func TestCheckMissingMainAttr(t *testing.T) {
	if !CheckMissingMainAttr(errors.New("module '__main__' has no attribute 'x'")) {
		t.Fatal("expected missing __main__ attr match")
	}
	if CheckMissingMainAttr(errors.New("other")) {
		t.Fatal("unexpected match")
	}
}
