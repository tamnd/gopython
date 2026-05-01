package pyruntime

import (
	"fmt"

	"github.com/tamnd/gopython/pystate"
)

type FinalizeHooks struct {
	ResolveFinalThreadState func(runtime *pystate.RuntimeState) *pystate.ThreadState
	WaitForThreadShutdown   func(*pystate.ThreadState)
	FinishPendingCalls      func(*pystate.ThreadState)
	AtExitCall              func(*pystate.InterpreterState)
	FinalizeSubinterpreters func()
	StopTheWorld            func(*pystate.RuntimeState)
	StartTheWorld           func(*pystate.RuntimeState)
	MarkShuttingDown        func(*pystate.ThreadState)
	DeleteThreadStateList   func([]*pystate.ThreadState)
	FlushStdFiles           func() error
	DisableSignals          func()
	GCCollect               func()
	ImportFiniExternal      func(*pystate.InterpreterState)
	FinalizeModules         func(*pystate.ThreadState)
	EvalFini                func()
	TraceMallocFini         func()
	ImportFiniCore          func(*pystate.InterpreterState)
	ImportFini              func()
	FaulthandlerFini        func()
	HashFini                func()
	FinalizeInterpClear     func(*pystate.ThreadState)
	FinalizeInterpDelete    func(*pystate.InterpreterState)
	CallExitFuncs           func(*pystate.RuntimeState)
	RuntimeFinalize         func()
}

func ResolveFinalThreadState(runtime *pystate.RuntimeState) *pystate.ThreadState {
	if runtime == nil {
		return nil
	}
	if ts := pystate.MainThreadState(runtime); ts != nil {
		return ts
	}
	return pystate.CurrentThreadState(runtime)
}

func FinalizeRuntime(runtime *pystate.RuntimeState, hooks FinalizeHooks) int {
	if runtime == nil || !runtime.Initialized() {
		return 0
	}

	resolve := hooks.ResolveFinalThreadState
	if resolve == nil {
		resolve = ResolveFinalThreadState
	}
	tstate := resolve(runtime)
	if tstate == nil || tstate.Interp == nil {
		return 0
	}

	tstate.Status.Finalizing = true
	interp := tstate.Interp

	if hooks.WaitForThreadShutdown != nil {
		hooks.WaitForThreadShutdown(tstate)
	}
	if hooks.FinishPendingCalls != nil {
		hooks.FinishPendingCalls(tstate)
	}
	if hooks.AtExitCall != nil {
		hooks.AtExitCall(interp)
	}
	if hooks.FinalizeSubinterpreters != nil {
		hooks.FinalizeSubinterpreters()
	}

	if hooks.StopTheWorld != nil {
		hooks.StopTheWorld(runtime)
	}
	others := pystate.RemoveExcept(runtime, tstate)
	for _, other := range others {
		other.Status.Finalizing = true
		if hooks.MarkShuttingDown != nil {
			hooks.MarkShuttingDown(other)
		}
	}
	if hooks.StartTheWorld != nil {
		hooks.StartTheWorld(runtime)
	}
	if hooks.DeleteThreadStateList != nil {
		hooks.DeleteThreadStateList(others)
	}

	status := 0
	if hooks.FlushStdFiles != nil {
		if err := hooks.FlushStdFiles(); err != nil {
			status = -1
		}
	}
	if hooks.DisableSignals != nil {
		hooks.DisableSignals()
	}
	if hooks.GCCollect != nil {
		hooks.GCCollect()
	}
	if hooks.ImportFiniExternal != nil {
		hooks.ImportFiniExternal(interp)
	}
	if hooks.FinalizeModules != nil {
		hooks.FinalizeModules(tstate)
	}
	if hooks.EvalFini != nil {
		hooks.EvalFini()
	}
	if hooks.FlushStdFiles != nil {
		if err := hooks.FlushStdFiles(); err != nil {
			status = -1
		}
	}
	if hooks.TraceMallocFini != nil {
		hooks.TraceMallocFini()
	}
	if hooks.ImportFiniCore != nil {
		hooks.ImportFiniCore(interp)
	}
	if hooks.ImportFini != nil {
		hooks.ImportFini()
	}
	if hooks.FaulthandlerFini != nil {
		hooks.FaulthandlerFini()
	}
	if hooks.HashFini != nil {
		hooks.HashFini()
	}
	if hooks.FinalizeInterpClear != nil {
		hooks.FinalizeInterpClear(tstate)
	}
	tstate.Status.Cleared = true
	if hooks.FinalizeInterpDelete != nil {
		hooks.FinalizeInterpDelete(interp)
	}
	pystate.DeleteInterpreter(runtime, interp)
	tstate.Status.Finalized = true
	if hooks.CallExitFuncs != nil {
		hooks.CallExitFuncs(runtime)
	}
	if hooks.RuntimeFinalize != nil {
		hooks.RuntimeFinalize()
	} else {
		RuntimeFinalize()
	}
	return status
}

func FinalizeEx(hooks FinalizeHooks) int {
	lifecycleState.mu.Lock()
	runtime := lifecycleState.runtime
	lifecycleState.mu.Unlock()
	return FinalizeRuntime(runtime, hooks)
}

func Finalize(hooks FinalizeHooks) {
	_ = FinalizeEx(hooks)
}

func EndInterpreter(tstate *pystate.ThreadState, hooks FinalizeHooks) error {
	if tstate == nil || tstate.Interp == nil {
		return fmt.Errorf("missing thread state")
	}
	runtime := tstate.Interp.Runtime
	if runtime == nil {
		return fmt.Errorf("missing runtime")
	}
	tstate.Status.Finalizing = true
	if hooks.WaitForThreadShutdown != nil {
		hooks.WaitForThreadShutdown(tstate)
	}
	if hooks.FinishPendingCalls != nil {
		hooks.FinishPendingCalls(tstate)
	}
	if hooks.AtExitCall != nil {
		hooks.AtExitCall(tstate.Interp)
	}
	if hooks.StopTheWorld != nil {
		hooks.StopTheWorld(runtime)
	}
	others := pystate.RemoveExcept(runtime, tstate)
	for _, other := range others {
		if hooks.MarkShuttingDown != nil {
			hooks.MarkShuttingDown(other)
		}
	}
	if hooks.StartTheWorld != nil {
		hooks.StartTheWorld(runtime)
	}
	if hooks.DeleteThreadStateList != nil {
		hooks.DeleteThreadStateList(others)
	}
	if hooks.ImportFiniExternal != nil {
		hooks.ImportFiniExternal(tstate.Interp)
	}
	if hooks.FinalizeModules != nil {
		hooks.FinalizeModules(tstate)
	}
	if hooks.ImportFiniCore != nil {
		hooks.ImportFiniCore(tstate.Interp)
	}
	if hooks.FinalizeInterpClear != nil {
		hooks.FinalizeInterpClear(tstate)
	}
	tstate.Status.Cleared = true
	if hooks.FinalizeInterpDelete != nil {
		hooks.FinalizeInterpDelete(tstate.Interp)
	}
	pystate.DeleteInterpreter(runtime, tstate.Interp)
	tstate.Status.Finalized = true
	return nil
}
