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

func TestSyncModuleUsesInterpreterCache(t *testing.T) {
	current := "current"
	cache := map[string]any{}
	main := &SyncModule{Filename: "main.py"}
	loads := 0
	hooks := MainModuleHooks{
		GetMainModule: func() (any, error) { return current, nil },
		LoadFromPath: func(filename string, modname string) (any, error) {
			loads++
			return filename + ":" + modname, nil
		},
		SetMainModule: func(module any) error {
			current = module.(string)
			return nil
		},
		GetCachedModule: func(name string) (any, error) {
			return cache[name], nil
		},
		SetCachedModule: func(name string, module any) error {
			cache[name] = module
			return nil
		},
	}
	if err := EnsureIsolatedMain(main, hooks); err != nil {
		t.Fatalf("first EnsureIsolatedMain error: %v", err)
	}
	if loads != 1 {
		t.Fatalf("loads = %d, want 1", loads)
	}
	main.Clear()
	main.Filename = "main.py"
	if err := EnsureIsolatedMain(main, hooks); err != nil {
		t.Fatalf("second EnsureIsolatedMain error: %v", err)
	}
	if loads != 1 {
		t.Fatalf("loads = %d, want cached reuse", loads)
	}
}

func TestSyncModuleResolvesFilenameFromHook(t *testing.T) {
	current := "current"
	main := &SyncModule{}
	err := EnsureIsolatedMain(main, MainModuleHooks{
		GetMainFilename: func() (string, error) { return "main.py", nil },
		GetMainModule:   func() (any, error) { return current, nil },
		LoadFromPath: func(filename string, modname string) (any, error) {
			return filename + ":" + modname, nil
		},
		SetMainModule: func(module any) error {
			current = module.(string)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("EnsureIsolatedMain returned error: %v", err)
	}
	if main.Filename != "main.py" {
		t.Fatalf("main.Filename = %q", main.Filename)
	}
}

func TestSyncModuleFilenameHookFailureFallsBackToNotImplemented(t *testing.T) {
	main := &SyncModule{}
	err := EnsureIsolatedMain(main, MainModuleHooks{
		GetMainFilename: func() (string, error) { return "", errors.New("boom") },
	})
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("EnsureIsolatedMain error = %v", err)
	}
	if main.Cached.Failed == nil || main.Cached.Failed.Error() != "not implemented" {
		t.Fatalf("cached failure = %v", main.Cached.Failed)
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
