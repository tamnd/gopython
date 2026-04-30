package pythread

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/tamnd/gopython/pytime"
)

const (
	lockUnlocked    = uint32(0)
	lockLocked      = uint32(1)
	lockHasParked   = uint32(2)
	onceInitialized = uint32(4)
	timeToBeFairNS  = pytime.Time(1_000_000)
)

type LockFlags uint8

const (
	LockDetach LockFlags = 1 << iota
	LockHandleSignals
)

type mutexEntry struct {
	timeToBeFair pytime.Time
	handedOff    bool
}

type PyMutex struct {
	bits atomic.Uint32
}

type RawMutex struct {
	mu      sync.Mutex
	locked  bool
	waiters []*Semaphore
}

type Event struct {
	v atomic.Uint32
}

type OnceFlag struct {
	v atomic.Uint32
}

type RecursiveMutex struct {
	mutex PyMutex
	owner atomic.Uint64
	level int32
}

type RWMutex struct {
	bits atomic.Uint64
}

type SeqLock struct {
	sequence atomic.Uint32
}

var currentThreadIdent = func() ThreadIdent { return 1 }

func (m *PyMutex) key() uintptr {
	return uintptr(unsafe.Pointer(m))
}

func (m *PyMutex) IsLocked() bool {
	return m.bits.Load()&lockLocked != 0
}

func (m *PyMutex) Lock() {
	_ = m.LockTimed(-1, LockDetach)
}

func (m *PyMutex) LockTimed(timeout pytime.Time, flags LockFlags) LockStatus {
	v := m.bits.Load()
	if v&lockLocked == 0 && m.bits.CompareAndSwap(v, v|lockLocked) {
		return LockAcquired
	}
	if timeout == 0 {
		return LockFailure
	}

	now, _ := pytime.MonotonicRaw()
	endtime := pytime.Time(0)
	if timeout > 0 {
		endtime = pytime.Add(now, timeout)
	}
	entry := &mutexEntry{
		timeToBeFair: pytime.Add(now, timeToBeFairNS),
	}

	for {
		v = m.bits.Load()
		if v&lockLocked == 0 {
			if m.bits.CompareAndSwap(v, v|lockLocked) {
				return LockAcquired
			}
			continue
		}
		if timeout == 0 {
			return LockFailure
		}
		if v&lockHasParked == 0 {
			if !m.bits.CompareAndSwap(v, v|lockHasParked) {
				continue
			}
			v |= lockHasParked
		}

		ret := defaultParkingLot.Park(m.key(), func() bool {
			return m.bits.Load() == v
		}, timeout, entry, flags&LockDetach != 0)
		if ret == ParkOK && entry.handedOff {
			return LockAcquired
		}
		if ret == ParkTimeout {
			return LockFailure
		}
		if timeout > 0 {
			timeout = pytime.DeadlineGet(endtime)
			if timeout <= 0 {
				timeout = 0
			}
		}
	}
}

func mutexUnpark(arg any, parkArg any, hasMoreWaiters bool) {
	m := arg.(*PyMutex)
	v := uint32(0)
	if parkArg != nil {
		entry := parkArg.(*mutexEntry)
		now, _ := pytime.MonotonicRaw()
		shouldBeFair := now > entry.timeToBeFair
		entry.handedOff = shouldBeFair
		if shouldBeFair {
			v |= lockLocked
		}
		if hasMoreWaiters {
			v |= lockHasParked
		}
	}
	m.bits.Store(v)
}

func (m *PyMutex) TryUnlock() int {
	for {
		v := m.bits.Load()
		if v&lockLocked == 0 {
			return -1
		}
		if v&lockHasParked != 0 {
			defaultParkingLot.Unpark(m.key(), mutexUnpark, m)
			return 0
		}
		if m.bits.CompareAndSwap(v, lockUnlocked) {
			return 0
		}
	}
}

func (m *PyMutex) Unlock() {
	if m.TryUnlock() < 0 {
		panic("unlocking mutex that is not locked")
	}
}

func (m *RawMutex) Lock() {
	m.mu.Lock()
	if !m.locked {
		m.locked = true
		m.mu.Unlock()
		return
	}
	sema := &Semaphore{}
	sema.Init()
	m.waiters = append(m.waiters, sema)
	m.mu.Unlock()
	for sema.Wait(-1, false) != ParkOK {
	}
}

func (m *RawMutex) Unlock() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.locked {
		panic("unlocking mutex that is not locked")
	}
	if len(m.waiters) == 0 {
		m.locked = false
		return
	}
	waiter := m.waiters[0]
	m.waiters = m.waiters[1:]
	waiter.Wakeup()
}

func (e *Event) IsSet() bool {
	return e.v.Load() == lockLocked
}

func (e *Event) Notify() {
	v := e.v.Swap(lockLocked)
	if v == lockHasParked {
		defaultParkingLot.UnparkAll(uintptr(unsafe.Pointer(e)))
	}
}

func (e *Event) Wait() {
	for !e.WaitTimed(-1, true) {
	}
}

func (e *Event) WaitTimed(timeout pytime.Time, detach bool) bool {
	for {
		v := e.v.Load()
		if v == lockLocked {
			return true
		}
		if v == lockUnlocked && !e.v.CompareAndSwap(v, lockHasParked) {
			continue
		}
		expected := lockHasParked
		_ = defaultParkingLot.Park(uintptr(unsafe.Pointer(e)), func() bool {
			return e.v.Load() == expected
		}, timeout, nil, detach)
		return e.v.Load() == lockLocked
	}
}

func unlockOnce(o *OnceFlag, res int) int {
	var newValue uint32
	switch res {
	case -1:
		newValue = lockUnlocked
	case 0:
		newValue = onceInitialized
	default:
		panic("invalid result from OnceFlag.CallOnce")
	}
	oldValue := o.v.Swap(newValue)
	if oldValue&lockHasParked != 0 {
		defaultParkingLot.UnparkAll(uintptr(unsafe.Pointer(o)))
	}
	return res
}

func (o *OnceFlag) CallOnceSlow(fn func(any) int, arg any) int {
	v := o.v.Load()
	for {
		if v == lockUnlocked {
			if !o.v.CompareAndSwap(v, lockLocked) {
				v = o.v.Load()
				continue
			}
			return unlockOnce(o, fn(arg))
		}
		if v == onceInitialized {
			return 0
		}
		if v&lockHasParked == 0 {
			if !o.v.CompareAndSwap(v, v|lockHasParked) {
				v = o.v.Load()
				continue
			}
			v |= lockHasParked
		}
		_ = defaultParkingLot.Park(uintptr(unsafe.Pointer(o)), func() bool {
			return o.v.Load() == v
		}, -1, nil, true)
		v = o.v.Load()
	}
}

func (m *RecursiveMutex) IsLockedByCurrentThread() bool {
	return m.owner.Load() == uint64(currentThreadIdent())
}

func (m *RecursiveMutex) Lock() {
	thread := currentThreadIdent()
	if m.owner.Load() == uint64(thread) {
		m.level++
		return
	}
	m.mutex.Lock()
	m.owner.Store(uint64(thread))
}

func (m *RecursiveMutex) LockTimed(timeout pytime.Time, flags LockFlags) LockStatus {
	thread := currentThreadIdent()
	if m.owner.Load() == uint64(thread) {
		m.level++
		return LockAcquired
	}
	status := m.mutex.LockTimed(timeout, flags)
	if status == LockAcquired {
		m.owner.Store(uint64(thread))
	}
	return status
}

func (m *RecursiveMutex) TryUnlock() int {
	thread := currentThreadIdent()
	if m.owner.Load() != uint64(thread) {
		return -1
	}
	if m.level > 0 {
		m.level--
		return 0
	}
	m.owner.Store(0)
	m.mutex.Unlock()
	return 0
}

func (m *RecursiveMutex) Unlock() {
	if m.TryUnlock() < 0 {
		panic("unlocking a recursive mutex that is not owned by the current thread")
	}
}

const (
	writeLocked   = uint64(1)
	rwReaderShift = 2
)

func rwReaderCount(bits uint64) uint64 {
	return bits >> rwReaderShift
}

func (m *RWMutex) key() uintptr {
	return uintptr(unsafe.Pointer(m))
}

func (m *RWMutex) setParkedAndWait(bits uint64) uint64 {
	if bits&uint64(lockHasParked) == 0 {
		if !m.bits.CompareAndSwap(bits, bits|uint64(lockHasParked)) {
			return m.bits.Load()
		}
		bits |= uint64(lockHasParked)
	}
	_ = defaultParkingLot.Park(m.key(), func() bool {
		return m.bits.Load() == bits
	}, -1, nil, true)
	return m.bits.Load()
}

func (m *RWMutex) RLock() {
	bits := m.bits.Load()
	for {
		if bits&writeLocked != 0 || bits&uint64(lockHasParked) != 0 {
			bits = m.setParkedAndWait(bits)
			continue
		}
		if m.bits.CompareAndSwap(bits, bits+(1<<rwReaderShift)) {
			return
		}
		bits = m.bits.Load()
	}
}

func (m *RWMutex) RUnlock() {
	bits := m.bits.Add(^uint64((1 << rwReaderShift) - 1))
	newBits := bits - (1 << rwReaderShift)
	if rwReaderCount(newBits) == 0 && newBits&uint64(lockHasParked) != 0 {
		defaultParkingLot.UnparkAll(m.key())
	}
}

func (m *RWMutex) Lock() {
	bits := m.bits.Load()
	for {
		if bits&^uint64(lockHasParked) == 0 {
			if m.bits.CompareAndSwap(bits, bits|writeLocked) {
				return
			}
			bits = m.bits.Load()
			continue
		}
		bits = m.setParkedAndWait(bits)
	}
}

func (m *RWMutex) Unlock() {
	oldBits := m.bits.Swap(0)
	if oldBits&writeLocked == 0 {
		panic("lock was not write-locked")
	}
	if oldBits&uint64(lockHasParked) != 0 {
		defaultParkingLot.UnparkAll(m.key())
	}
}

func (s *SeqLock) LockWrite() {
	prev := s.sequence.Load()
	for {
		if prev&1 == 1 {
			time.Sleep(time.Millisecond)
			prev = s.sequence.Load()
			continue
		}
		if s.sequence.CompareAndSwap(prev, prev+1) {
			return
		}
		time.Sleep(time.Millisecond)
		prev = s.sequence.Load()
	}
}

func (s *SeqLock) AbandonWrite() {
	newSeq := s.sequence.Load() - 1
	s.sequence.Store(newSeq)
}

func (s *SeqLock) UnlockWrite() {
	newSeq := s.sequence.Load() + 1
	s.sequence.Store(newSeq)
}

func (s *SeqLock) BeginRead() uint32 {
	sequence := s.sequence.Load()
	for sequence&1 == 1 {
		time.Sleep(time.Millisecond)
		sequence = s.sequence.Load()
	}
	return sequence
}

func (s *SeqLock) EndRead(previous uint32) bool {
	if s.sequence.Load() == previous {
		return true
	}
	time.Sleep(time.Millisecond)
	return false
}

func (s *SeqLock) AfterFork() int {
	if s.sequence.Load()&1 == 1 {
		s.sequence.Store(0)
		return 1
	}
	return 0
}

func (m *PyMutex) AcquireTimed(timeoutMicroseconds int64, detach bool) LockStatus {
	flags := LockFlags(0)
	if detach {
		flags |= LockDetach
	}
	return m.LockTimed(pytime.Time(timeoutMicroseconds)*pytime.Time(pytime.USToNS), flags)
}
