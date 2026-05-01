package pycrossinterp

import (
	"errors"
	"testing"
)

func TestSessionEnterAndExit(t *testing.T) {
	session := NewSession()
	current := int64(1)
	mainNS := map[string]any{}
	cleared := int64(0)
	if err := Enter(session, 2, map[string]any{}, SessionHooks{
		CurrentInterpID: func() int64 { return current },
		SwitchTo: func(id int64) error {
			current = id
			return nil
		},
		SetRunningMain: func(id int64) error { return nil },
		GetMainNamespace: func(id int64) (map[string]any, error) {
			return mainNS, nil
		},
		ClearRunningMain: func(id int64) { cleared = id },
		SwitchBack: func(id int64) error {
			current = id
			return nil
		},
	}, nil); err != nil {
		t.Fatalf("Enter error: %v", err)
	}
	if !sessionActive(session) || !session.Switched || current != 2 {
		t.Fatalf("session = %#v current=%d", session, current)
	}
	session.Preserve("x", 1)
	result := &SessionResult{}
	if err := Exit(session, nil, nil, SessionHooks{
		ClearRunningMain: func(id int64) { cleared = id },
		SwitchBack: func(id int64) error {
			current = id
			return nil
		},
	}, result); err != nil {
		t.Fatalf("Exit error: %v", err)
	}
	if current != 1 || cleared != 2 || result.Preserved["x"] != float64(1) {
		t.Fatalf("current=%d cleared=%d result=%#v", current, cleared, result)
	}
}

func TestSessionEnterApplyNamespace(t *testing.T) {
	session := NewSession()
	mainNS := map[string]any{}
	err := Enter(session, 1, map[string]any{"a": map[string]any{"x": 1}}, SessionHooks{
		CurrentInterpID: func() int64 { return 1 },
		SetRunningMain:  func(int64) error { return nil },
		GetMainNamespace: func(id int64) (map[string]any, error) {
			return mainNS, nil
		},
	}, nil)
	if err != nil {
		t.Fatalf("Enter error: %v", err)
	}
	got := mainNS["a"].(map[string]any)
	if got["x"] != float64(1) {
		t.Fatalf("mainNS = %#v", mainNS)
	}
}

func TestGetMainNamespaceAndFailureHelpers(t *testing.T) {
	session := &Session{
		Status:       SessionActive,
		InitInterpID: 3,
	}
	ns, err := GetMainNamespace(session, SessionHooks{
		GetMainNamespace: func(id int64) (map[string]any, error) {
			return map[string]any{"x": 1}, nil
		},
	}, nil)
	if err != nil || ns["x"] != 1 {
		t.Fatalf("GetMainNamespace = (%#v, %v)", ns, err)
	}

	failure := NewFailure()
	InitFailureUTF8(failure, ErrPreserveFailure, "bad")
	if GetFailureCode(failure) != ErrPreserveFailure || failure.Msg != "bad" {
		t.Fatalf("failure = %#v", failure)
	}
	if err := InitFailure(failure, ErrApplyNSFailure, 123); err != nil || failure.Msg != "123" {
		t.Fatalf("failure = %#v err=%v", failure, err)
	}
	FreeFailure(failure)
	if failure.Code != ErrNoError || failure.Msg != "" {
		t.Fatalf("failure after free = %#v", failure)
	}

	if err := Preserve(session, "y", 2, nil); err != nil {
		t.Fatalf("Preserve returned error: %v", err)
	}
	if got, ok := session.GetPreserved("y"); !ok || got != 2 {
		t.Fatalf("preserved = (%v, %t)", got, ok)
	}
	if err := Preserve(NewSession(), "z", 3, nil); err == nil {
		t.Fatal("expected inactive preserve failure")
	}
}

func TestSessionEnterFailurePropagates(t *testing.T) {
	session := NewSession()
	result := &SessionResult{}
	err := Enter(session, 3, nil, SessionHooks{
		CurrentInterpID: func() int64 { return 3 },
		SetRunningMain: func(int64) error {
			return errors.New("already running")
		},
	}, result)
	if err == nil {
		t.Fatal("expected enter failure")
	}
	if result.ErrCode != ErrAlreadyRunning {
		t.Fatalf("result = %#v", result)
	}
}

func TestSessionExitPropagatesErrors(t *testing.T) {
	session := &Session{
		Status:       SessionActive,
		InitInterpID: 5,
		PrevInterpID: 5,
		Running:      true,
	}
	result := &SessionResult{}
	override := &Failure{Code: ErrNotShareable, Msg: "custom"}
	err := Exit(session, errors.New("boom"), override, SessionHooks{
		ClearRunningMain: func(int64) {},
	}, result)
	var ns *NotShareableError
	if !errors.As(err, &ns) {
		t.Fatalf("err = %v", err)
	}
	if result.ErrCode != ErrNotShareable {
		t.Fatalf("result = %#v", result)
	}
}

func TestSessionPreserveAcrossSwitch(t *testing.T) {
	session := &Session{
		Status:       SessionActive,
		Switched:     true,
		InitInterpID: 5,
		PrevInterpID: 1,
		Running:      true,
	}
	session.Preserve("x", map[string]any{"a": 1})
	result := &SessionResult{}
	err := Exit(session, nil, nil, SessionHooks{
		ClearRunningMain: func(int64) {},
		SwitchBack:       func(int64) error { return nil },
	}, result)
	if err != nil {
		t.Fatalf("Exit error: %v", err)
	}
	got := result.Preserved["x"].(map[string]any)
	if got["a"] != float64(1) {
		t.Fatalf("result = %#v", result)
	}
	if GetPreserved(result, "x") == nil {
		t.Fatalf("GetPreserved = %#v", result)
	}
	ClearResult(result)
	if result.Preserved != nil || result.ExcInfo != nil || result.ErrCode != ErrNoError {
		t.Fatalf("result after clear = %#v", result)
	}
}

func TestApplyErrorCodeAndFailureCapture(t *testing.T) {
	if err := ApplyErrorCode(ErrApplyNSFailure); err == nil || err.Error() != "failed to apply namespace to __main__" {
		t.Fatalf("ApplyErrorCode = %v", err)
	}
	override := &Failure{Code: ErrPreserveFailure, Msg: "preserve"}
	state := CaptureErrorState(7, errors.New("bad"), override)
	if state.Override == nil || state.Override.Code != ErrPreserveFailure || !state.Uncaught.IsSet() {
		t.Fatalf("state = %#v", state)
	}
}
