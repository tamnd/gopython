//go:build !windows

package pythread

const threadMinStackSize = 0x8000

func setStackSizePlatform(size uint64) int {
	if size == 0 {
		threadStackSize.Store(0)
		return 0
	}
	if size >= threadMinStackSize {
		threadStackSize.Store(size)
		return 0
	}
	return -1
}
