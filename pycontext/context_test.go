package pycontext

import "testing"

func TestCopyCurrentAndCopy(t *testing.T) {
	ResetRuntimeForTests()
	varA := NewVar("a", nil)
	if _, err := Set(varA, 1); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	current := CopyCurrent()
	value, ok, err := current.Get(varA, nil)
	if err != nil || !ok || value != 1 {
		t.Fatalf("current.Get = (%v, %t, %v)", value, ok, err)
	}
	cloned := current.Copy()
	if _, err := Set(varA, 2); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	value, ok, err = cloned.Get(varA, nil)
	if err != nil || !ok || value != 1 {
		t.Fatalf("cloned.Get = (%v, %t, %v)", value, ok, err)
	}
}

func TestEnterExitAndRun(t *testing.T) {
	ResetRuntimeForTests()
	ctx := New()
	v := NewVar("v", nil)
	ctx.mu.Lock()
	ctx.vars[v] = 10
	ctx.mu.Unlock()

	if err := Enter(ctx); err != nil {
		t.Fatalf("Enter returned error: %v", err)
	}
	if value, err := Get(v, nil); err != nil || value != 10 {
		t.Fatalf("Get in entered context = (%v, %v)", value, err)
	}
	if err := Exit(ctx); err != nil {
		t.Fatalf("Exit returned error: %v", err)
	}

	value, err := ctx.Run(func() any {
		got, innerErr := Get(v, nil)
		if innerErr != nil {
			t.Fatalf("inner Get returned error: %v", innerErr)
		}
		return got
	})
	if err != nil || value != 10 {
		t.Fatalf("Run = (%v, %v)", value, err)
	}
}

func TestVarSetGetReset(t *testing.T) {
	ResetRuntimeForTests()
	v := NewVar("flag", "default")
	value, err := Get(v, nil)
	if err != nil || value != "default" {
		t.Fatalf("default Get = (%v, %v)", value, err)
	}
	token, err := Set(v, "new")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	value, err = Get(v, nil)
	if err != nil || value != "new" {
		t.Fatalf("set Get = (%v, %v)", value, err)
	}
	if err := Reset(v, token); err != nil {
		t.Fatalf("Reset returned error: %v", err)
	}
	value, err = Get(v, nil)
	if err != nil || value != "default" {
		t.Fatalf("reset Get = (%v, %v)", value, err)
	}
}

func TestContextCollections(t *testing.T) {
	ResetRuntimeForTests()
	ctx := New()
	v1 := NewVar("a", nil)
	v2 := NewVar("b", nil)
	ctx.mu.Lock()
	ctx.vars[v1] = 1
	ctx.vars[v2] = 2
	ctx.mu.Unlock()
	if ctx.Len() != 2 || !ctx.Contains(v1) || ctx.Contains(NewVar("c", nil)) {
		t.Fatalf("context collection helpers failed")
	}
	if len(ctx.Keys()) != 2 || len(ctx.Values()) != 2 || len(ctx.Items()) != 2 {
		t.Fatalf("context iter helpers failed")
	}
}

func TestWatchers(t *testing.T) {
	ResetRuntimeForTests()
	ctx := New()
	calls := 0
	id, err := AddWatcher(func(event Event, current *Context) error {
		if event != ContextSwitched {
			t.Fatalf("event = %v", event)
		}
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("AddWatcher returned error: %v", err)
	}
	if err := Enter(ctx); err != nil {
		t.Fatalf("Enter returned error: %v", err)
	}
	if err := Exit(ctx); err != nil {
		t.Fatalf("Exit returned error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("watcher calls = %d, want 2", calls)
	}
	if err := ClearWatcher(id); err != nil {
		t.Fatalf("ClearWatcher returned error: %v", err)
	}
}

func TestInitModule(t *testing.T) {
	module := InitModule()
	if module.CopyContext == nil || module.ContextType == nil || module.VarType == nil || module.TokenType == nil {
		t.Fatalf("module = %#v", module)
	}
}

func TestNewHamtForTests(t *testing.T) {
	hamt := NewHamtForTests()
	if hamt == nil || len(hamt) != 0 {
		t.Fatalf("hamt = %#v", hamt)
	}
}
