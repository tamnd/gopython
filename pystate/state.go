package pystate

import "sync"

type ThreadStatus struct {
	Initialized   bool
	Bound         bool
	Unbound       bool
	BoundGILState bool
	Active        bool
	Finalizing    bool
	Cleared       bool
	Finalized     bool
}

type ThreadState struct {
	Interp         *InterpreterState
	ThreadID       uint64
	NativeThreadID uint64
	Status         ThreadStatus
}

type InterpreterState struct {
	ID                int64
	Runtime           *RuntimeState
	RunningMain       bool
	ContextWatchers   [8]func(any) error
	ActiveContextBits uint8
	contextWatchersMu sync.Mutex
}

type RuntimeState struct {
	mu              sync.Mutex
	initialized     bool
	mainThread      uint64
	nextInterpreter int64
	interpreters    []*InterpreterState
	current         *ThreadState
}

func NewRuntimeState() *RuntimeState {
	return &RuntimeState{}
}

func (runtime *RuntimeState) Init(mainThread uint64) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	runtime.initialized = true
	runtime.mainThread = mainThread
}

func (runtime *RuntimeState) Initialized() bool {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	return runtime.initialized
}

func (runtime *RuntimeState) EnableInterpreters() {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	runtime.nextInterpreter = 0
}

func (runtime *RuntimeState) NewInterpreter() *InterpreterState {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	interp := &InterpreterState{
		ID:      runtime.nextInterpreter,
		Runtime: runtime,
	}
	runtime.nextInterpreter++
	runtime.interpreters = append(runtime.interpreters, interp)
	return interp
}

func (runtime *RuntimeState) Interpreters() []*InterpreterState {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	out := make([]*InterpreterState, len(runtime.interpreters))
	copy(out, runtime.interpreters)
	return out
}

func NewThreadState(interp *InterpreterState) *ThreadState {
	return &ThreadState{Interp: interp}
}

func BindThread(ts *ThreadState, threadID, nativeThreadID uint64) {
	ts.ThreadID = threadID
	ts.NativeThreadID = nativeThreadID
	ts.Status.Bound = true
}

func UnbindThread(ts *ThreadState) {
	ts.Status.Unbound = true
}

func BindGILState(runtime *RuntimeState, ts *ThreadState) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.current != nil {
		runtime.current.Status.BoundGILState = false
	}
	runtime.current = ts
	ts.Status.BoundGILState = true
}

func UnbindGILState(runtime *RuntimeState, ts *ThreadState) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.current == ts {
		runtime.current = nil
	}
	ts.Status.BoundGILState = false
}

func CurrentThreadState(runtime *RuntimeState) *ThreadState {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	return runtime.current
}

func SetRunningMain(interp *InterpreterState) error {
	if interp == nil {
		return ErrNoInterpreter
	}
	if interp.RunningMain {
		return nil
	}
	interp.RunningMain = true
	return nil
}

func ClearRunningMain(interp *InterpreterState) {
	if interp != nil {
		interp.RunningMain = false
	}
}

func AddContextWatcher(interp *InterpreterState, cb func(any) error) (int, error) {
	interp.contextWatchersMu.Lock()
	defer interp.contextWatchersMu.Unlock()
	for i, current := range interp.ContextWatchers {
		if current == nil {
			interp.ContextWatchers[i] = cb
			interp.ActiveContextBits |= 1 << i
			return i, nil
		}
	}
	return -1, ErrNoWatcherSlot
}

func ClearContextWatcher(interp *InterpreterState, id int) error {
	interp.contextWatchersMu.Lock()
	defer interp.contextWatchersMu.Unlock()
	if id < 0 || id >= len(interp.ContextWatchers) {
		return ErrInvalidWatcherID
	}
	if interp.ContextWatchers[id] == nil {
		return ErrMissingWatcher
	}
	interp.ContextWatchers[id] = nil
	interp.ActiveContextBits &^= 1 << id
	return nil
}
