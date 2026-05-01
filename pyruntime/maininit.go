package pyruntime

import (
	"fmt"

	"github.com/tamnd/gopython/pyconfig"
)

type InitMainHooks struct {
	InitImportConfig   func(*pyconfig.Config) error
	UpdateConfig       func(pathOnly bool) error
	InitImportExternal func() error
	InitSignals        func(install int) error
	InitPerf           func(mode int) error
	InitSysStreams     func() error
	InitBuiltinsOpen   func() error
	AddMainModule      func() error
	ImportSite         func() error
	InsertSysPath0     func(path []rune) error
}

func ReconfigureMain(state *BootstrapState, hooks InitMainHooks) error {
	if state == nil || state.Thread == nil || state.Runtime == nil {
		return fmt.Errorf("runtime core not initialized")
	}
	if hooks.UpdateConfig != nil {
		if err := hooks.UpdateConfig(false); err != nil {
			return fmt.Errorf("fail to reconfigure Python: %w", err)
		}
	}
	return nil
}

func InitInterpMain(state *BootstrapState, hooks InitMainHooks, isMainInterp bool) error {
	if state == nil || state.Thread == nil || state.Runtime == nil {
		return fmt.Errorf("runtime core not initialized")
	}
	config := &state.Config

	if config.InstallImportlib == 0 {
		if isMainInterp {
			SetInitialized(true)
		}
		return nil
	}

	if hooks.InitImportConfig != nil {
		if err := hooks.InitImportConfig(config); err != nil {
			return err
		}
	}
	if hooks.UpdateConfig != nil {
		if err := hooks.UpdateConfig(true); err != nil {
			return fmt.Errorf("failed to update the Python config: %w", err)
		}
	}
	if hooks.InitImportExternal != nil {
		if err := hooks.InitImportExternal(); err != nil {
			return err
		}
	}
	if isMainInterp && hooks.InitSignals != nil {
		if err := hooks.InitSignals(config.InstallSignalHandlers); err != nil {
			return err
		}
	}
	if isMainInterp && config.PerfProfiling > 0 && hooks.InitPerf != nil {
		if err := hooks.InitPerf(config.PerfProfiling); err != nil {
			return err
		}
	}
	if hooks.InitSysStreams != nil {
		if err := hooks.InitSysStreams(); err != nil {
			return err
		}
	}
	if hooks.InitBuiltinsOpen != nil {
		if err := hooks.InitBuiltinsOpen(); err != nil {
			return err
		}
	}
	if hooks.AddMainModule != nil {
		if err := hooks.AddMainModule(); err != nil {
			return err
		}
	}
	if isMainInterp {
		SetInitialized(true)
	}
	if config.SiteImport != 0 && hooks.ImportSite != nil {
		if err := hooks.ImportSite(); err != nil {
			return err
		}
	}
	if !isMainInterp && len(config.SysPath0) > 0 && hooks.InsertSysPath0 != nil {
		if err := hooks.InsertSysPath0(config.SysPath0); err != nil {
			return fmt.Errorf("can't initialize sys.path[0]: %w", err)
		}
	}
	return nil
}
