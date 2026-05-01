package pyos

import (
	"errors"
	"sync"
)

var errTooManyArguments = errors.New("handler takes too many arguments")

var trampolineState struct {
	mu sync.RWMutex
	fn func(callable any, self any, args any, kw any) (any, error)
}

func SetEmscriptenTrampoline(fn func(callable any, self any, args any, kw any) (any, error)) {
	trampolineState.mu.Lock()
	trampolineState.fn = fn
	trampolineState.mu.Unlock()
}

func EmscriptenTrampolineCall(callable any, self any, args any, kw any) (any, error) {
	trampolineState.mu.RLock()
	fn := trampolineState.fn
	trampolineState.mu.RUnlock()
	if fn == nil {
		return callTrampolineFallback(callable, self, args, kw)
	}
	return fn(callable, self, args, kw)
}

func callTrampolineFallback(callable any, self any, args any, kw any) (any, error) {
	switch fn := callable.(type) {
	case func(any, any, any) any:
		return fn(self, args, kw), nil
	case func(any, any) any:
		return fn(self, args), nil
	case func(any) any:
		return fn(self), nil
	case func() any:
		return fn(), nil
	default:
		return nil, errTooManyArguments
	}
}
