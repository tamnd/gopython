//go:build !js || !wasm

package pyos

func CheckEmscriptenSignals()               {}
func CheckEmscriptenSignalsPeriodically()   {}
func EmscriptenSignalHandlingEnabled() bool { return false }
