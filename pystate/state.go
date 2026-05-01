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
	Whence         int
	Dict           map[string]any
	Status         ThreadStatus
}

type InterpreterState struct {
	ID                int64
	Runtime           *RuntimeState
	Ready             bool
	RunningMain       bool
	Whence            int
	Finalizing        bool
	MainModule        any
	Dict              map[string]any
	IDRefCount        int
	RequiresIDRefFlag bool
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
	main            *ThreadState
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
		Ready:   true,
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
	if runtime.main == nil {
		runtime.main = ts
	}
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

func MainThreadState(runtime *RuntimeState) *ThreadState {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	return runtime.main
}

func SetMainThreadState(runtime *RuntimeState, ts *ThreadState) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	runtime.main = ts
}

func RemoveExcept(runtime *RuntimeState, keep *ThreadState) []*ThreadState {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	var removed []*ThreadState
	if keep == nil {
		removed = append(removed, runtime.current)
		runtime.current = nil
		runtime.main = nil
		return removed
	}
	if runtime.current != nil && runtime.current != keep {
		removed = append(removed, runtime.current)
	}
	runtime.current = keep
	return removed
}

func DeleteInterpreter(runtime *RuntimeState, interp *InterpreterState) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	out := runtime.interpreters[:0]
	for _, current := range runtime.interpreters {
		if current != interp {
			out = append(out, current)
		}
	}
	runtime.interpreters = out
	if runtime.main != nil && runtime.main.Interp == interp {
		runtime.main = nil
	}
	if runtime.current != nil && runtime.current.Interp == interp {
		runtime.current = nil
	}
}

func RuntimeStateFini(runtime *RuntimeState) {
	if runtime == nil {
		return
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	runtime.initialized = false
	runtime.mainThread = 0
	runtime.interpreters = nil
	runtime.main = nil
	runtime.current = nil
}

func ClearInterpreter(interp *InterpreterState) {
	if interp == nil {
		return
	}
	interp.MainModule = nil
	interp.Dict = nil
	interp.RunningMain = false
	interp.Finalizing = false
	interp.Ready = false
	interp.ActiveContextBits = 0
	interp.ContextWatchers = [8]func(any) error{}
}

func ClearInterpreterFromThread(ts *ThreadState) {
	if ts == nil {
		return
	}
	ClearInterpreter(ts.Interp)
}

func DeleteInterpreterState(interp *InterpreterState) {
	if interp == nil || interp.Runtime == nil {
		return
	}
	DeleteInterpreter(interp.Runtime, interp)
}

func SetInterpreterAlreadyRunning() error {
	return ErrInterpreterAlreadyRunning
}

func SetRunningMain(interp *InterpreterState) error {
	if interp == nil {
		return ErrNoInterpreter
	}
	if interp.RunningMain {
		return ErrInterpreterAlreadyRunning
	}
	interp.RunningMain = true
	return nil
}

func ClearRunningMain(interp *InterpreterState) {
	if interp != nil {
		interp.RunningMain = false
	}
}

func IsRunningMain(interp *InterpreterState) bool {
	return interp != nil && interp.RunningMain
}

func ThreadIsRunningMain(ts *ThreadState) bool {
	return ts != nil && ts.Interp != nil && ts.Interp.RunningMain && ts == MainThreadState(ts.Interp.Runtime)
}

func ReinitRunningMain(ts *ThreadState) {
	if ts == nil || ts.Interp == nil {
		return
	}
	if !ThreadIsRunningMain(ts) {
		ts.Interp.RunningMain = false
	}
}

func IsInterpreterReady(interp *InterpreterState) bool {
	return interp != nil && interp.Ready
}

func SetInterpreterWhence(interp *InterpreterState, whence int) {
	if interp != nil {
		interp.Whence = whence
	}
}

func GetMainModule(ts *ThreadState) any {
	if ts == nil || ts.Interp == nil {
		return nil
	}
	return ts.Interp.MainModule
}

func CheckMainModule(module any) error {
	if module == nil {
		return ErrMainModuleNotFound
	}
	m, ok := module.(map[string]any)
	if !ok {
		return ErrInvalidMainModule
	}
	name, ok := m["__name__"].(string)
	if !ok || name != "__main__" {
		return ErrInvalidMainModule
	}
	return nil
}

func InterpreterGetDict(interp *InterpreterState) map[string]any {
	if interp == nil {
		return nil
	}
	if interp.Dict == nil {
		interp.Dict = map[string]any{}
	}
	return interp.Dict
}

func ThreadGetDict(ts *ThreadState) map[string]any {
	if ts == nil {
		return nil
	}
	if ts.Dict == nil {
		ts.Dict = map[string]any{}
	}
	return ts.Dict
}

func GetInterpreterIDObject(interp *InterpreterState) any {
	if interp == nil || interp.ID < 0 {
		return nil
	}
	return interp.ID
}

func InterpreterIDIncref(interp *InterpreterState) {
	if interp != nil {
		interp.IDRefCount++
	}
}

func InterpreterIDDecref(interp *InterpreterState) {
	if interp != nil && interp.IDRefCount > 0 {
		interp.IDRefCount--
	}
}

func InterpreterRequiresIDRef(interp *InterpreterState) bool {
	return interp != nil && interp.RequiresIDRefFlag
}

func RequireInterpreterIDRef(interp *InterpreterState, required bool) {
	if interp != nil {
		interp.RequiresIDRefFlag = required
	}
}

func AttachThread(ts *ThreadState) {
	if ts != nil {
		ts.Status.Active = true
	}
}

func DetachThread(ts *ThreadState) {
	if ts != nil {
		ts.Status.Active = false
	}
}

func SuspendThread(ts *ThreadState) {
	if ts != nil {
		ts.Status.Active = false
	}
}

func SetThreadShuttingDown(ts *ThreadState) {
	if ts != nil {
		ts.Status.Finalizing = true
	}
}

func IsMainThread(runtime *RuntimeState, ts *ThreadState) bool {
	return runtime != nil && ts != nil && MainThreadState(runtime) == ts
}

func IsMainInterpreterFinalizing(interp *InterpreterState) bool {
	return interp != nil && interp.Finalizing
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
