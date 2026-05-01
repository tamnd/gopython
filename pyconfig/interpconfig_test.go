package pyconfig

import "testing"

func TestInterpreterConfigAsMap(t *testing.T) {
	config := InterpreterConfig{
		UseMainObmalloc:            true,
		AllowFork:                  true,
		AllowExec:                  false,
		AllowThreads:               true,
		AllowDaemonThreads:         false,
		CheckMultiInterpExtensions: true,
		GIL:                        OwnGIL,
	}
	dict, err := config.AsMap()
	if err != nil {
		t.Fatalf("AsMap returned error: %v", err)
	}
	if dict["gil"] != "own" {
		t.Fatalf("gil = %v, want own", dict["gil"])
	}
	if dict["allow_exec"] != false {
		t.Fatalf("allow_exec = %v, want false", dict["allow_exec"])
	}
}

func TestInterpreterConfigAsMapRejectsInvalidGIL(t *testing.T) {
	_, err := InterpreterConfig{GIL: GILMode(99)}.AsMap()
	if err == nil {
		t.Fatal("AsMap should reject invalid GIL mode")
	}
}

func TestInitInterpreterConfigFromMap(t *testing.T) {
	var config InterpreterConfig
	err := InitInterpreterConfigFromMap(&config, map[string]any{
		"use_main_obmalloc":             true,
		"allow_fork":                    false,
		"allow_exec":                    true,
		"allow_threads":                 true,
		"allow_daemon_threads":          false,
		"check_multi_interp_extensions": true,
		"gil":                           "shared",
	})
	if err != nil {
		t.Fatalf("InitInterpreterConfigFromMap returned error: %v", err)
	}
	if !config.UseMainObmalloc || config.AllowFork || !config.AllowExec || !config.AllowThreads {
		t.Fatalf("config = %#v", config)
	}
	if config.GIL != SharedGIL {
		t.Fatalf("GIL = %v, want SharedGIL", config.GIL)
	}
}

func TestInitInterpreterConfigFromMapErrors(t *testing.T) {
	tests := []map[string]any{
		{
			"use_main_obmalloc":             true,
			"allow_fork":                    false,
			"allow_exec":                    true,
			"allow_threads":                 true,
			"allow_daemon_threads":          false,
			"check_multi_interp_extensions": true,
		},
		{
			"use_main_obmalloc":             true,
			"allow_fork":                    false,
			"allow_exec":                    true,
			"allow_threads":                 true,
			"allow_daemon_threads":          false,
			"check_multi_interp_extensions": "bad",
			"gil":                           "shared",
		},
		{
			"use_main_obmalloc":             true,
			"allow_fork":                    false,
			"allow_exec":                    true,
			"allow_threads":                 true,
			"allow_daemon_threads":          false,
			"check_multi_interp_extensions": true,
			"gil":                           "wrong",
		},
		{
			"use_main_obmalloc":             true,
			"allow_fork":                    false,
			"allow_exec":                    true,
			"allow_threads":                 true,
			"allow_daemon_threads":          false,
			"check_multi_interp_extensions": true,
			"gil":                           "shared",
			"extra":                         1,
		},
	}
	for _, dict := range tests {
		var config InterpreterConfig
		if err := InitInterpreterConfigFromMap(&config, dict); err == nil {
			t.Fatalf("InitInterpreterConfigFromMap(%v) should fail", dict)
		}
	}
}

func TestUpdateInterpreterConfigFromMap(t *testing.T) {
	config := InterpreterConfig{
		UseMainObmalloc:            true,
		AllowFork:                  true,
		AllowExec:                  true,
		AllowThreads:               false,
		AllowDaemonThreads:         false,
		CheckMultiInterpExtensions: false,
		GIL:                        SharedGIL,
	}
	err := UpdateInterpreterConfigFromMap(&config, map[string]any{
		"allow_threads": true,
		"gil":           "own",
	})
	if err != nil {
		t.Fatalf("UpdateInterpreterConfigFromMap returned error: %v", err)
	}
	if !config.AllowThreads || config.GIL != OwnGIL || !config.UseMainObmalloc || !config.AllowFork {
		t.Fatalf("config = %#v", config)
	}
}

func TestInterpreterConfigInitFromState(t *testing.T) {
	var config InterpreterConfig
	config.InitFromState(InterpreterStateSnapshot{
		FeatureFlags: RuntimeFlagUseMainObmalloc |
			RuntimeFlagExec |
			RuntimeFlagThreads |
			RuntimeFlagMultiInterpExtensions,
		OwnGIL: true,
	})
	if !config.UseMainObmalloc || config.AllowFork || !config.AllowExec || !config.AllowThreads {
		t.Fatalf("config = %#v", config)
	}
	if !config.CheckMultiInterpExtensions || config.GIL != OwnGIL {
		t.Fatalf("config = %#v", config)
	}
}

func TestInterpreterConfigWrapperHelpers(t *testing.T) {
	config := &InterpreterConfig{GIL: SharedGIL}
	dict, err := InterpreterConfigAsMap(config)
	if err != nil || dict["gil"] != "shared" {
		t.Fatalf("InterpreterConfigAsMap = (%#v, %v)", dict, err)
	}
	if _, err := InterpreterConfigAsMap(nil); err == nil {
		t.Fatal("expected nil config error")
	}

	var fromState InterpreterConfig
	if err := InitInterpreterConfigFromState(&fromState, InterpreterStateSnapshot{OwnGIL: true}); err != nil {
		t.Fatalf("InitInterpreterConfigFromState returned error: %v", err)
	}
	if fromState.GIL != OwnGIL {
		t.Fatalf("fromState = %#v", fromState)
	}
	if err := InitInterpreterConfigFromState(nil, InterpreterStateSnapshot{}); err == nil {
		t.Fatal("expected nil config error")
	}
}
