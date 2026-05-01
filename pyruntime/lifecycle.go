package pyruntime

import (
	"os"
	"strings"
	"sync"

	"github.com/tamnd/gopython/pyconfig"
	"github.com/tamnd/gopython/pystate"
)

var lifecycleState = struct {
	mu                 sync.Mutex
	runtimeInitialized bool
	coreInitialized    bool
	initialized        bool
	runtime            *pystate.RuntimeState
}{
	runtime: pystate.NewRuntimeState(),
}

var localeTargets = []string{"C.UTF-8", "C.utf8", "UTF-8"}

func RuntimeInitialize(mainThread uint64) *pystate.RuntimeState {
	lifecycleState.mu.Lock()
	defer lifecycleState.mu.Unlock()
	if lifecycleState.runtimeInitialized {
		return lifecycleState.runtime
	}
	lifecycleState.runtimeInitialized = true
	lifecycleState.runtime.Init(mainThread)
	return lifecycleState.runtime
}

func RuntimeFinalize() {
	lifecycleState.mu.Lock()
	defer lifecycleState.mu.Unlock()
	lifecycleState.runtimeInitialized = false
	lifecycleState.coreInitialized = false
	lifecycleState.initialized = false
	lifecycleState.runtime = pystate.NewRuntimeState()
}

func IsCoreInitialized() bool {
	lifecycleState.mu.Lock()
	defer lifecycleState.mu.Unlock()
	return lifecycleState.coreInitialized
}

func IsInitialized() bool {
	lifecycleState.mu.Lock()
	defer lifecycleState.mu.Unlock()
	return lifecycleState.initialized
}

func SetCoreInitialized(value bool) {
	lifecycleState.mu.Lock()
	defer lifecycleState.mu.Unlock()
	lifecycleState.coreInitialized = value
}

func SetInitialized(value bool) {
	lifecycleState.mu.Lock()
	defer lifecycleState.mu.Unlock()
	lifecycleState.initialized = value
}

func LegacyLocaleDetected(lcAll string, ctype string, windows bool) bool {
	if windows {
		return false
	}
	if lcAll != "" {
		return false
	}
	return ctype == "C"
}

func IsLocaleCoercionTarget(ctype string) bool {
	for _, target := range localeTargets {
		if ctype == target {
			return true
		}
	}
	return false
}

func CoerceLegacyLocale(lcAll string, current string, candidates []string) (string, bool) {
	if lcAll != "" {
		return current, false
	}
	for _, candidate := range candidates {
		if candidate != "" {
			return candidate, true
		}
	}
	return current, false
}

func SetLocaleFromEnv(lcAll, lcCType, lang string) string {
	for _, value := range []string{lcAll, lcCType, lang} {
		if value != "" {
			return value
		}
	}
	return ""
}

type InterpreterConfigView struct {
	UseMainObmalloc            bool
	AllowFork                  bool
	AllowExec                  bool
	AllowThreads               bool
	AllowDaemonThreads         bool
	CheckMultiInterpExtensions bool
	GIL                        pyconfig.GILMode
}

const (
	FeatureUseMainObmalloc uint32 = 1 << iota
	FeatureFork
	FeatureExec
	FeatureThreads
	FeatureDaemonThreads
	FeatureMultiInterpExtensions
)

func InitInterpreterSettings(config InterpreterConfigView) (uint32, error) {
	var flags uint32
	if config.UseMainObmalloc {
		flags |= FeatureUseMainObmalloc
	} else if !config.CheckMultiInterpExtensions {
		return 0, errSinglePhaseExtension
	}
	if config.AllowFork {
		flags |= FeatureFork
	}
	if config.AllowExec {
		flags |= FeatureExec
	}
	if config.AllowThreads {
		flags |= FeatureThreads
	}
	if config.AllowDaemonThreads {
		flags |= FeatureDaemonThreads
	}
	if config.CheckMultiInterpExtensions {
		flags |= FeatureMultiInterpExtensions
	}
	switch config.GIL {
	case pyconfig.DefaultGIL, pyconfig.SharedGIL, pyconfig.OwnGIL:
	default:
		return 0, errInvalidGIL
	}
	return flags, nil
}

func UpdateEnvForLocale(name, value string) {
	if name != "" {
		_ = os.Setenv(name, value)
	}
}

func NormalizeAndroidLocale(lcAll, lcCType, lang string) string {
	for _, value := range []string{lcAll, lcCType, lang} {
		if value == "C.UTF-8" || value == "en_US.UTF-8" {
			return "C.UTF-8"
		}
		if value != "" {
			return "C"
		}
	}
	return "C.UTF-8"
}

func EmitLegacyLocaleWarning(enabled bool) string {
	if !enabled {
		return ""
	}
	return "Python runtime initialized with LC_CTYPE=C (a locale with default ASCII encoding), which may cause Unicode compatibility problems. Using C.UTF-8, C.utf8, or UTF-8 (if available) as alternative Unicode-compatible locales is recommended.\n"
}

func IsInspectEnvSet() bool {
	return strings.TrimSpace(os.Getenv("PYTHONINSPECT")) != ""
}
