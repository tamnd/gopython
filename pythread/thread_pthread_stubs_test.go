package pythread

import (
	"errors"
	"testing"
)

func TestStubTLSLifecycle(t *testing.T) {
	var state StubTLSState

	key, err := state.KeyCreate()
	if err != nil {
		t.Fatalf("KeyCreate returned error: %v", err)
	}
	if got := state.GetSpecific(key); got != nil {
		t.Fatalf("new key value = %v, want nil", got)
	}
	if err := state.SetSpecific(key, "value"); err != nil {
		t.Fatalf("SetSpecific returned error: %v", err)
	}
	if got := state.GetSpecific(key); got != "value" {
		t.Fatalf("GetSpecific = %v, want value", got)
	}
	if err := state.KeyDelete(key); err != nil {
		t.Fatalf("KeyDelete returned error: %v", err)
	}
	if got := state.GetSpecific(key); got != nil {
		t.Fatalf("deleted key value = %v, want nil", got)
	}
}

func TestStubTLSInvalidKey(t *testing.T) {
	var state StubTLSState
	if err := state.SetSpecific(7, "value"); err == nil {
		t.Fatal("SetSpecific should fail for an unused key")
	}
	if err := state.KeyDelete(7); err == nil {
		t.Fatal("KeyDelete should fail for an unused key")
	}
}

func TestStubTLSExhaustionAndSelf(t *testing.T) {
	var state StubTLSState
	for i := 0; i < PthreadKeysMax; i++ {
		if _, err := state.KeyCreate(); err != nil {
			t.Fatalf("KeyCreate failed at %d: %v", i, err)
		}
	}
	if _, err := state.KeyCreate(); !errors.Is(err, ErrThreadUnavailable) {
		t.Fatalf("unexpected exhaustion error: %v", err)
	}
	if state.Self() == 0 {
		t.Fatal("Self should return a stable non-zero identity")
	}
}

func TestStartStubThreadFailsExplicitly(t *testing.T) {
	if _, err := StartStubThread(func(any) {}, nil); !errors.Is(err, ErrThreadUnavailable) {
		t.Fatalf("unexpected stub thread error: %v", err)
	}
}
