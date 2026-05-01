package pyruntime

import (
	"errors"
	"testing"

	"github.com/tamnd/gopython/pyconfig"
	"github.com/tamnd/gopython/pystate"
)

func TestNewInterpreterFromConfig(t *testing.T) {
	runtime := pystate.NewRuntimeState()
	runtime.Init(1)
	mainInterp := runtime.NewInterpreter()
	mainTS := pystate.NewThreadState(mainInterp)
	pystate.BindThread(mainTS, 1, 1)
	pystate.BindGILState(runtime, mainTS)
	pystate.SetMainThreadState(runtime, mainTS)

	order := []string{}
	ts, err := NewInterpreterFromConfig(runtime, pyconfig.InterpreterConfig{
		UseMainObmalloc:            true,
		AllowFork:                  true,
		AllowExec:                  true,
		AllowThreads:               true,
		AllowDaemonThreads:         true,
		CheckMultiInterpExtensions: true,
		GIL:                        pyconfig.OwnGIL,
	}, InterpreterWhenceCAPI, NewInterpreterHooks{
		OnGILStateDisable: func(*pystate.RuntimeState) { order = append(order, "disable-gilstate") },
		DetachThread:      func(*pystate.ThreadState) { order = append(order, "detach") },
		InitObjectState:   func(*pystate.InterpreterState) error { order = append(order, "object"); return nil },
		InitObmalloc:      func(*pystate.InterpreterState) error { order = append(order, "obmalloc"); return nil },
		InitGIL:           func(*pystate.ThreadState, pyconfig.GILMode) { order = append(order, "gil") },
		InterpInit:        func(*pystate.ThreadState) error { order = append(order, "interp-init"); return nil },
		InitInterpMain:    func(*pystate.ThreadState) error { order = append(order, "main-init"); return nil },
	})
	if err != nil {
		t.Fatalf("NewInterpreterFromConfig error: %v", err)
	}
	if ts == nil || ts.Interp == nil || ts.Interp.Whence != InterpreterWhenceCAPI {
		t.Fatalf("ts = %#v", ts)
	}
	if len(runtime.Interpreters()) != 2 {
		t.Fatalf("interpreters = %#v", runtime.Interpreters())
	}
	if len(order) != 7 {
		t.Fatalf("order = %#v", order)
	}
}

func TestNewInterpreterFailureRestoresMainThread(t *testing.T) {
	runtime := pystate.NewRuntimeState()
	runtime.Init(1)
	mainInterp := runtime.NewInterpreter()
	mainTS := pystate.NewThreadState(mainInterp)
	pystate.BindThread(mainTS, 1, 1)
	pystate.BindGILState(runtime, mainTS)
	pystate.SetMainThreadState(runtime, mainTS)

	attached := false
	_, err := NewInterpreter(runtime, NewInterpreterHooks{
		DetachThread: func(*pystate.ThreadState) {},
		AttachThread: func(ts *pystate.ThreadState) { attached = ts == mainTS },
		InitObmalloc: func(*pystate.InterpreterState) error { return errors.New("boom") },
		InitObjectState: func(*pystate.InterpreterState) error {
			return nil
		},
	})
	if err == nil {
		t.Fatal("expected failure")
	}
	if !attached {
		t.Fatal("expected saved thread reattach")
	}
	if len(runtime.Interpreters()) != 1 {
		t.Fatalf("interpreters = %#v", runtime.Interpreters())
	}
}

func TestNewInterpreterRequiresInitializedRuntime(t *testing.T) {
	if _, err := NewInterpreterFromConfig(pystate.NewRuntimeState(), pyconfig.InterpreterConfig{}, InterpreterWhenceCAPI, NewInterpreterHooks{}); err == nil {
		t.Fatal("expected initialization error")
	}
}

func TestIsInterpreterFinalizing(t *testing.T) {
	runtime := pystate.NewRuntimeState()
	runtime.Init(1)
	interp := runtime.NewInterpreter()
	ts := pystate.NewThreadState(interp)
	pystate.BindThread(ts, 1, 1)
	pystate.BindGILState(runtime, ts)
	if IsInterpreterFinalizing(runtime, interp) {
		t.Fatal("unexpected finalizing")
	}
	interp.Finalizing = true
	if !IsInterpreterFinalizing(runtime, interp) {
		t.Fatal("expected interp finalizing")
	}
}

func TestFinalizeSubinterpreters(t *testing.T) {
	runtime := pystate.NewRuntimeState()
	runtime.Init(1)
	mainInterp := runtime.NewInterpreter()
	sub1 := runtime.NewInterpreter()
	sub2 := runtime.NewInterpreter()
	mainTS := pystate.NewThreadState(mainInterp)
	pystate.BindThread(mainTS, 1, 1)
	pystate.BindGILState(runtime, mainTS)
	pystate.SetMainThreadState(runtime, mainTS)

	ended := []int64{}
	err := FinalizeSubinterpreters(runtime, FinalizeHooks{
		FinalizeInterpDelete: func(interp *pystate.InterpreterState) {
			ended = append(ended, interp.ID)
		},
	})
	if err != nil {
		t.Fatalf("FinalizeSubinterpreters error: %v", err)
	}
	if len(ended) != 2 || ended[0] != sub1.ID || ended[1] != sub2.ID {
		t.Fatalf("ended = %#v", ended)
	}
	if len(runtime.Interpreters()) != 1 {
		t.Fatalf("interpreters = %#v", runtime.Interpreters())
	}
	if pystate.CurrentThreadState(runtime) != mainTS {
		t.Fatalf("current = %#v", pystate.CurrentThreadState(runtime))
	}
}
