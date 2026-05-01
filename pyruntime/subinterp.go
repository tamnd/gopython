package pyruntime

import (
	"fmt"

	"github.com/tamnd/gopython/pyconfig"
	"github.com/tamnd/gopython/pystate"
)

const (
	InterpreterWhenceCAPI = 1 + iota
	InterpreterWhenceLegacyCAPI
	InterpreterWhenceFinalize
)

type NewInterpreterHooks struct {
	InitObjectState   func(*pystate.InterpreterState) error
	InitObmalloc      func(*pystate.InterpreterState) error
	InitGIL           func(*pystate.ThreadState, pyconfig.GILMode)
	InterpInit        func(*pystate.ThreadState) error
	InitInterpMain    func(*pystate.ThreadState) error
	DetachThread      func(*pystate.ThreadState)
	AttachThread      func(*pystate.ThreadState)
	OnGILStateDisable func(*pystate.RuntimeState)
}

func legacyInterpreterConfig() pyconfig.InterpreterConfig {
	return pyconfig.InterpreterConfig{
		UseMainObmalloc:            true,
		AllowFork:                  true,
		AllowExec:                  true,
		AllowThreads:               true,
		AllowDaemonThreads:         true,
		CheckMultiInterpExtensions: true,
		GIL:                        pyconfig.OwnGIL,
	}
}

func NewInterpreterFromConfig(runtime *pystate.RuntimeState, config pyconfig.InterpreterConfig, whence int, hooks NewInterpreterHooks) (*pystate.ThreadState, error) {
	if runtime == nil || !runtime.Initialized() {
		return nil, fmt.Errorf("Py_Initialize must be called first")
	}
	if hooks.OnGILStateDisable != nil {
		hooks.OnGILStateDisable(runtime)
	}

	save := pystate.CurrentThreadState(runtime)
	if save != nil && hooks.DetachThread != nil {
		hooks.DetachThread(save)
	}

	interp := runtime.NewInterpreter()
	interp.Whence = whence
	interpCfgView := InterpreterConfigView{
		UseMainObmalloc:            config.UseMainObmalloc,
		AllowFork:                  config.AllowFork,
		AllowExec:                  config.AllowExec,
		AllowThreads:               config.AllowThreads,
		AllowDaemonThreads:         config.AllowDaemonThreads,
		CheckMultiInterpExtensions: config.CheckMultiInterpExtensions,
		GIL:                        config.GIL,
	}
	if _, err := InitInterpreterSettings(interpCfgView); err != nil {
		pystate.DeleteInterpreter(runtime, interp)
		if save != nil && hooks.AttachThread != nil {
			hooks.AttachThread(save)
		}
		return nil, err
	}
	if hooks.InitObjectState != nil {
		if err := hooks.InitObjectState(interp); err != nil {
			pystate.DeleteInterpreter(runtime, interp)
			if save != nil && hooks.AttachThread != nil {
				hooks.AttachThread(save)
			}
			return nil, err
		}
	}
	if hooks.InitObmalloc != nil {
		if err := hooks.InitObmalloc(interp); err != nil {
			pystate.DeleteInterpreter(runtime, interp)
			if save != nil && hooks.AttachThread != nil {
				hooks.AttachThread(save)
			}
			return nil, err
		}
	}

	tstate := pystate.NewThreadState(interp)
	tstate.Whence = InterpreterWhenceCAPI
	pystate.BindThread(tstate, uint64(interp.ID+1), uint64(interp.ID+1))
	pystate.BindGILState(runtime, tstate)

	if hooks.InitGIL != nil {
		hooks.InitGIL(tstate, config.GIL)
	}
	if hooks.InterpInit != nil {
		if err := hooks.InterpInit(tstate); err != nil {
			_ = EndInterpreter(tstate, FinalizeHooks{})
			if save != nil && hooks.AttachThread != nil {
				hooks.AttachThread(save)
			}
			return nil, err
		}
	}
	if hooks.InitInterpMain != nil {
		if err := hooks.InitInterpMain(tstate); err != nil {
			_ = EndInterpreter(tstate, FinalizeHooks{})
			if save != nil && hooks.AttachThread != nil {
				hooks.AttachThread(save)
			}
			return nil, err
		}
	}
	return tstate, nil
}

func NewInterpreter(runtime *pystate.RuntimeState, hooks NewInterpreterHooks) (*pystate.ThreadState, error) {
	return NewInterpreterFromConfig(runtime, legacyInterpreterConfig(), InterpreterWhenceLegacyCAPI, hooks)
}

func IsInterpreterFinalizing(runtime *pystate.RuntimeState, interp *pystate.InterpreterState) bool {
	if runtime != nil {
		if ts := pystate.CurrentThreadState(runtime); ts != nil && ts.Status.Finalizing {
			return true
		}
	}
	return interp != nil && interp.Finalizing
}

func FinalizeSubinterpreters(runtime *pystate.RuntimeState, hooks FinalizeHooks) error {
	if runtime == nil {
		return nil
	}
	all := runtime.Interpreters()
	if len(all) <= 1 {
		return nil
	}
	mainTS := pystate.MainThreadState(runtime)
	if mainTS == nil {
		return fmt.Errorf("missing main thread state")
	}
	for _, interp := range all {
		if mainTS.Interp == interp {
			continue
		}
		tstate := pystate.NewThreadState(interp)
		tstate.Whence = InterpreterWhenceFinalize
		pystate.BindThread(tstate, uint64(interp.ID+1), uint64(interp.ID+1))
		pystate.BindGILState(runtime, tstate)
		if err := EndInterpreter(tstate, hooks); err != nil {
			return err
		}
	}
	pystate.BindGILState(runtime, mainTS)
	return nil
}
