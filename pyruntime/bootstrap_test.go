package pyruntime

import (
	"errors"
	"testing"

	"github.com/tamnd/gopython/pyconfig"
	"github.com/tamnd/gopython/pystate"
)

func TestPyInitConfig(t *testing.T) {
	RuntimeFinalize()
	runtime := pystate.NewRuntimeState()
	runtime.Init(5)
	order := []string{}
	state, err := PyInitConfig(runtime, pyconfig.Config{}, BootstrapHooks{
		InitGlobalObjects: func() error { order = append(order, "global"); return nil },
		InitTypes:         func() error { order = append(order, "types"); return nil },
		InitBuiltins:      func() error { order = append(order, "builtins"); return nil },
		InitCrossInterp:   func() error { order = append(order, "xi"); return nil },
		InitImportCore:    func() error { order = append(order, "import"); return nil },
	})
	if err != nil {
		t.Fatalf("PyInitConfig returned error: %v", err)
	}
	if state == nil || state.Thread == nil || !IsCoreInitialized() {
		t.Fatalf("state = %#v", state)
	}
	if len(order) != 5 {
		t.Fatalf("order = %#v", order)
	}
}

func TestPyInitConfigFailure(t *testing.T) {
	_, err := PyInitConfig(nil, pyconfig.Config{}, BootstrapHooks{
		InitTypes: func() error { return errors.New("fail") },
	})
	if err == nil {
		t.Fatal("expected init failure")
	}
}

func TestReconfigure(t *testing.T) {
	state := &BootstrapState{}
	cfg := pyconfig.Config{Isolated: 1}
	if err := Reconfigure(state, cfg); err != nil {
		t.Fatalf("Reconfigure returned error: %v", err)
	}
	if state.Config.Isolated != 1 {
		t.Fatalf("state.Config = %#v", state.Config)
	}
	if err := Reconfigure(nil, cfg); err == nil {
		t.Fatal("expected reconfigure error")
	}
}
