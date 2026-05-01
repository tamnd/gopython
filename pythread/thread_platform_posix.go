//go:build !windows

package pythread

import "os"

const threadMinStackSize = 0x8000

func setStackSizePlatform(size uint64) int {
	if size == 0 {
		threadStackSize.Store(0)
		return 0
	}

	minimum := uint64(threadMinStackSize)
	if pageSize := os.Getpagesize(); pageSize > 0 && uint64(pageSize) > minimum {
		minimum = uint64(pageSize)
	}
	if size < minimum {
		return -1
	}

	threadStackSize.Store(size)
	return 0
}
