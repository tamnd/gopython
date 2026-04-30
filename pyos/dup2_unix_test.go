//go:build unix

package pyos

import (
	"errors"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestDup2SameFDReturnsSameUnix(t *testing.T) {
	fd, err := Dup2(-1, -1)
	if err != nil {
		t.Fatalf("Dup2(-1, -1) returned error: %v", err)
	}
	if fd != -1 {
		t.Fatalf("Dup2(-1, -1) = %d, want -1", fd)
	}
}

func TestDup2InvalidSourceUnix(t *testing.T) {
	fd, err := Dup2(-1, 9)
	if err == nil {
		t.Fatalf("Dup2(-1, 9) returned success fd %d", fd)
	}
	if !errors.Is(err, unix.EBADF) {
		t.Fatalf("Dup2(-1, 9) error = %v, want EBADF", err)
	}
}

func TestDup2DuplicatesOntoTargetUnix(t *testing.T) {
	readFile, writeFile, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe source failed: %v", err)
	}
	defer readFile.Close()
	defer writeFile.Close()

	target, err := os.CreateTemp(t.TempDir(), "dup2-target-*")
	if err != nil {
		t.Fatalf("os.CreateTemp failed: %v", err)
	}
	defer target.Close()

	targetFD := int(target.Fd())
	fd, err := Dup2(int(writeFile.Fd()), targetFD)
	if err != nil {
		t.Fatalf("Dup2(%d, %d) returned error: %v", writeFile.Fd(), targetFD, err)
	}
	if fd != targetFD {
		t.Fatalf("Dup2(%d, %d) = %d, want %d", writeFile.Fd(), targetFD, fd, targetFD)
	}

	if _, err := unix.Write(targetFD, []byte("ok")); err != nil {
		t.Fatalf("unix.Write(%d) failed: %v", targetFD, err)
	}
	buf := make([]byte, 2)
	if _, err := readFile.Read(buf); err != nil {
		t.Fatalf("readFile.Read failed: %v", err)
	}
	if string(buf) != "ok" {
		t.Fatalf("read = %q, want %q", buf, "ok")
	}
}
