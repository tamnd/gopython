//go:build windows

package pythread

import "golang.org/x/sys/windows"

var currentNativeThreadID = func() uint64 {
	return uint64(windows.GetCurrentThreadId())
}
