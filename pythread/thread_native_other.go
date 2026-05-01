//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !windows

package pythread

var currentNativeThreadID = func() uint64 {
	return uint64(defaultCurrentThreadIdent())
}
