package pyos

import (
	"errors"
	"testing"
	"time"
)

type fakeAsyncReader struct {
	reads [][]byte
}

func (r *fakeAsyncReader) ReadAsync(buffer []byte) (int, error) {
	next := r.reads[0]
	r.reads = r.reads[1:]
	copy(buffer, next)
	return len(next), nil
}

type fakeAsyncPoller struct {
	mask PollMask
	err  error
}

func (p fakeAsyncPoller) PollAsync(timeout time.Duration) (PollMask, error) {
	return p.mask, p.err
}

type fakeSyncPoller struct {
	mask PollMask
	err  error
}

func (p fakeSyncPoller) Poll(timeout time.Duration) (PollMask, error) {
	return p.mask, p.err
}

func TestEmscriptenGetUIDAndUmaskFallbacks(t *testing.T) {
	if got := EmscriptenGetUID(nil); got != 0 {
		t.Fatalf("uid = %d, want 0", got)
	}
	if got := EmscriptenGetUID(func() (int, bool) { return 42, true }); got != 42 {
		t.Fatalf("uid = %d, want 42", got)
	}
	if got := EmscriptenUmask(18, nil); got != 0 {
		t.Fatalf("umask = %d, want 0", got)
	}
	if got := EmscriptenUmask(18, func(mask int) (int, bool) { return mask + 1, true }); got != 19 {
		t.Fatalf("umask = %d, want 19", got)
	}
}

func TestEmscriptenFDReadUsesAsyncReader(t *testing.T) {
	stream := &fakeAsyncReader{
		reads: [][]byte{
			[]byte("ab"),
			[]byte("c"),
		},
	}
	iovs := [][]byte{
		make([]byte, 2),
		make([]byte, 2),
	}

	n, err := EmscriptenFDRead(stream, iovs, nil)
	if err != nil {
		t.Fatalf("EmscriptenFDRead returned error: %v", err)
	}
	if n != 3 {
		t.Fatalf("bytes read = %d, want 3", n)
	}
	if string(iovs[0]) != "ab" {
		t.Fatalf("first iov = %q, want ab", string(iovs[0]))
	}
	if string(iovs[1][:1]) != "c" {
		t.Fatalf("second iov = %q, want c", string(iovs[1][:1]))
	}
}

func TestEmscriptenFDReadFallsBack(t *testing.T) {
	n, err := EmscriptenFDRead(struct{}{}, nil, func() (int, error) { return 7, nil })
	if err != nil || n != 7 {
		t.Fatalf("fallback = (%d, %v), want (7, nil)", n, err)
	}
	_, err = EmscriptenFDRead(struct{}{}, nil, nil)
	if !errors.Is(err, ErrEmscriptenFallback) {
		t.Fatalf("unexpected fallback error: %v", err)
	}
}

func TestEmscriptenPollUsesAsyncAndSyncPollers(t *testing.T) {
	fds := []PollFD{
		{Stream: fakeAsyncPoller{mask: PollIn | PollErr}, Events: PollIn},
		{Stream: fakeSyncPoller{mask: PollOut}, Events: PollOut},
		{Stream: struct{}{}, Events: PollIn},
	}

	n, err := EmscriptenPoll(fds, time.Second, nil)
	if err != nil {
		t.Fatalf("EmscriptenPoll returned error: %v", err)
	}
	if n != 3 {
		t.Fatalf("ready count = %d, want 3", n)
	}
	if fds[0].Revents != (PollIn | PollErr) {
		t.Fatalf("async revents = %#x", fds[0].Revents)
	}
	if fds[1].Revents != PollOut {
		t.Fatalf("sync revents = %#x", fds[1].Revents)
	}
	if fds[2].Revents != PollNVal {
		t.Fatalf("invalid revents = %#x", fds[2].Revents)
	}
}

func TestEmscriptenIoctlSpecialCases(t *testing.T) {
	if got, err := EmscriptenIoctl(1, FIOCLEX, nil, nil, nil, nil); err != nil || got != 0 {
		t.Fatalf("FIOCLEX = (%d, %v), want (0, nil)", got, err)
	}

	state := 0
	nonblock := 1
	got, err := EmscriptenIoctl(
		3,
		FIONBIO,
		&nonblock,
		nil,
		func(fd int) (int, error) { return state, nil },
		func(fd int, flags int) (int, error) {
			state = flags
			return flags, nil
		},
	)
	if err != nil {
		t.Fatalf("FIONBIO returned error: %v", err)
	}
	if got&O_NONBLOCK == 0 || state&O_NONBLOCK == 0 {
		t.Fatalf("nonblock flag not set: got=%#x state=%#x", got, state)
	}
}

func TestEmscriptenIoctlFallbackAndErrors(t *testing.T) {
	if _, err := EmscriptenIoctl(1, FIONBIO, nil, nil, nil, nil); err == nil {
		t.Fatal("expected FIONBIO argument error")
	}
	got, err := EmscriptenIoctl(1, 99, "arg", func(fd int, req int, arg any) (int, error) {
		return req + fd, nil
	}, nil, nil)
	if err != nil || got != 100 {
		t.Fatalf("fallback = (%d, %v), want (100, nil)", got, err)
	}
	_, err = EmscriptenIoctl(1, 99, nil, nil, nil, nil)
	if !errors.Is(err, ErrEmscriptenFallback) {
		t.Fatalf("unexpected fallback error: %v", err)
	}
}
