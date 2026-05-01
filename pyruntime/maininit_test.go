package pyruntime

import (
	"errors"
	"testing"

	"github.com/tamnd/gopython/pyconfig"
	"github.com/tamnd/gopython/pystate"
)

func TestReconfigureMain(t *testing.T) {
	state := &BootstrapState{
		Runtime: pystate.NewRuntimeState(),
		Thread:  &pystate.ThreadState{},
	}
	state.Runtime.Init(1)
	pathOnly := []bool{}
	if err := ReconfigureMain(state, InitMainHooks{
		UpdateConfig: func(path bool) error {
			pathOnly = append(pathOnly, path)
			return nil
		},
	}); err != nil {
		t.Fatalf("ReconfigureMain error: %v", err)
	}
	if len(pathOnly) != 1 || pathOnly[0] {
		t.Fatalf("pathOnly = %#v", pathOnly)
	}
	if err := ReconfigureMain(nil, InitMainHooks{}); err == nil {
		t.Fatal("expected missing runtime error")
	}
}

func TestInitInterpMainMainInterpreter(t *testing.T) {
	RuntimeFinalize()
	var config pyconfig.Config
	pyconfig.InitPythonConfig(&config)
	config.PerfProfiling = 1
	state := &BootstrapState{
		Runtime: pystate.NewRuntimeState(),
		Thread:  &pystate.ThreadState{},
		Config:  config,
	}
	state.Runtime.Init(1)
	order := []string{}
	err := InitInterpMain(state, InitMainHooks{
		InitImportConfig: func(*pyconfig.Config) error { order = append(order, "import-config"); return nil },
		UpdateConfig: func(path bool) error {
			if !path {
				t.Fatal("expected path-only update")
			}
			order = append(order, "update")
			return nil
		},
		InitImportExternal: func() error { order = append(order, "external"); return nil },
		InitSignals:        func(int) error { order = append(order, "signals"); return nil },
		InitPerf:           func(int) error { order = append(order, "perf"); return nil },
		InitSysStreams:     func() error { order = append(order, "streams"); return nil },
		InitBuiltinsOpen:   func() error { order = append(order, "open"); return nil },
		AddMainModule:      func() error { order = append(order, "main"); return nil },
		ImportSite:         func() error { order = append(order, "site"); return nil },
	}, true)
	if err != nil {
		t.Fatalf("InitInterpMain error: %v", err)
	}
	if !IsInitialized() {
		t.Fatal("expected initialized runtime")
	}
	if len(order) != 9 {
		t.Fatalf("order = %#v", order)
	}
}

func TestInitInterpMainSubinterpreterPath0(t *testing.T) {
	RuntimeFinalize()
	var config pyconfig.Config
	pyconfig.InitPythonConfig(&config)
	config.SysPath0 = []rune("pkg")
	state := &BootstrapState{
		Runtime: pystate.NewRuntimeState(),
		Thread:  &pystate.ThreadState{},
		Config:  config,
	}
	state.Runtime.Init(1)
	inserted := []rune(nil)
	err := InitInterpMain(state, InitMainHooks{
		InsertSysPath0: func(path []rune) error {
			inserted = append([]rune(nil), path...)
			return nil
		},
	}, false)
	if err != nil {
		t.Fatalf("InitInterpMain error: %v", err)
	}
	if string(inserted) != "pkg" {
		t.Fatalf("inserted = %q", string(inserted))
	}
	if IsInitialized() {
		t.Fatal("subinterpreter should not mark runtime initialized")
	}
}

func TestInitInterpMainNoImportlibAndFailures(t *testing.T) {
	RuntimeFinalize()
	state := &BootstrapState{
		Runtime: pystate.NewRuntimeState(),
		Thread:  &pystate.ThreadState{},
		Config: pyconfig.Config{
			InstallImportlib: 0,
		},
	}
	state.Runtime.Init(1)
	if err := InitInterpMain(state, InitMainHooks{}, true); err != nil {
		t.Fatalf("InitInterpMain no importlib error: %v", err)
	}
	if !IsInitialized() {
		t.Fatal("expected initialized runtime")
	}

	RuntimeFinalize()
	pyconfig.InitPythonConfig(&state.Config)
	if err := InitInterpMain(state, InitMainHooks{
		AddMainModule: func() error { return errors.New("boom") },
	}, true); err == nil {
		t.Fatal("expected hook failure")
	}
}
