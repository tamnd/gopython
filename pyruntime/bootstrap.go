package pyruntime

import (
	"fmt"

	"github.com/tamnd/gopython/pyconfig"
	"github.com/tamnd/gopython/pystate"
)

type BootstrapHooks struct {
	InitGlobalObjects func() error
	InitTypes         func() error
	InitBuiltins      func() error
	InitCrossInterp   func() error
	InitImportCore    func() error
}

type BootstrapState struct {
	Runtime *pystate.RuntimeState
	Thread  *pystate.ThreadState
	Config  pyconfig.Config
}

func PyInitConfig(runtime *pystate.RuntimeState, config pyconfig.Config, hooks BootstrapHooks) (*BootstrapState, error) {
	if runtime == nil {
		runtime = pystate.NewRuntimeState()
		runtime.Init(1)
	}
	interp := runtime.NewInterpreter()
	thread := pystate.NewThreadState(interp)
	pystate.BindThread(thread, 1, 1)
	pystate.BindGILState(runtime, thread)

	if hooks.InitGlobalObjects != nil {
		if err := hooks.InitGlobalObjects(); err != nil {
			return nil, err
		}
	}
	if hooks.InitTypes != nil {
		if err := hooks.InitTypes(); err != nil {
			return nil, err
		}
	}
	if hooks.InitBuiltins != nil {
		if err := hooks.InitBuiltins(); err != nil {
			return nil, err
		}
	}
	if hooks.InitCrossInterp != nil {
		if err := hooks.InitCrossInterp(); err != nil {
			return nil, err
		}
	}
	if hooks.InitImportCore != nil {
		if err := hooks.InitImportCore(); err != nil {
			return nil, err
		}
	}
	SetCoreInitialized(true)
	return &BootstrapState{
		Runtime: runtime,
		Thread:  thread,
		Config:  config,
	}, nil
}

func Reconfigure(state *BootstrapState, config pyconfig.Config) error {
	if state == nil {
		return fmt.Errorf("missing bootstrap state")
	}
	state.Config = config
	return nil
}
