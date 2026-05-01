package pycrossinterp

import (
	"errors"
	"testing"

	"github.com/tamnd/gopython/pyconfig"
	"github.com/tamnd/gopython/pyruntime"
	"github.com/tamnd/gopython/pystate"
)

func TestRuntimeStateLifecycle(t *testing.T) {
	var global GlobalState
	var state State
	if err := InitGlobalState(&global); err != nil {
		t.Fatalf("InitGlobalState error: %v", err)
	}
	if err := InitState(&state, true); err != nil {
		t.Fatalf("InitState error: %v", err)
	}
	if !state.Initialized || !state.Exceptions.StaticInitialized || !state.Exceptions.HeapInitialized {
		t.Fatalf("unexpected state flags: initialized=%t static=%t heap=%t", state.Initialized, state.Exceptions.StaticInitialized, state.Exceptions.HeapInitialized)
	}
	FiniState(&state, true)
	if state.Initialized || state.Exceptions.StaticInitialized || state.Exceptions.HeapInitialized {
		t.Fatalf("state after fini: initialized=%t static=%t heap=%t", state.Initialized, state.Exceptions.StaticInitialized, state.Exceptions.HeapInitialized)
	}
	FiniGlobalState(&global)
}

func TestInitAndTypeWrappers(t *testing.T) {
	runtime := pystate.NewRuntimeState()
	runtime.Init(1)
	interp := runtime.NewInterpreter()
	var global GlobalState
	var state State
	if err := Init(interp, &state, &global, true); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if err := InitTypes(&state); err != nil {
		t.Fatalf("InitTypes error: %v", err)
	}
	if !state.Exceptions.StaticInitialized {
		t.Fatalf("unexpected static init flag: %+v", state.Exceptions)
	}
	FiniTypes(&state)
	Fini(interp, &state, &global, true)
	if state.Initialized {
		t.Fatalf("state after fini: initialized=%t", state.Initialized)
	}
}

func TestInterpreterAPIWrappers(t *testing.T) {
	runtime := pystate.NewRuntimeState()
	runtime.Init(1)
	mainInterp := runtime.NewInterpreter()
	mainThread := pystate.NewThreadState(mainInterp)
	pystate.BindThread(mainThread, 1, 1)
	pystate.BindGILState(runtime, mainThread)
	pystate.SetMainThreadState(runtime, mainThread)

	result, err := NewInterpreter(runtime, pyconfig.InterpreterConfig{
		UseMainObmalloc:            true,
		AllowFork:                  true,
		AllowExec:                  true,
		AllowThreads:               true,
		AllowDaemonThreads:         true,
		CheckMultiInterpExtensions: true,
		GIL:                        pyconfig.OwnGIL,
	}, nil, pyruntime.NewInterpreterHooks{})
	if err != nil {
		t.Fatalf("NewInterpreter wrapper error: %v", err)
	}
	if result == nil || result.Interp == nil || result.Thread == nil || result.SaveThread != mainThread {
		t.Fatalf("result = %#v", result)
	}

	var save *pystate.ThreadState = mainThread
	if err := EndInterpreter(result.Interp, result.Thread, &save, pyruntime.FinalizeHooks{}); err != nil {
		t.Fatalf("EndInterpreter wrapper error: %v", err)
	}
	if save != mainThread {
		t.Fatalf("save = %#v", save)
	}
}

func TestInterpreterAPIErrorWrap(t *testing.T) {
	runtime := pystate.NewRuntimeState()
	runtime.Init(1)
	mainInterp := runtime.NewInterpreter()
	mainThread := pystate.NewThreadState(mainInterp)
	pystate.BindThread(mainThread, 1, 1)
	pystate.BindGILState(runtime, mainThread)

	_, err := NewInterpreter(runtime, pyconfig.InterpreterConfig{
		CheckMultiInterpExtensions: false,
		GIL:                        pyconfig.SharedGIL,
	}, nil, pyruntime.NewInterpreterHooks{
		InitObjectState: func(*pystate.InterpreterState) error { return nil },
		InitObmalloc:    func(*pystate.InterpreterState) error { return errors.New("boom") },
	})
	if err == nil {
		t.Fatal("expected wrapped creation error")
	}
}

func TestEndInterpreterNotReadyAndNilThread(t *testing.T) {
	runtime := pystate.NewRuntimeState()
	runtime.Init(1)
	mainInterp := runtime.NewInterpreter()
	mainThread := pystate.NewThreadState(mainInterp)
	pystate.BindThread(mainThread, 1, 1)
	pystate.BindGILState(runtime, mainThread)
	pystate.SetMainThreadState(runtime, mainThread)

	notReady := runtime.NewInterpreter()
	notReady.Ready = false
	if err := EndInterpreter(notReady, nil, nil, pyruntime.FinalizeHooks{}); err != nil {
		t.Fatalf("EndInterpreter not-ready error: %v", err)
	}

	other := runtime.NewInterpreter()
	save := mainThread
	if err := EndInterpreter(other, nil, &save, pyruntime.FinalizeHooks{}); err != nil {
		t.Fatalf("EndInterpreter nil-thread error: %v", err)
	}
	if save != mainThread {
		t.Fatalf("save = %#v", save)
	}
	if pystate.CurrentThreadState(runtime) != mainThread {
		t.Fatalf("current = %#v", pystate.CurrentThreadState(runtime))
	}
}
