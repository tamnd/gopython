package pythread

import (
	"errors"
	"sync"
	"unsafe"
)

const PthreadKeysMax = 128

var ErrThreadUnavailable = errors.New("pthread stubs cannot create threads")

type StubTLSKey int

type stubTLSEntry struct {
	inUse bool
	value any
}

type StubTLSState struct {
	mu      sync.Mutex
	entries [PthreadKeysMax]stubTLSEntry
}

func (s *StubTLSState) KeyCreate() (StubTLSKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for idx := 0; idx < PthreadKeysMax; idx++ {
		if !s.entries[idx].inUse {
			s.entries[idx].inUse = true
			s.entries[idx].value = nil
			return StubTLSKey(idx), nil
		}
	}
	return 0, ErrThreadUnavailable
}

func (s *StubTLSState) KeyDelete(key StubTLSKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.validKeyLocked(key) {
		return errors.New("invalid stub tls key")
	}
	s.entries[key].inUse = false
	s.entries[key].value = nil
	return nil
}

func (s *StubTLSState) GetSpecific(key StubTLSKey) any {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.validKeyLocked(key) {
		return nil
	}
	return s.entries[key].value
}

func (s *StubTLSState) SetSpecific(key StubTLSKey, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.validKeyLocked(key) {
		return errors.New("invalid stub tls key")
	}
	s.entries[key].value = value
	return nil
}

func (s *StubTLSState) Self() uintptr {
	return uintptr(unsafe.Pointer(s))
}

func StartStubThread(func(any), any) (ThreadIdent, error) {
	return InvalidThreadID, ErrThreadUnavailable
}

func (s *StubTLSState) validKeyLocked(key StubTLSKey) bool {
	return key >= 0 && int(key) < PthreadKeysMax && s.entries[key].inUse
}
