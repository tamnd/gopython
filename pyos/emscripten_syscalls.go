package pyos

import (
	"errors"
	"time"
)

const (
	FIOCLEX    = 0x5451
	FIONCLEX   = 0x5450
	FIONBIO    = 0x5421
	O_NONBLOCK = 0x800
)

var ErrEmscriptenFallback = errors.New("emscripten fallback required")

type AsyncReadStream interface {
	ReadAsync(buffer []byte) (int, error)
}

type SyncReadStream interface {
	Read(buffer []byte) (int, error)
}

type PollMask uint16

const (
	PollIn   PollMask = 0x001
	PollOut  PollMask = 0x004
	PollErr  PollMask = 0x008
	PollHup  PollMask = 0x010
	PollNVal PollMask = 0x020
)

type AsyncPollStream interface {
	PollAsync(timeout time.Duration) (PollMask, error)
}

type SyncPollStream interface {
	Poll(timeout time.Duration) (PollMask, error)
}

type PollFD struct {
	Stream  any
	Events  PollMask
	Revents PollMask
}

func EmscriptenGetUID(nodeGetUID func() (int, bool)) int {
	if nodeGetUID == nil {
		return 0
	}
	if uid, ok := nodeGetUID(); ok {
		return uid
	}
	return 0
}

func EmscriptenUmask(mask int, nodeUmask func(int) (int, bool)) int {
	if nodeUmask == nil {
		return 0
	}
	if value, ok := nodeUmask(mask); ok {
		return value
	}
	return 0
}

func EmscriptenFDRead(stream any, iovs [][]byte, fallback func() (int, error)) (int, error) {
	if async, ok := stream.(AsyncReadStream); ok {
		total := 0
		for _, buf := range iovs {
			n, err := async.ReadAsync(buf)
			total += n
			if err != nil {
				return total, err
			}
			if n < len(buf) {
				break
			}
		}
		return total, nil
	}
	if fallback != nil {
		return fallback()
	}
	return 0, ErrEmscriptenFallback
}

func EmscriptenPoll(fds []PollFD, timeout time.Duration, fallback func() (int, error)) (int, error) {
	usedAsync := false
	nonzero := 0
	for i := range fds {
		fd := &fds[i]
		fd.Revents = 0
		switch stream := fd.Stream.(type) {
		case AsyncPollStream:
			usedAsync = true
			mask, err := stream.PollAsync(timeout)
			if err != nil {
				return 0, err
			}
			fd.Revents = mask & (fd.Events | PollErr | PollHup)
		case SyncPollStream:
			mask, err := stream.Poll(timeout)
			if err != nil {
				return 0, err
			}
			fd.Revents = mask & (fd.Events | PollErr | PollHup)
		default:
			fd.Revents = PollNVal
		}
		if fd.Revents != 0 {
			nonzero++
		}
	}
	if usedAsync {
		return nonzero, nil
	}
	if fallback != nil {
		return fallback()
	}
	return nonzero, nil
}

func EmscriptenIoctl(
	fd int,
	request int,
	varargs any,
	fallback func(int, int, any) (int, error),
	fcntlGet func(int) (int, error),
	fcntlSet func(int, int) (int, error),
) (int, error) {
	switch request {
	case FIOCLEX, FIONCLEX:
		return 0, nil
	case FIONBIO:
		nonblock, ok := varargs.(*int)
		if !ok || nonblock == nil {
			return 0, errors.New("FIONBIO requires *int")
		}
		if fcntlGet == nil || fcntlSet == nil {
			return 0, ErrEmscriptenFallback
		}
		flags, err := fcntlGet(fd)
		if err != nil {
			return 0, err
		}
		if *nonblock != 0 {
			flags |= O_NONBLOCK
		} else {
			flags &^= O_NONBLOCK
		}
		return fcntlSet(fd, flags)
	default:
		if fallback != nil {
			return fallback(fd, request, varargs)
		}
		return 0, ErrEmscriptenFallback
	}
}
