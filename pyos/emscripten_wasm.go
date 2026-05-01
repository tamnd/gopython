//go:build js && wasm

package pyos

var (
	emscriptenSignalHandling bool
	emscriptenSignalClock    = 50
)

func CheckEmscriptenSignals() {}

func CheckEmscriptenSignalsPeriodically() {
	if emscriptenSignalClock == 0 {
		emscriptenSignalClock = 50
		CheckEmscriptenSignals()
	} else if emscriptenSignalHandling {
		emscriptenSignalClock--
	}
}

func EmscriptenSignalHandlingEnabled() bool {
	return emscriptenSignalHandling
}
