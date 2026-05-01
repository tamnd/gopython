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
	if CurrentThreadState(runtime) != ts || MainThreadState(runtime) != ts || !ts.Status.BoundGILState {
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

func TestRuntimeRemovalHelpers(t *testing.T) {
	runtime := NewRuntimeState()
	interp1 := runtime.NewInterpreter()
	interp2 := runtime.NewInterpreter()
	ts1 := NewThreadState(interp1)
	ts2 := NewThreadState(interp2)
	BindGILState(runtime, ts1)
	SetMainThreadState(runtime, ts1)
	BindGILState(runtime, ts2)

	removed := RemoveExcept(runtime, ts1)
	if len(removed) != 1 || removed[0] != ts2 {
		t.Fatalf("removed = %#v", removed)
	}
	if CurrentThreadState(runtime) != ts1 {
		t.Fatalf("current = %#v", CurrentThreadState(runtime))
	}

	DeleteInterpreter(runtime, interp2)
	if len(runtime.Interpreters()) != 1 {
		t.Fatalf("interpreters = %#v", runtime.Interpreters())
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

func TestInterpreterAccessorsAndStateWrappers(t *testing.T) {
	runtime := NewRuntimeState()
	runtime.Init(7)
	interp := runtime.NewInterpreter()
	ts := NewThreadState(interp)
	BindThread(ts, 1, 2)
	BindGILState(runtime, ts)
	SetMainThreadState(runtime, ts)

	interp.MainModule = map[string]any{"__name__": "__main__"}
	if err := CheckMainModule(GetMainModule(ts)); err != nil {
		t.Fatalf("CheckMainModule returned error: %v", err)
	}
	if err := CheckMainModule(nil); err == nil {
		t.Fatal("expected missing main module error")
	}
	if err := CheckMainModule(map[string]any{"__name__": "other"}); err == nil {
		t.Fatal("expected invalid main module error")
	}

	interpDict := InterpreterGetDict(interp)
	threadDict := ThreadGetDict(ts)
	interpDict["x"] = 1
	threadDict["y"] = 2
	if InterpreterGetDict(interp)["x"] != 1 || ThreadGetDict(ts)["y"] != 2 {
		t.Fatalf("dicts = %#v %#v", interpDict, threadDict)
	}

	if GetInterpreterIDObject(interp) != int64(0) {
		t.Fatalf("id object = %#v", GetInterpreterIDObject(interp))
	}
	RequireInterpreterIDRef(interp, true)
	InterpreterIDIncref(interp)
	InterpreterIDDecref(interp)
	if !InterpreterRequiresIDRef(interp) || interp.IDRefCount != 0 {
		t.Fatalf("interp = %#v", interp)
	}

	if !IsInterpreterReady(interp) {
		t.Fatal("expected ready interpreter")
	}
	SetInterpreterWhence(interp, 5)
	if interp.Whence != 5 {
		t.Fatalf("whence = %d", interp.Whence)
	}

	if err := SetRunningMain(interp); err != nil {
		t.Fatalf("SetRunningMain returned error: %v", err)
	}
	if !IsRunningMain(interp) || !ThreadIsRunningMain(ts) {
		t.Fatalf("running main state = %#v %#v", interp, ts)
	}
	DetachThread(ts)
	ReinitRunningMain(ts)
	if !IsRunningMain(interp) {
		t.Fatal("main thread should remain running while still main thread")
	}

	SetThreadShuttingDown(ts)
	if !ts.Status.Finalizing {
		t.Fatalf("ts = %#v", ts)
	}
	AttachThread(ts)
	if !ts.Status.Active {
		t.Fatalf("ts = %#v", ts)
	}
	SuspendThread(ts)
	if ts.Status.Active {
		t.Fatalf("ts = %#v", ts)
	}
	if !IsMainThread(runtime, ts) {
		t.Fatal("expected main thread")
	}

	interp.Finalizing = true
	if !IsMainInterpreterFinalizing(interp) {
		t.Fatal("expected finalizing interpreter")
	}

	ClearInterpreterFromThread(ts)
	if interp.Ready || interp.MainModule != nil {
		t.Fatalf("interp = %#v", interp)
	}
}

func TestRuntimeFinalizeAndDeleteWrappers(t *testing.T) {
	runtime := NewRuntimeState()
	runtime.Init(3)
	interp := runtime.NewInterpreter()
	DeleteInterpreterState(interp)
	if len(runtime.Interpreters()) != 0 {
		t.Fatalf("interpreters = %#v", runtime.Interpreters())
	}

	interp = runtime.NewInterpreter()
	ClearInterpreter(interp)
	if interp.Ready {
		t.Fatalf("interp = %#v", interp)
	}

	RuntimeStateFini(runtime)
	if runtime.Initialized() || len(runtime.Interpreters()) != 0 {
		t.Fatalf("runtime = %#v", runtime)
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
