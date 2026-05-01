package pyruntime

import (
	"fmt"
	"os"

	"github.com/tamnd/gopython/pyconfig"
)

type FrozenMainHooks struct {
	SetBytesArgv     func(config *pyconfig.Config, argv [][]byte) error
	Initialize       func(config *pyconfig.Config) error
	SetRunningMain   func() error
	ClearRunningMain func()
	InitWindows      func()
	TermWindows      func()
	ImportFrozenMain func() (int, error)
	RunStdin         func() error
	Finalize         func() error
	Version          func() string
	Copyright        func() string
	Stdout           *os.File
	StdinIsTTY       bool
}

func FrozenMain(argv [][]byte, hooks FrozenMainHooks) (int, error) {
	var config pyconfig.Config
	pyconfig.InitPythonConfig(&config)
	config.PathconfigWarnings = 0
	config.ParseArgv = 0

	if hooks.SetBytesArgv != nil {
		if err := hooks.SetBytesArgv(&config, argv); err != nil {
			return 1, err
		}
	}

	inspect := os.Getenv("PYTHONINSPECT") != ""

	if hooks.InitWindows != nil {
		hooks.InitWindows()
	}

	if hooks.Initialize != nil {
		if err := hooks.Initialize(&config); err != nil {
			return 1, err
		}
	}

	if hooks.SetRunningMain != nil {
		if err := hooks.SetRunningMain(); err != nil {
			return 1, err
		}
	}

	if config.Verbose > 0 && hooks.Stdout != nil && hooks.Version != nil && hooks.Copyright != nil {
		_, _ = fmt.Fprintf(hooks.Stdout, "Python %s\n%s\n", hooks.Version(), hooks.Copyright())
	}

	status := 1
	if hooks.ImportFrozenMain != nil {
		n, err := hooks.ImportFrozenMain()
		if err != nil {
			status = 1
		} else if n == 0 {
			return 1, fmt.Errorf("the __main__ module is not frozen")
		} else if n > 0 {
			status = 0
		}
	}

	if inspect && hooks.StdinIsTTY && hooks.RunStdin != nil {
		if err := hooks.RunStdin(); err != nil {
			status = 1
		}
	}

	if hooks.TermWindows != nil {
		hooks.TermWindows()
	}
	if hooks.ClearRunningMain != nil {
		hooks.ClearRunningMain()
	}
	if hooks.Finalize != nil {
		if err := hooks.Finalize(); err != nil {
			return 120, nil
		}
	}
	return status, nil
}
