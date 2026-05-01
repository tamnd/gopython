package pycrossinterp

import (
	"fmt"

	"github.com/tamnd/gopython/pyconfig"
	"github.com/tamnd/gopython/pyruntime"
	"github.com/tamnd/gopython/pystate"
)

type ExceptionTypes struct {
	StaticInitialized bool
	HeapInitialized   bool
}

type GlobalState struct {
	DataLookup LookupState
}

type State struct {
	DataLookup  LookupState
	Exceptions  ExceptionTypes
	Initialized bool
}

func InitGlobalState(state *GlobalState) error {
	if state == nil {
		return fmt.Errorf("missing global cross-interpreter state")
	}
	state.DataLookup.Init()
	return nil
}

func FiniGlobalState(state *GlobalState) {
	if state == nil {
		return
	}
	state.DataLookup.Fini()
}

func InitState(state *State, initStaticTypes bool) error {
	if state == nil {
		return fmt.Errorf("failed to get interpreter's cross-interpreter state")
	}
	state.DataLookup.Init()
	if initStaticTypes {
		state.Exceptions.StaticInitialized = true
	}
	state.Exceptions.HeapInitialized = true
	state.Initialized = true
	return nil
}

func FiniState(state *State, finiStaticTypes bool) {
	if state == nil {
		return
	}
	state.Exceptions.HeapInitialized = false
	if finiStaticTypes {
		state.Exceptions.StaticInitialized = false
	}
	state.DataLookup.Fini()
	state.Initialized = false
}

func Init(interp *pystate.InterpreterState, state *State, global *GlobalState, isMain bool) error {
	if isMain {
		if err := InitGlobalState(global); err != nil {
			return fmt.Errorf("failed to initialize  global cross-interpreter state")
		}
	}
	if interp == nil || state == nil {
		return fmt.Errorf("failed to get interpreter's cross-interpreter state")
	}
	if err := InitState(state, false); err != nil {
		return fmt.Errorf("failed to initialize interpreter's cross-interpreter state")
	}
	return nil
}

func Fini(interp *pystate.InterpreterState, state *State, global *GlobalState, isMain bool) {
	_ = interp
	FiniState(state, false)
	if isMain {
		FiniGlobalState(global)
	}
}

func InitTypes(state *State) error {
	if state == nil {
		return fmt.Errorf("failed to initialize the cross-interpreter exception types")
	}
	state.Exceptions.StaticInitialized = true
	return nil
}

func FiniTypes(state *State) {
	if state == nil {
		return
	}
	state.Exceptions.StaticInitialized = false
}

type InterpreterAPIResult struct {
	Interp     *pystate.InterpreterState
	Thread     *pystate.ThreadState
	SaveThread *pystate.ThreadState
}

func NewInterpreter(runtime *pystate.RuntimeState, config pyconfig.InterpreterConfig, whence *int, hooks pyruntime.NewInterpreterHooks) (*InterpreterAPIResult, error) {
	save := pystate.CurrentThreadState(runtime)
	useWhence := pyruntime.InterpreterWhenceCAPI
	if whence != nil {
		useWhence = *whence
	}
	thread, err := pyruntime.NewInterpreterFromConfig(runtime, config, useWhence, hooks)
	if err != nil {
		return nil, fmt.Errorf("sub-interpreter creation failed: %w", err)
	}
	return &InterpreterAPIResult{
		Interp:     thread.Interp,
		Thread:     thread,
		SaveThread: save,
	}, nil
}

func EndInterpreter(interp *pystate.InterpreterState, thread *pystate.ThreadState, save **pystate.ThreadState, hooks pyruntime.FinalizeHooks) error {
	if interp == nil {
		return nil
	}
	if thread == nil {
		pystate.DeleteInterpreter(interp.Runtime, interp)
		return nil
	}
	if err := pyruntime.EndInterpreter(thread, hooks); err != nil {
		return err
	}
	if save != nil {
		*save = nil
	}
	return nil
}
