//go:build linux

package pythread

import "golang.org/x/sys/unix"

var currentNativeThreadID = func() uint64 {
	return uint64(unix.Gettid())
}
