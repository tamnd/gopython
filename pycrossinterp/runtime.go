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
	useWhence := pyruntime.InterpreterWhenceXI
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

func NewInterpreterStateOnly(runtime *pystate.RuntimeState, config pyconfig.InterpreterConfig, whence *int, hooks pyruntime.NewInterpreterHooks) (*pystate.InterpreterState, *pystate.ThreadState, error) {
	result, err := NewInterpreter(runtime, config, whence, hooks)
	if err != nil {
		return nil, nil, err
	}
	if result == nil || result.Interp == nil {
		return nil, result.SaveThread, nil
	}
	if result.Thread != nil {
		if result.SaveThread != nil {
			pystate.BindGILState(runtime, result.SaveThread)
		}
		result.Thread.Status.Cleared = true
		result.Thread.Status.Finalized = true
		result.Thread = nil
	}
	return result.Interp, result.SaveThread, nil
}

func EndInterpreter(interp *pystate.InterpreterState, thread *pystate.ThreadState, save **pystate.ThreadState, hooks pyruntime.FinalizeHooks) error {
	if interp == nil {
		return nil
	}
	if !interp.Ready {
		pystate.DeleteInterpreter(interp.Runtime, interp)
		return nil
	}
	runtime := interp.Runtime
	cur := pystate.CurrentThreadState(runtime)
	restore := (*pystate.ThreadState)(nil)
	if save != nil {
		restore = *save
	}
	if thread == nil {
		if cur != nil && cur.Interp == interp {
			thread = cur
		} else {
			thread = pystate.NewThreadState(interp)
			thread.Whence = pyruntime.InterpreterWhenceFinalize
			pystate.BindThread(thread, uint64(interp.ID+1), uint64(interp.ID+1))
			restore = cur
			pystate.BindGILState(runtime, thread)
		}
	} else if cur != thread {
		if cur != nil && cur.Interp != interp {
			restore = cur
		}
		pystate.BindGILState(runtime, thread)
	}
	if err := pyruntime.EndInterpreter(thread, hooks); err != nil {
		return err
	}
	if restore != nil {
		pystate.BindGILState(runtime, restore)
	}
	if save != nil {
		*save = restore
	}
	return nil
}
