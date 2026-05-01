package pyos

import "testing"

func TestDesktopEmscriptenNoOp(t *testing.T) {
	CheckEmscriptenSignals()
	CheckEmscriptenSignalsPeriodically()
	if EmscriptenSignalHandlingEnabled() {
		t.Fatalf("EmscriptenSignalHandlingEnabled() = true")
	}
}
