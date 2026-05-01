package pystate

import "testing"

func TestRuntimeLifecycle(t *testing.T) {
	runtime := NewRuntimeState()
	if runtime.Initialized() {
		t.Fatal("runtime should start uninitialized")
	}
	runtime.Init(11)
	if !runtime.Initialized() {
		t.Fatal("runtime should be initialized")
	}
	runtime.EnableInterpreters()
	interp := runtime.NewInterpreter()
	if interp.ID != 0 {
		t.Fatalf("interp.ID = %d, want 0", interp.ID)
	}
	if len(runtime.Interpreters()) != 1 {
		t.Fatal("expected one interpreter")
	}
}

func TestThreadBindingAndCurrent(t *testing.T) {
	runtime := NewRuntimeState()
	interp := runtime.NewInterpreter()
	ts := NewThreadState(interp)
	BindThread(ts, 101, 202)
	if !ts.Status.Bound || ts.ThreadID != 101 || ts.NativeThreadID != 202 {
		t.Fatalf("thread state = %#v", ts)
	}
	BindGILState(runtime, ts)
	if CurrentThreadState(runtime) != ts || !ts.Status.BoundGILState {
		t.Fatalf("current thread binding failed: %#v", ts)
	}
	UnbindGILState(runtime, ts)
	if CurrentThreadState(runtime) != nil || ts.Status.BoundGILState {
		t.Fatalf("current thread unbinding failed: %#v", ts)
	}
	UnbindThread(ts)
	if !ts.Status.Unbound {
		t.Fatalf("thread should be marked unbound: %#v", ts)
	}
}

func TestRunningMain(t *testing.T) {
	runtime := NewRuntimeState()
	interp := runtime.NewInterpreter()
	if err := SetRunningMain(interp); err != nil {
		t.Fatalf("SetRunningMain returned error: %v", err)
	}
	if !interp.RunningMain {
		t.Fatal("interp should be marked running main")
	}
	ClearRunningMain(interp)
	if interp.RunningMain {
		t.Fatal("interp should be cleared from running main")
	}
}

func TestContextWatchers(t *testing.T) {
	runtime := NewRuntimeState()
	interp := runtime.NewInterpreter()
	id, err := AddContextWatcher(interp, func(any) error { return nil })
	if err != nil {
		t.Fatalf("AddContextWatcher returned error: %v", err)
	}
	if interp.ActiveContextBits == 0 {
		t.Fatal("expected active watcher bit")
	}
	if err := ClearContextWatcher(interp, id); err != nil {
		t.Fatalf("ClearContextWatcher returned error: %v", err)
	}
	if interp.ActiveContextBits != 0 {
		t.Fatal("expected watcher bits cleared")
	}
}
