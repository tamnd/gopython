package pythread

import "sync"

type Mutex struct {
	mu sync.Mutex
}

func (m *Mutex) Lock() {
	m.mu.Lock()
}

func (m *Mutex) Unlock() {
	m.mu.Unlock()
}

func (m *Mutex) LockFast() bool {
	return m.mu.TryLock()
}

type StopTheWorldState struct {
	WorldStopped bool
}

type InterpreterState struct {
	StopTheWorld StopTheWorldState
	QSBR         QSBRShared
	BRC          BRCState
}

func NewInterpreterState() *InterpreterState {
	interp := &InterpreterState{}
	interp.QSBR.init()
	return interp
}

type ThreadState struct {
	Interp               *InterpreterState
	HoldsGIL             bool
	ThreadID             uintptr
	ExplicitMergePending bool
	criticalSection      *criticalFrame
	qsbr                 *QSBRThreadState
	brc                  BRCThreadState
}

func NewThreadState(interp *InterpreterState) *ThreadState {
	return &ThreadState{Interp: interp}
}

func (t *ThreadState) QSBR() *QSBRThreadState {
	return t.qsbr
}
