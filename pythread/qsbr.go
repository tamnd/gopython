package pythread

import (
	"fmt"
	"sync/atomic"
)

const (
	QSBROffline = uint64(0)
	QSBRInitial = uint64(1)
	QSBRIncr    = uint64(2)
	qsbrMinSize = 8
)

type QSBRThreadState struct {
	seq                uint64
	shared             *QSBRShared
	tstate             *ThreadState
	DeferredCount      int
	DeferredMemory     uintptr
	DeferredPageMemory uintptr
	ShouldProcess      bool
	Allocated          bool
	freelistNext       *QSBRThreadState
}

type qsbrPad struct {
	qsbr QSBRThreadState
}

type QSBRShared struct {
	wrSeq    uint64
	rdSeq    uint64
	array    []qsbrPad
	mutex    Mutex
	freelist *QSBRThreadState
}

func (s *QSBRShared) init() {
	if atomic.LoadUint64(&s.wrSeq) == 0 {
		atomic.StoreUint64(&s.wrSeq, QSBRInitial)
	}
	if atomic.LoadUint64(&s.rdSeq) == 0 {
		atomic.StoreUint64(&s.rdSeq, QSBRInitial)
	}
}

func (s *QSBRShared) Current() uint64 {
	s.init()
	return atomic.LoadUint64(&s.wrSeq)
}

func (s *QSBRShared) QuiescentState(qsbr *QSBRThreadState) {
	seq := s.Current()
	atomic.StoreUint64(&qsbr.seq, seq)
}

func GoalReached(qsbr *QSBRThreadState, goal uint64) bool {
	rdSeq := atomic.LoadUint64(&qsbr.shared.rdSeq)
	return qsbrLEQ(goal, rdSeq)
}

func (s *QSBRShared) Advance() uint64 {
	s.init()
	return atomic.AddUint64(&s.wrSeq, QSBRIncr)
}

func (s *QSBRShared) Next() uint64 {
	return s.Current() + QSBRIncr
}

func (s *QSBRShared) Poll(qsbr *QSBRThreadState, goal uint64) bool {
	if GoalReached(qsbr, goal) {
		return true
	}
	rdSeq := s.pollScan()
	return qsbrLEQ(goal, rdSeq)
}

func (s *QSBRShared) Attach(qsbr *QSBRThreadState) {
	if atomic.LoadUint64(&qsbr.seq) != 0 {
		panic("qsbr attach: already attached")
	}
	seq := s.Current()
	atomic.StoreUint64(&qsbr.seq, seq)
}

func (s *QSBRShared) Detach(qsbr *QSBRThreadState) {
	if atomic.LoadUint64(&qsbr.seq) == 0 {
		panic("qsbr detach: already detached")
	}
	atomic.StoreUint64(&qsbr.seq, QSBROffline)
}

func (s *QSBRShared) Reserve() int {
	s.init()
	s.mutex.Lock()
	defer s.mutex.Unlock()

	qsbr := s.allocate()
	if qsbr == nil {
		if !s.growThreadArray() {
			return -1
		}
		qsbr = s.allocate()
	}
	if qsbr == nil {
		return -1
	}
	return s.indexOf(qsbr)
}

func (s *QSBRShared) Register(tstate *ThreadState, index int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if index < 0 || index >= len(s.array) {
		panic(fmt.Sprintf("qsbr register: invalid index %d", index))
	}
	qsbr := &s.array[index].qsbr
	if !qsbr.Allocated || qsbr.tstate != nil {
		panic("qsbr register: slot is not available")
	}
	qsbr.tstate = tstate
	tstate.qsbr = qsbr
}

func (s *QSBRShared) Unregister(tstate *ThreadState) {
	if tstate == nil || tstate.Interp == nil || tstate.qsbr == nil {
		panic("qsbr unregister: missing thread state")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	qsbr := tstate.qsbr
	if atomic.LoadUint64(&qsbr.seq) != 0 {
		panic("qsbr unregister: thread state must be detached")
	}
	if !qsbr.Allocated || qsbr.tstate != tstate {
		panic("qsbr unregister: slot is not registered")
	}

	tstate.qsbr = nil
	qsbr.tstate = nil
	qsbr.Allocated = false
	qsbr.shared = s
	qsbr.freelistNext = s.freelist
	s.freelist = qsbr
}

func (s *QSBRShared) Finish() {
	s.array = nil
	s.freelist = nil
	atomic.StoreUint64(&s.wrSeq, 0)
	atomic.StoreUint64(&s.rdSeq, 0)
}

func (s *QSBRShared) AfterFork(tstate *ThreadState) {
	if tstate == nil || tstate.qsbr == nil {
		panic("qsbr after fork: missing current thread state")
	}

	thisQSBR := tstate.qsbr
	s.mutex = Mutex{}
	s.freelist = nil
	for i := range s.array {
		qsbr := &s.array[i].qsbr
		if qsbr != thisQSBR && qsbr.Allocated {
			qsbr.tstate = nil
			qsbr.Allocated = false
			qsbr.freelistNext = s.freelist
			s.freelist = qsbr
		}
	}
}

func (s *QSBRShared) allocate() *QSBRThreadState {
	qsbr := s.freelist
	if qsbr == nil {
		return nil
	}
	s.freelist = qsbr.freelistNext
	qsbr.freelistNext = nil
	qsbr.shared = s
	qsbr.Allocated = true
	return qsbr
}

func (s *QSBRShared) initializeNewArray() {
	for i := range s.array {
		qsbr := &s.array[i].qsbr
		if qsbr.tstate != nil {
			qsbr.tstate.qsbr = qsbr
		}
		if !qsbr.Allocated {
			qsbr.freelistNext = s.freelist
			s.freelist = qsbr
		}
	}
}

func (s *QSBRShared) growThreadArray() bool {
	newSize := len(s.array) * 2
	if newSize < qsbrMinSize {
		newSize = qsbrMinSize
	}
	array := make([]qsbrPad, newSize)
	copy(array, s.array)
	s.array = array
	s.freelist = nil
	s.initializeNewArray()
	return true
}

func (s *QSBRShared) pollScan() uint64 {
	minSeq := atomic.LoadUint64(&s.wrSeq)
	for i := range s.array {
		seq := atomic.LoadUint64(&s.array[i].qsbr.seq)
		if seq != QSBROffline && qsbrLT(seq, minSeq) {
			minSeq = seq
		}
	}
	for {
		rdSeq := atomic.LoadUint64(&s.rdSeq)
		if !qsbrLT(rdSeq, minSeq) {
			return rdSeq
		}
		if atomic.CompareAndSwapUint64(&s.rdSeq, rdSeq, minSeq) {
			return minSeq
		}
	}
}

func (s *QSBRShared) indexOf(qsbr *QSBRThreadState) int {
	for i := range s.array {
		if &s.array[i].qsbr == qsbr {
			return i
		}
	}
	return -1
}

func qsbrLT(a, b uint64) bool {
	return int64(a-b) < 0
}

func qsbrLEQ(a, b uint64) bool {
	return int64(a-b) <= 0
}
