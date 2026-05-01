package pycontext

import (
	"fmt"
	"sync"
)

const MaxWatchers = 8

type Event int

const (
	ContextSwitched Event = 1
)

type WatchCallback func(Event, *Context) error

type Context struct {
	mu      sync.RWMutex
	vars    map[*Var]any
	prev    *Context
	entered bool
}

type Var struct {
	name       string
	defaultVal any
}

type Token struct {
	ctx    *Context
	varRef *Var
	oldVal any
	hasOld bool
	used   bool
}

type Module struct {
	CopyContext func() *Context
	ContextType any
	VarType     any
	TokenType   any
}

var runtimeState = struct {
	mu       sync.Mutex
	current  *Context
	version  uint64
	watchers [MaxWatchers]WatchCallback
}{
	current: newEmptyContext(),
}

func New() *Context {
	return newEmptyContext()
}

func Copy(ctx *Context) *Context {
	if ctx == nil {
		return nil
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return newContextFromVars(ctx.vars)
}

func CopyCurrent() *Context {
	runtimeState.mu.Lock()
	defer runtimeState.mu.Unlock()
	if runtimeState.current == nil {
		runtimeState.current = newEmptyContext()
	}
	return newContextFromVars(runtimeState.current.snapshotLocked())
}

func Enter(ctx *Context) error {
	if ctx == nil {
		return fmt.Errorf("an instance of Context was expected")
	}
	runtimeState.mu.Lock()
	defer runtimeState.mu.Unlock()
	if ctx.entered {
		return fmt.Errorf("cannot enter context: context is already entered")
	}
	ctx.prev = runtimeState.current
	ctx.entered = true
	runtimeState.current = ctx
	runtimeState.version++
	notifyWatchersLocked(ContextSwitched, runtimeState.current)
	return nil
}

func Exit(ctx *Context) error {
	if ctx == nil {
		return fmt.Errorf("an instance of Context was expected")
	}
	runtimeState.mu.Lock()
	defer runtimeState.mu.Unlock()
	if !ctx.entered {
		return fmt.Errorf("cannot exit context: context has not been entered")
	}
	if runtimeState.current != ctx {
		return fmt.Errorf("cannot exit context: thread state references a different context object")
	}
	runtimeState.current = ctx.prev
	ctx.prev = nil
	ctx.entered = false
	runtimeState.version++
	notifyWatchersLocked(ContextSwitched, runtimeState.current)
	return nil
}

func (ctx *Context) Run(fn func() any) (any, error) {
	if fn == nil {
		return nil, fmt.Errorf("run() missing 1 required positional argument")
	}
	if err := Enter(ctx); err != nil {
		return nil, err
	}
	result := fn()
	if err := Exit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (ctx *Context) Get(key *Var, defaultValue any) (any, bool, error) {
	if key == nil {
		return nil, false, fmt.Errorf("a ContextVar key was expected")
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	value, ok := ctx.vars[key]
	if ok {
		return value, true, nil
	}
	if defaultValue != nil {
		return defaultValue, false, nil
	}
	return nil, false, nil
}

func (ctx *Context) Keys() []*Var {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	out := make([]*Var, 0, len(ctx.vars))
	for key := range ctx.vars {
		out = append(out, key)
	}
	return out
}

func (ctx *Context) Values() []any {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	out := make([]any, 0, len(ctx.vars))
	for _, value := range ctx.vars {
		out = append(out, value)
	}
	return out
}

func (ctx *Context) Items() [][2]any {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	out := make([][2]any, 0, len(ctx.vars))
	for key, value := range ctx.vars {
		out = append(out, [2]any{key, value})
	}
	return out
}

func (ctx *Context) Len() int {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return len(ctx.vars)
}

func (ctx *Context) Contains(key *Var) bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	_, ok := ctx.vars[key]
	return ok
}

func (ctx *Context) Copy() *Context {
	return Copy(ctx)
}

func NewVar(name string, def any) *Var {
	return &Var{name: name, defaultVal: def}
}

func (v *Var) Name() string {
	return v.name
}

func (v *Var) Default() any {
	return v.defaultVal
}

func Get(v *Var, def any) (any, error) {
	if v == nil {
		return nil, fmt.Errorf("an instance of ContextVar was expected")
	}
	runtimeState.mu.Lock()
	if runtimeState.current == nil {
		runtimeState.current = newEmptyContext()
	}
	current := runtimeState.current
	runtimeState.mu.Unlock()

	current.mu.RLock()
	value, ok := current.vars[v]
	current.mu.RUnlock()
	if ok {
		return value, nil
	}
	if def != nil {
		return def, nil
	}
	if v.defaultVal != nil {
		return v.defaultVal, nil
	}
	return nil, nil
}

func Set(v *Var, value any) (*Token, error) {
	if v == nil {
		return nil, fmt.Errorf("an instance of ContextVar was expected")
	}
	ctx := currentContext()
	ctx.mu.Lock()
	old, ok := ctx.vars[v]
	ctx.vars[v] = value
	ctx.mu.Unlock()
	return &Token{ctx: ctx, varRef: v, oldVal: old, hasOld: ok}, nil
}

func Reset(v *Var, tok *Token) error {
	if v == nil {
		return fmt.Errorf("an instance of ContextVar was expected")
	}
	if tok == nil {
		return fmt.Errorf("an instance of Token was expected")
	}
	if tok.used {
		return fmt.Errorf("token has already been used once")
	}
	if tok.varRef != v {
		return fmt.Errorf("token was created by a different ContextVar")
	}
	ctx := currentContext()
	if ctx != tok.ctx {
		return fmt.Errorf("token was created in a different Context")
	}
	tok.used = true
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if tok.hasOld {
		ctx.vars[v] = tok.oldVal
		return nil
	}
	if _, ok := ctx.vars[v]; !ok {
		return fmt.Errorf("lookup error")
	}
	delete(ctx.vars, v)
	return nil
}

func (v *Var) Get(defaultValue any) (any, error) {
	return Get(v, defaultValue)
}

func (v *Var) Set(value any) (*Token, error) {
	return Set(v, value)
}

func (v *Var) Reset(tok *Token) error {
	return Reset(v, tok)
}

func AddWatcher(callback WatchCallback) (int, error) {
	runtimeState.mu.Lock()
	defer runtimeState.mu.Unlock()
	for i, current := range runtimeState.watchers {
		if current == nil {
			runtimeState.watchers[i] = callback
			return i, nil
		}
	}
	return -1, fmt.Errorf("no more context watcher IDs available")
}

func ClearWatcher(id int) error {
	runtimeState.mu.Lock()
	defer runtimeState.mu.Unlock()
	if id < 0 || id >= MaxWatchers {
		return fmt.Errorf("invalid context watcher ID %d", id)
	}
	if runtimeState.watchers[id] == nil {
		return fmt.Errorf("no context watcher set for ID %d", id)
	}
	runtimeState.watchers[id] = nil
	return nil
}

func InitModule() Module {
	return Module{
		CopyContext: CopyCurrent,
		ContextType: (*Context)(nil),
		VarType:     (*Var)(nil),
		TokenType:   (*Token)(nil),
	}
}

func ResetRuntimeForTests() {
	runtimeState.mu.Lock()
	defer runtimeState.mu.Unlock()
	runtimeState.current = newEmptyContext()
	runtimeState.version = 0
	runtimeState.watchers = [MaxWatchers]WatchCallback{}
}

func currentContext() *Context {
	runtimeState.mu.Lock()
	defer runtimeState.mu.Unlock()
	if runtimeState.current == nil {
		runtimeState.current = newEmptyContext()
	}
	return runtimeState.current
}

func newEmptyContext() *Context {
	return &Context{vars: map[*Var]any{}}
}

func newContextFromVars(vars map[*Var]any) *Context {
	copyVars := make(map[*Var]any, len(vars))
	for key, value := range vars {
		copyVars[key] = value
	}
	return &Context{vars: copyVars}
}

func (ctx *Context) snapshotLocked() map[*Var]any {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	out := make(map[*Var]any, len(ctx.vars))
	for key, value := range ctx.vars {
		out[key] = value
	}
	return out
}

func notifyWatchersLocked(event Event, ctx *Context) {
	for _, watcher := range runtimeState.watchers {
		if watcher != nil {
			_ = watcher(event, ctx)
		}
	}
}
