package pyos

import (
	"math"
	"runtime"
	"syscall"
)

var (
	readFDHook  = syscall.Read
	writeFDHook = syscall.Write
)

func ReadFD(fd int, buf []byte) (int, error) {
	if len(buf) > maxReadWriteCount() {
		buf = buf[:maxReadWriteCount()]
	}
	for {
		n, err := readFDHook(fd, buf)
		if err == syscall.EINTR {
			continue
		}
		if err != nil {
			return -1, err
		}
		return n, nil
	}
}

func WriteFD(fd int, buf []byte) (int, error) {
	return writeFDImpl(fd, buf, true)
}

func WriteFDNoRaise(fd int, buf []byte) (int, error) {
	return writeFDImpl(fd, buf, false)
}

func writeFDImpl(fd int, buf []byte, retryOnEINTR bool) (int, error) {
	if len(buf) > maxReadWriteCount() {
		buf = buf[:maxReadWriteCount()]
	}
	for {
		n, err := writeFDHook(fd, buf)
		if err == syscall.EINTR && retryOnEINTR {
			continue
		}
		if err != nil {
			return -1, err
		}
		return n, nil
	}
}

func maxReadWriteCount() int {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return math.MaxInt32
	}
	return math.MaxInt
}
