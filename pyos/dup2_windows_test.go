//go:build windows

package pyos

import (
	"syscall"
	"testing"
	"unsafe"
)

const oBinary = 0x8000

var (
	procCRTpipe  = msvcrtDLL.NewProc("_pipe")
	procCRTread  = msvcrtDLL.NewProc("_read")
	procCRTwrite = msvcrtDLL.NewProc("_write")
)

func TestDup2SameFDReturnsSameWindows(t *testing.T) {
	fd, err := Dup2(-1, -1)
	if err != nil {
		t.Fatalf("Dup2(-1, -1) returned error: %v", err)
	}
	if fd != -1 {
		t.Fatalf("Dup2(-1, -1) = %d, want -1", fd)
	}
}

func TestDup2InvalidSourceWindows(t *testing.T) {
	rfd, wfd := newCRTPipe(t)
	defer closeCRTFD(t, rfd)
	defer closeCRTFD(t, wfd)

	fd, err := Dup2(-1, wfd)
	if err == nil {
		t.Fatalf("Dup2(-1, %d) returned success fd %d", wfd, fd)
	}
}

func TestDup2DuplicatesOntoTargetWindows(t *testing.T) {
	rfd, wfd := newCRTPipe(t)
	defer closeCRTFD(t, rfd)
	defer closeCRTFD(t, wfd)

	targetRead, targetWrite := newCRTPipe(t)
	defer closeCRTFD(t, targetRead)
	defer closeCRTFD(t, targetWrite)

	fd, err := Dup2(wfd, targetWrite)
	if err != nil {
		t.Fatalf("Dup2(%d, %d) returned error: %v", wfd, targetWrite, err)
	}
	if fd != targetWrite {
		t.Fatalf("Dup2(%d, %d) = %d, want %d", wfd, targetWrite, fd, targetWrite)
	}

	if err := crtWriteAll(targetWrite, []byte("ok")); err != nil {
		t.Fatalf("crtWriteAll failed: %v", err)
	}
	got, err := crtReadExact(rfd, 2)
	if err != nil {
		t.Fatalf("crtReadExact failed: %v", err)
	}
	if string(got) != "ok" {
		t.Fatalf("read = %q, want %q", got, "ok")
	}
}

func newCRTPipe(t *testing.T) (int, int) {
	t.Helper()

	var fds [2]int32
	r1, _, err := procCRTpipe.Call(uintptr(unsafe.Pointer(&fds[0])), 0, oBinary)
	if int32(r1) != 0 {
		t.Fatalf("_pipe failed: %v", err)
	}
	return int(fds[0]), int(fds[1])
}

func closeCRTFD(t *testing.T, fd int) {
	t.Helper()
	r1, _, err := procCRTclose.Call(uintptr(fd))
	if int32(r1) == -1 {
		t.Fatalf("_close(%d) failed: %v", fd, err)
	}
}

func crtWriteAll(fd int, data []byte) error {
	for len(data) > 0 {
		r1, _, err := procCRTwrite.Call(
			uintptr(fd),
			uintptr(unsafe.Pointer(&data[0])),
			uintptr(len(data)),
		)
		n := int(int32(r1))
		if n == -1 {
			if err != syscall.Errno(0) {
				return err
			}
			return syscall.EINVAL
		}
		data = data[n:]
	}
	return nil
}

func crtReadExact(fd int, size int) ([]byte, error) {
	buf := make([]byte, size)
	read := 0
	for read < size {
		r1, _, err := procCRTread.Call(
			uintptr(fd),
			uintptr(unsafe.Pointer(&buf[read])),
			uintptr(size-read),
		)
		n := int(int32(r1))
		if n == -1 {
			if err != syscall.Errno(0) {
				return nil, err
			}
			return nil, syscall.EINVAL
		}
		if n == 0 {
			return nil, syscall.ERROR_BROKEN_PIPE
		}
		read += n
	}
	return buf, nil
}
