//go:build !darwin && !linux && !windows

package pythread

var currentNativeThreadID = func() uint64 {
	return uint64(defaultCurrentThreadIdent())
}
