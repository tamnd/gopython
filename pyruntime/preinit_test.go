package pyruntime

import (
	"errors"
	"testing"

	"github.com/tamnd/gopython/pyconfig"
	"github.com/tamnd/gopython/pystate"
)

func TestPreInitializeFromConfig(t *testing.T) {
	ok, err := PreInitializeFromConfig(pyconfig.PreConfig{}, false)
	if err != nil || !ok {
		t.Fatalf("PreInitializeFromConfig = (%t, %v)", ok, err)
	}
	ok, err = PreInitializeFromConfig(pyconfig.PreConfig{}, true)
	if err != nil || !ok {
		t.Fatalf("PreInitializeFromConfig existing = (%t, %v)", ok, err)
	}
}

func TestPreInitializeFromPyArgv(t *testing.T) {
	RuntimeFinalize()
	state := &PreInitState{}
	var preconfig pyconfig.PreConfig
	pyconfig.InitPythonPreConfig(&preconfig)
	args := &pyconfig.Argv{
		UseBytesArgv: true,
		BytesArgv:    [][]byte{[]byte("python"), []byte("-c")},
	}
	if err := PreInitializeFromPyArgv(state, preconfig, args); err != nil {
		t.Fatalf("PreInitializeFromPyArgv error: %v", err)
	}
	if !state.Preinitialized || state.Preinitializing {
		t.Fatalf("state = %#v", state)
	}
}

func TestPreInitializeFromConfigObject(t *testing.T) {
	RuntimeFinalize()
	state := &PreInitState{}
	var config pyconfig.Config
	pyconfig.InitPythonConfig(&config)
	config.Argv.AppendWide([]rune("python"))
	if err := PreInitializeFromConfigObject(state, config, nil); err != nil {
		t.Fatalf("PreInitializeFromConfigObject error: %v", err)
	}
	if !state.Preinitialized || state.Preconfig.ParseArgv != 1 {
		t.Fatalf("state = %#v", state)
	}
}

func TestInitCoreFromConfig(t *testing.T) {
	RuntimeFinalize()
	runtime := pystate.NewRuntimeState()
	runtime.Init(1)
	state, err := InitCoreFromConfig(runtime, pyconfig.Config{}, BootstrapHooks{}, false)
	if err != nil || state == nil || state.Runtime == nil {
		t.Fatalf("InitCoreFromConfig = (%#v, %v)", state, err)
	}
	state2, err := InitCoreFromConfig(runtime, pyconfig.Config{}, BootstrapHooks{}, true)
	if err != nil || state2 == nil {
		t.Fatalf("InitCoreFromConfig existing = (%#v, %v)", state2, err)
	}
}

func TestInitMain(t *testing.T) {
	RuntimeFinalize()
	state := &BootstrapState{Runtime: pystate.NewRuntimeState()}
	state.Runtime.Init(1)
	if err := InitMain(state, false, nil, nil, false); err != nil || !IsInitialized() {
		t.Fatalf("InitMain no importlib = %v", err)
	}
	RuntimeFinalize()
	state.Runtime.Init(1)
	err := InitMain(state, true, func() error { return nil }, nil, false)
	if err != nil || !IsInitialized() {
		t.Fatalf("InitMain importlib = %v", err)
	}
	err = InitMain(state, true, nil, func() error { return errors.New("reconfig") }, true)
	if err == nil {
		t.Fatal("expected reconfigure failure")
	}
	if err := InitMain(nil, true, nil, nil, false); err == nil {
		t.Fatal("expected missing runtime error")
	}
}

func TestInitializeFromConfig(t *testing.T) {
	RuntimeFinalize()
	var config pyconfig.Config
	pyconfig.InitCompatConfig(&config)
	called := false
	state, err := InitializeFromConfig(config, nil, BootstrapHooks{}, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("InitializeFromConfig error: %v", err)
	}
	if state == nil || !called || !IsInitialized() {
		t.Fatalf("state = %#v called=%t initialized=%t", state, called, IsInitialized())
	}
}

func TestInitializeEx(t *testing.T) {
	RuntimeFinalize()
	state, err := InitializeEx(1, nil, BootstrapHooks{}, func() error { return nil })
	if err != nil {
		t.Fatalf("InitializeEx error: %v", err)
	}
	if state == nil || state.Config.InstallSignalHandlers != 1 {
		t.Fatalf("state = %#v", state)
	}
}
