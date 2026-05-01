package pyruntime

import (
	"os"
	"testing"

	"github.com/tamnd/gopython/pyconfig"
)

func TestRuntimeLifecycleFlags(t *testing.T) {
	RuntimeFinalize()
	if IsCoreInitialized() || IsInitialized() {
		t.Fatal("runtime flags should start false")
	}
	runtime := RuntimeInitialize(33)
	if runtime == nil {
		t.Fatal("RuntimeInitialize returned nil")
	}
	SetCoreInitialized(true)
	SetInitialized(true)
	if !IsCoreInitialized() || !IsInitialized() {
		t.Fatal("runtime flags should be true")
	}
	RuntimeFinalize()
	if IsCoreInitialized() || IsInitialized() {
		t.Fatal("runtime flags should be reset")
	}
}

func TestLocaleHelpers(t *testing.T) {
	if !LegacyLocaleDetected("", "C", false) {
		t.Fatal("expected legacy locale detection")
	}
	if LegacyLocaleDetected("x", "C", false) || LegacyLocaleDetected("", "C", true) {
		t.Fatal("unexpected legacy locale detection")
	}
	if !IsLocaleCoercionTarget("C.UTF-8") || IsLocaleCoercionTarget("fr_FR.UTF-8") {
		t.Fatal("locale coercion target mismatch")
	}
	locale, ok := CoerceLegacyLocale("", "C", []string{"C.UTF-8"})
	if !ok || locale != "C.UTF-8" {
		t.Fatalf("CoerceLegacyLocale = (%q, %t)", locale, ok)
	}
	if got := SetLocaleFromEnv("", "en_US.UTF-8", ""); got != "en_US.UTF-8" {
		t.Fatalf("SetLocaleFromEnv = %q", got)
	}
	if got := NormalizeAndroidLocale("", "", ""); got != "C.UTF-8" {
		t.Fatalf("NormalizeAndroidLocale = %q", got)
	}
	if EmitLegacyLocaleWarning(false) != "" || EmitLegacyLocaleWarning(true) == "" {
		t.Fatal("legacy locale warning mismatch")
	}
}

func TestInitInterpreterSettings(t *testing.T) {
	flags, err := InitInterpreterSettings(InterpreterConfigView{
		UseMainObmalloc:            true,
		AllowFork:                  true,
		AllowExec:                  true,
		AllowThreads:               true,
		AllowDaemonThreads:         false,
		CheckMultiInterpExtensions: true,
		GIL:                        pyconfig.SharedGIL,
	})
	if err != nil {
		t.Fatalf("InitInterpreterSettings returned error: %v", err)
	}
	if flags&FeatureUseMainObmalloc == 0 || flags&FeatureFork == 0 || flags&FeatureExec == 0 || flags&FeatureThreads == 0 || flags&FeatureMultiInterpExtensions == 0 {
		t.Fatalf("flags = %b", flags)
	}
	if _, err := InitInterpreterSettings(InterpreterConfigView{
		UseMainObmalloc:            false,
		CheckMultiInterpExtensions: false,
		GIL:                        pyconfig.SharedGIL,
	}); err == nil {
		t.Fatal("expected single-phase extension error")
	}
	if _, err := InitInterpreterSettings(InterpreterConfigView{
		UseMainObmalloc:            true,
		CheckMultiInterpExtensions: true,
		GIL:                        pyconfig.GILMode(99),
	}); err == nil {
		t.Fatal("expected invalid GIL error")
	}
}

func TestInspectEnv(t *testing.T) {
	t.Setenv("PYTHONINSPECT", " 1 ")
	if !IsInspectEnvSet() {
		t.Fatal("expected inspect env set")
	}
	_ = os.Unsetenv("PYTHONINSPECT")
	if IsInspectEnvSet() {
		t.Fatal("expected inspect env clear")
	}
}
