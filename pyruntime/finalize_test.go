package pyruntime

import (
	"errors"
	"testing"

	"github.com/tamnd/gopython/pystate"
)

func TestFinalizeRuntimeOrderAndStatus(t *testing.T) {
	RuntimeFinalize()
	runtime := RuntimeInitialize(1)
	interp := runtime.NewInterpreter()
	ts := pystate.NewThreadState(interp)
	pystate.BindThread(ts, 1, 1)
	pystate.BindGILState(runtime, ts)
	pystate.SetMainThreadState(runtime, ts)
	SetCoreInitialized(true)
	SetInitialized(true)

	order := []string{}
	status := FinalizeRuntime(runtime, FinalizeHooks{
		WaitForThreadShutdown: func(*pystate.ThreadState) { order = append(order, "wait") },
		FinishPendingCalls:    func(*pystate.ThreadState) { order = append(order, "pending") },
		AtExitCall:            func(*pystate.InterpreterState) { order = append(order, "atexit") },
		FinalizeSubinterpreters: func() {
			order = append(order, "subinterpreters")
		},
		StopTheWorld:  func(*pystate.RuntimeState) { order = append(order, "stop") },
		StartTheWorld: func(*pystate.RuntimeState) { order = append(order, "start") },
		DeleteThreadStateList: func([]*pystate.ThreadState) {
			order = append(order, "delete-list")
		},
		FlushStdFiles: func() error {
			order = append(order, "flush")
			if len(order) == 8 {
				return errors.New("flush")
			}
			return nil
		},
		DisableSignals:      func() { order = append(order, "signals") },
		GCCollect:           func() { order = append(order, "gc") },
		ImportFiniExternal:  func(*pystate.InterpreterState) { order = append(order, "import-external") },
		FinalizeModules:     func(*pystate.ThreadState) { order = append(order, "modules") },
		EvalFini:            func() { order = append(order, "eval") },
		TraceMallocFini:     func() { order = append(order, "tracemalloc") },
		ImportFiniCore:      func(*pystate.InterpreterState) { order = append(order, "import-core") },
		ImportFini:          func() { order = append(order, "import") },
		FaulthandlerFini:    func() { order = append(order, "faulthandler") },
		HashFini:            func() { order = append(order, "hash") },
		FinalizeInterpClear: func(*pystate.ThreadState) { order = append(order, "clear") },
		FinalizeInterpDelete: func(*pystate.InterpreterState) {
			order = append(order, "delete")
		},
		CallExitFuncs:   func(*pystate.RuntimeState) { order = append(order, "exitfuncs") },
		RuntimeFinalize: func() { order = append(order, "runtime-finalize"); RuntimeFinalize() },
	})
	if status != -1 {
		t.Fatalf("status = %d, want -1", status)
	}
	if len(order) < 18 {
		t.Fatalf("order = %#v", order)
	}
	if IsInitialized() || IsCoreInitialized() {
		t.Fatal("runtime flags should be reset")
	}
	if !ts.Status.Finalizing || !ts.Status.Cleared || !ts.Status.Finalized {
		t.Fatalf("thread status = %#v", ts.Status)
	}
}

func TestFinalizeExEarlyExit(t *testing.T) {
	RuntimeFinalize()
	if status := FinalizeEx(FinalizeHooks{}); status != 0 {
		t.Fatalf("status = %d, want 0", status)
	}
}

func TestEndInterpreter(t *testing.T) {
	runtime := pystate.NewRuntimeState()
	runtime.Init(1)
	interp1 := runtime.NewInterpreter()
	interp2 := runtime.NewInterpreter()
	ts1 := pystate.NewThreadState(interp1)
	ts2 := pystate.NewThreadState(interp2)
	pystate.BindGILState(runtime, ts1)
	pystate.SetMainThreadState(runtime, ts1)
	pystate.BindGILState(runtime, ts2)

	called := []string{}
	err := EndInterpreter(ts2, FinalizeHooks{
		WaitForThreadShutdown: func(*pystate.ThreadState) { called = append(called, "wait") },
		FinishPendingCalls:    func(*pystate.ThreadState) { called = append(called, "pending") },
		AtExitCall:            func(*pystate.InterpreterState) { called = append(called, "atexit") },
		StopTheWorld:          func(*pystate.RuntimeState) { called = append(called, "stop") },
		StartTheWorld:         func(*pystate.RuntimeState) { called = append(called, "start") },
		DeleteThreadStateList: func([]*pystate.ThreadState) { called = append(called, "delete-list") },
		ImportFiniExternal:    func(*pystate.InterpreterState) { called = append(called, "import-external") },
		FinalizeModules:       func(*pystate.ThreadState) { called = append(called, "modules") },
		ImportFiniCore:        func(*pystate.InterpreterState) { called = append(called, "import-core") },
		FinalizeInterpClear:   func(*pystate.ThreadState) { called = append(called, "clear") },
		FinalizeInterpDelete:  func(*pystate.InterpreterState) { called = append(called, "delete") },
	})
	if err != nil {
		t.Fatalf("EndInterpreter error: %v", err)
	}
	if len(runtime.Interpreters()) != 1 {
		t.Fatalf("interpreters = %#v", runtime.Interpreters())
	}
	if !ts2.Status.Cleared || !ts2.Status.Finalized {
		t.Fatalf("thread status = %#v", ts2.Status)
	}
	if len(called) != 11 {
		t.Fatalf("called = %#v", called)
	}
}
