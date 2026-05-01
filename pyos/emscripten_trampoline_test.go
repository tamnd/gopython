package pyos

import (
	"errors"
	"testing"
)

func TestEmscriptenTrampolineCallFallbackSignatures(t *testing.T) {
	t.Cleanup(func() { SetEmscriptenTrampoline(nil) })

	result, err := EmscriptenTrampolineCall(func(self any, args any, kw any) any {
		return []any{self, args, kw}
	}, "self", "args", "kw")
	if err != nil {
		t.Fatalf("three-arg fallback returned error: %v", err)
	}
	got := result.([]any)
	if len(got) != 3 || got[0] != "self" || got[1] != "args" || got[2] != "kw" {
		t.Fatalf("unexpected three-arg result: %#v", got)
	}

	result, err = EmscriptenTrampolineCall(func(self any, args any) any {
		return []any{self, args}
	}, "self", "args", "kw")
	if err != nil {
		t.Fatalf("two-arg fallback returned error: %v", err)
	}
	if got := result.([]any); len(got) != 2 || got[0] != "self" || got[1] != "args" {
		t.Fatalf("unexpected two-arg result: %#v", got)
	}

	result, err = EmscriptenTrampolineCall(func(self any) any {
		return self
	}, "self", "args", "kw")
	if err != nil || result != "self" {
		t.Fatalf("one-arg fallback = (%v, %v), want (self, nil)", result, err)
	}

	result, err = EmscriptenTrampolineCall(func() any {
		return "zero"
	}, "self", "args", "kw")
	if err != nil || result != "zero" {
		t.Fatalf("zero-arg fallback = (%v, %v), want (zero, nil)", result, err)
	}
}

func TestEmscriptenTrampolineCallReturnsArgumentError(t *testing.T) {
	t.Cleanup(func() { SetEmscriptenTrampoline(nil) })

	_, err := EmscriptenTrampolineCall(42, "self", "args", "kw")
	if !errors.Is(err, errTooManyArguments) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmscriptenTrampolineCallUsesConfiguredTrampoline(t *testing.T) {
	t.Cleanup(func() { SetEmscriptenTrampoline(nil) })

	SetEmscriptenTrampoline(func(callable any, self any, args any, kw any) (any, error) {
		return []any{callable, self, args, kw}, nil
	})

	result, err := EmscriptenTrampolineCall("callable", "self", "args", "kw")
	if err != nil {
		t.Fatalf("configured trampoline returned error: %v", err)
	}
	got := result.([]any)
	if len(got) != 4 || got[0] != "callable" || got[3] != "kw" {
		t.Fatalf("unexpected configured trampoline result: %#v", got)
	}
}
