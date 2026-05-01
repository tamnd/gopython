package pycrossinterp

import (
	"errors"
	"testing"
)

func TestLookupState(t *testing.T) {
	var lookup LookupState
	lookup.Init()
	lookup.Global.Add("", func(v any) (any, error) { return len(v.(string)), nil })
	lookup.Local.Add(0, func(v any) (any, error) { return v.(int) + 1, nil })
	if fn := lookup.Lookup("abc"); fn == nil {
		t.Fatal("expected global lookup")
	} else if value, err := fn("abc"); err != nil || value != 3 {
		t.Fatalf("global lookup = (%v, %v)", value, err)
	}
	if fn := lookup.Lookup(4); fn == nil {
		t.Fatal("expected local lookup")
	} else if value, err := fn(4); err != nil || value != 5 {
		t.Fatalf("local lookup = (%v, %v)", value, err)
	}
	lookup.Fini()
	if fn := lookup.Lookup("abc"); fn != nil {
		t.Fatal("expected lookup cleared after fini")
	}
}

func TestNotShareableErrors(t *testing.T) {
	err := SetNotShareableError("not shareable")
	if err == nil || err.Error() != "not shareable" {
		t.Fatalf("SetNotShareableError = %v", err)
	}
	err = FormatNotShareableError("bad %s", "value")
	if err == nil || err.Error() != "bad value" {
		t.Fatalf("FormatNotShareableError = %v", err)
	}
	root := errors.New("root")
	ns := &NotShareableError{Message: "wrapped", Cause: root}
	if got := UnwrapNotShareableError(ns); got != root {
		t.Fatalf("UnwrapNotShareableError = %v, want root", got)
	}
	if got := UnwrapNotShareableError(errors.New("plain")); got != nil {
		t.Fatalf("UnwrapNotShareableError plain = %v, want nil", got)
	}
}
