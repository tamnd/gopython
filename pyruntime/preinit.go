package pyruntime

import (
	"fmt"

	"github.com/tamnd/gopython/pyconfig"
	"github.com/tamnd/gopython/pystate"
)

type PreInitState struct {
	Runtime         *pystate.RuntimeState
	Preinitializing bool
	Preinitialized  bool
	Preconfig       pyconfig.PreConfig
}

func PreInitializeFromBytesArgs(state *PreInitState, src pyconfig.PreConfig, argv [][]byte) error {
	args := &pyconfig.Argv{
		UseBytesArgv: true,
		BytesArgv:    argv,
	}
	return PreInitializeFromPyArgv(state, src, args)
}

func PreInitializeFromArgs(state *PreInitState, src pyconfig.PreConfig, argv [][]rune) error {
	args := &pyconfig.Argv{
		UseBytesArgv: false,
		WideArgv:     argv,
	}
	return PreInitializeFromPyArgv(state, src, args)
}

func PreInitialize(state *PreInitState, src pyconfig.PreConfig) error {
	return PreInitializeFromPyArgv(state, src, nil)
}

func PreInitializeFromConfig(config pyconfig.PreConfig, alreadyPreinitialized bool) (bool, error) {
	if alreadyPreinitialized {
		return true, nil
	}
	return true, nil
}

func PreInitializeFromPyArgv(state *PreInitState, src pyconfig.PreConfig, args *pyconfig.Argv) error {
	if state == nil {
		return fmt.Errorf("missing preinit state")
	}
	if state.Runtime == nil {
		state.Runtime = RuntimeInitialize(0)
	}
	if state.Preinitialized {
		return nil
	}

	state.Preinitializing = true
	config := src
	if args != nil && src.ParseArgv != 0 {
		var argv pyconfig.WideStringList
		if err := pyconfig.ArgvAsWideStringList(args, &argv); err != nil {
			return err
		}
	}
	state.Preconfig = config
	state.Preinitializing = false
	state.Preinitialized = true
	return nil
}

func PreInitializeFromConfigObject(state *PreInitState, config pyconfig.Config, args *pyconfig.Argv) error {
	if state == nil {
		return fmt.Errorf("missing preinit state")
	}
	if state.Preinitialized {
		return nil
	}

	var preconfig pyconfig.PreConfig
	switch config.ConfigInit {
	case pyconfig.ConfigInitIsolated:
		pyconfig.InitIsolatedPreConfig(&preconfig)
	default:
		pyconfig.InitPythonPreConfig(&preconfig)
	}
	preconfig.ParseArgv = config.ParseArgv

	if config.ParseArgv == 0 {
		return PreInitializeFromPyArgv(state, preconfig, nil)
	}
	if args == nil {
		args = &pyconfig.Argv{
			UseBytesArgv: false,
			WideArgv:     config.Argv.Items,
		}
	}
	return PreInitializeFromPyArgv(state, preconfig, args)
}

func IsRuntimeInitialized() bool {
	lifecycleState.mu.Lock()
	defer lifecycleState.mu.Unlock()
	return lifecycleState.runtimeInitialized
}

func InitCoreFromConfig(runtime *pystate.RuntimeState, config pyconfig.Config, hooks BootstrapHooks, alreadyCore bool) (*BootstrapState, error) {
	if runtime == nil {
		runtime = pystate.NewRuntimeState()
		runtime.Init(1)
	}
	if !alreadyCore {
		return PyInitConfig(runtime, config, hooks)
	}
	return &BootstrapState{Runtime: runtime, Config: config, Thread: pystate.CurrentThreadState(runtime)}, nil
}

func InitMain(state *BootstrapState, installImportlib bool, initInterpMain func() error, reconfigure func() error, alreadyInitialized bool) error {
	if state == nil || state.Runtime == nil {
		return fmt.Errorf("runtime core not initialized")
	}
	if alreadyInitialized {
		if reconfigure != nil {
			return reconfigure()
		}
		return nil
	}
	if !installImportlib {
		SetInitialized(true)
		return nil
	}
	if initInterpMain != nil {
		if err := initInterpMain(); err != nil {
			return err
		}
	}
	SetInitialized(true)
	return nil
}

func InitializeFromConfig(config pyconfig.Config, runtime *pystate.RuntimeState, hooks BootstrapHooks, initInterpMain func() error) (*BootstrapState, error) {
	state, err := InitCoreFromConfig(runtime, config, hooks, IsCoreInitialized())
	if err != nil {
		return nil, err
	}
	if config.InitMain != 0 {
		if err := InitMain(state, config.InstallImportlib != 0, initInterpMain, func() error {
			return Reconfigure(state, config)
		}, IsInitialized()); err != nil {
			return nil, err
		}
	}
	return state, nil
}

func InitializeEx(installSignals int, runtime *pystate.RuntimeState, hooks BootstrapHooks, initInterpMain func() error) (*BootstrapState, error) {
	if IsInitialized() {
		return nil, nil
	}
	var config pyconfig.Config
	pyconfig.InitCompatConfig(&config)
	config.InstallSignalHandlers = installSignals
	return InitializeFromConfig(config, runtime, hooks, initInterpMain)
}
