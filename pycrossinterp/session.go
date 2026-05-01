package pycrossinterp

import "fmt"

type ErrCode int

const (
	ErrNoError ErrCode = iota
	ErrOther
	ErrNoMemory
	ErrAlreadyRunning
	ErrMainNSFailure
	ErrApplyNSFailure
	ErrPreserveFailure
	ErrExcPropagationFailure
	ErrNotShareable
	ErrUncaughtException
)

type Failure struct {
	Code ErrCode
	Msg  string
}

type ErrorState struct {
	InterpID int64
	Override *Failure
	Uncaught ExcInfo
}

type SessionResult struct {
	ExcInfo   any
	Preserved map[string]any
	ErrCode   ErrCode
}

type Session struct {
	Status       int
	Switched     bool
	PrevInterpID int64
	InitInterpID int64
	Running      bool
	MainNS       map[string]any
	Preserved    map[string]any
}

const (
	SessionUnused = 0
	SessionActive = 1
)

type SessionHooks struct {
	CurrentInterpID  func() int64
	SwitchTo         func(int64) error
	SwitchBack       func(int64) error
	SetRunningMain   func(int64) error
	ClearRunningMain func(int64)
	GetMainNamespace func(int64) (map[string]any, error)
}

func NewSession() *Session {
	return &Session{}
}

func FreeSession(session *Session) error {
	if session == nil {
		return nil
	}
	if session.Status != SessionUnused {
		return fmt.Errorf("session still active")
	}
	return nil
}

func sessionActive(session *Session) bool {
	return session != nil && session.Status == SessionActive
}

func (session *Session) Preserve(name string, value any) {
	if session.Preserved == nil {
		session.Preserved = map[string]any{}
	}
	session.Preserved[name] = value
}

func (session *Session) GetPreserved(name string) (any, bool) {
	if session.Preserved == nil {
		return nil, false
	}
	value, ok := session.Preserved[name]
	return value, ok
}

func enterSession(session *Session, interpID int64, hooks SessionHooks) error {
	if session == nil {
		return fmt.Errorf("missing session")
	}
	if session.Status != SessionUnused {
		return fmt.Errorf("session already active")
	}
	currentID := int64(0)
	if hooks.CurrentInterpID != nil {
		currentID = hooks.CurrentInterpID()
	}
	switched := currentID != interpID
	if switched && hooks.SwitchTo != nil {
		if err := hooks.SwitchTo(interpID); err != nil {
			return err
		}
	}
	*session = Session{
		Status:       SessionActive,
		Switched:     switched,
		PrevInterpID: currentID,
		InitInterpID: interpID,
	}
	return nil
}

func exitSession(session *Session, hooks SessionHooks) error {
	if !sessionActive(session) {
		return nil
	}
	if session.Running && hooks.ClearRunningMain != nil {
		hooks.ClearRunningMain(session.InitInterpID)
	}
	if session.Switched && hooks.SwitchBack != nil {
		if err := hooks.SwitchBack(session.PrevInterpID); err != nil {
			return err
		}
	}
	*session = Session{}
	return nil
}

func PropagateNotShareableError(err error, override *Failure) {
	if override == nil {
		return
	}
	if _, ok := err.(*NotShareableError); ok {
		override.Code = ErrNotShareable
	}
}

func NewFailure() *Failure {
	return &Failure{Code: ErrNoError}
}

func FreeFailure(failure *Failure) {
	if failure == nil {
		return
	}
	*failure = Failure{Code: ErrNoError}
}

func GetFailureCode(failure *Failure) ErrCode {
	if failure == nil {
		return ErrNoError
	}
	return failure.Code
}

func InitFailureUTF8(failure *Failure, code ErrCode, msg string) {
	if failure == nil {
		return
	}
	*failure = Failure{Code: code, Msg: msg}
}

func InitFailure(failure *Failure, code ErrCode, obj any) error {
	if failure == nil {
		return nil
	}
	*failure = Failure{Code: code}
	if obj == nil {
		return nil
	}
	failure.Msg = fmt.Sprint(obj)
	return nil
}

func CaptureErrorState(interpID int64, raised error, override *Failure) ErrorState {
	state := ErrorState{InterpID: interpID}
	if override != nil && override.Code != ErrNoError && override.Code != ErrUncaughtException {
		copy := *override
		state.Override = &copy
	}
	if raised != nil {
		info := InitExcInfoFromException("Exception", "Exception", "builtins", raised.Error(), raised.Error())
		state.Uncaught = *info
		if state.Override == nil {
			state.Override = &Failure{Code: ErrUncaughtException}
		}
	}
	return state
}

func ApplyErrorCode(code ErrCode) error {
	switch code {
	case ErrOther:
		return &InterpreterError{Message: ""}
	case ErrNoMemory:
		return fmt.Errorf("out of memory")
	case ErrAlreadyRunning:
		return &InterpreterError{Message: "interpreter already running"}
	case ErrMainNSFailure:
		return &InterpreterError{Message: "failed to get __main__ namespace"}
	case ErrApplyNSFailure:
		return &InterpreterError{Message: "failed to apply namespace to __main__"}
	case ErrPreserveFailure:
		return &InterpreterError{Message: "failed to preserve objects across session"}
	case ErrExcPropagationFailure:
		return &InterpreterError{Message: "failed to transfer exception between interpreters"}
	case ErrNotShareable:
		return SetXIDataLookupFailure(nil, "", nil)
	default:
		return fmt.Errorf("unsupported error code %d", code)
	}
}

func ApplyError(state ErrorState, failure string) (any, error) {
	if failure != "" {
		return nil, nil
	}
	code := ErrUncaughtException
	if state.Override != nil {
		code = state.Override.Code
	}
	if code == ErrUncaughtException {
		return state.Uncaught.AsObject(), nil
	}
	if code == ErrNotShareable {
		msg := state.Uncaught.Message
		if state.Override != nil && state.Override.Msg != "" {
			msg = state.Override.Msg
		}
		return nil, &NotShareableError{Message: msg}
	}
	err := ApplyErrorCode(code)
	if state.Uncaught.IsSet() {
		return nil, fmt.Errorf("%w: %s", err, state.Uncaught.Apply())
	}
	return nil, err
}

func ensureMainNS(session *Session, hooks SessionHooks, override *Failure) error {
	if session.MainNS != nil {
		return nil
	}
	if hooks.GetMainNamespace == nil {
		if override != nil {
			override.Code = ErrMainNSFailure
		}
		return fmt.Errorf("missing main namespace hook")
	}
	ns, err := hooks.GetMainNamespace(session.InitInterpID)
	if err != nil {
		if override != nil {
			override.Code = ErrMainNSFailure
		}
		return err
	}
	session.MainNS = ns
	return nil
}

func GetMainNamespace(session *Session, hooks SessionHooks, failure *Failure) (map[string]any, error) {
	if !sessionActive(session) {
		return nil, fmt.Errorf("session not active")
	}
	if err := ensureMainNS(session, hooks, failure); err != nil {
		return nil, err
	}
	return session.MainNS, nil
}

func GetPreserved(result *SessionResult, name string) any {
	if result == nil || result.Preserved == nil {
		return nil
	}
	return result.Preserved[name]
}

func ClearResult(result *SessionResult) {
	if result == nil {
		return
	}
	result.Preserved = nil
	result.ExcInfo = nil
	result.ErrCode = ErrNoError
}

func popPreserved(session *Session, failure *Failure) (*SharedNamespace, map[string]any, error) {
	if session.Preserved == nil {
		return nil, nil, nil
	}
	if !session.Switched {
		preserved := session.Preserved
		session.Preserved = nil
		return nil, preserved, nil
	}
	if len(session.Preserved) == 0 {
		session.Preserved = nil
		return nil, nil, nil
	}
	names := make([]string, 0, len(session.Preserved))
	for name := range session.Preserved {
		names = append(names, name)
	}
	shared, err := CreateSharedNamespace(names)
	if err != nil {
		if failure != nil {
			failure.Code = ErrPreserveFailure
		}
		return nil, nil, err
	}
	if err := shared.Fill(session.Preserved, XIDataFullFallback); err != nil {
		if failure != nil {
			if _, ok := err.(*NotShareableError); ok {
				failure.Code = ErrNotShareable
				failure.Msg = err.Error()
			} else {
				failure.Code = ErrPreserveFailure
			}
		}
		shared.Destroy(session.InitInterpID, nil, nil)
		return nil, nil, err
	}
	session.Preserved = nil
	return shared, nil, nil
}

func finishPreserved(shared *SharedNamespace) (map[string]any, error) {
	if shared == nil {
		return nil, nil
	}
	defer shared.Destroy(0, nil, nil)
	ns := map[string]any{}
	if err := shared.Apply(ns, nil); err != nil {
		return nil, err
	}
	return ns, nil
}

func Enter(session *Session, interpID int64, nsupdates map[string]any, hooks SessionHooks, result *SessionResult) error {
	var sharedns *SharedNamespace
	if len(nsupdates) > 0 {
		names := make([]string, 0, len(nsupdates))
		for name := range nsupdates {
			names = append(names, name)
		}
		var err error
		sharedns, err = CreateSharedNamespace(names)
		if err != nil {
			if result != nil {
				result.ErrCode = ErrApplyNSFailure
			}
			return err
		}
		if err := sharedns.Fill(nsupdates, XIDataFullFallback); err != nil {
			sharedns.Destroy(0, nil, nil)
			if result != nil {
				result.ErrCode = ErrApplyNSFailure
			}
			return err
		}
	}

	if err := enterSession(session, interpID, hooks); err != nil {
		return err
	}
	override := Failure{Code: ErrUncaughtException}

	if hooks.SetRunningMain != nil {
		if err := hooks.SetRunningMain(interpID); err != nil {
			override.Code = ErrAlreadyRunning
			state := CaptureErrorState(interpID, err, &override)
			_ = exitSession(session, hooks)
			if sharedns != nil {
				sharedns.Destroy(interpID, nil, nil)
			}
			excinfo, applyErr := ApplyError(state, "")
			if result != nil {
				result.ExcInfo = excinfo
				result.ErrCode = override.Code
			}
			return applyErr
		}
	}
	session.Running = true

	if sharedns != nil {
		if err := ensureMainNS(session, hooks, &override); err != nil {
			state := CaptureErrorState(interpID, err, &override)
			_ = exitSession(session, hooks)
			sharedns.Destroy(interpID, nil, nil)
			excinfo, applyErr := ApplyError(state, "")
			if result != nil {
				result.ExcInfo = excinfo
				result.ErrCode = override.Code
			}
			return applyErr
		}
		if err := sharedns.Apply(session.MainNS, nil); err != nil {
			override.Code = ErrApplyNSFailure
			state := CaptureErrorState(interpID, err, &override)
			_ = exitSession(session, hooks)
			sharedns.Destroy(interpID, nil, nil)
			excinfo, applyErr := ApplyError(state, "")
			if result != nil {
				result.ExcInfo = excinfo
				result.ErrCode = override.Code
			}
			return applyErr
		}
		sharedns.Destroy(interpID, nil, nil)
	}
	return nil
}

func Exit(session *Session, raised error, override *Failure, hooks SessionHooks, result *SessionResult) error {
	if session == nil {
		return fmt.Errorf("missing session")
	}
	state := ErrorState{}
	if raised != nil || (override != nil && override.Code != ErrNoError) {
		state = CaptureErrorState(session.InitInterpID, raised, override)
	}
	var localFailure Failure
	sharedPreserved, directPreserved, popErr := popPreserved(session, &localFailure)
	if popErr != nil && override == nil {
		override = &localFailure
	}
	if err := exitSession(session, hooks); err != nil {
		return err
	}
	preserved := directPreserved
	if sharedPreserved != nil {
		var err error
		preserved, err = finishPreserved(sharedPreserved)
		if err != nil && override == nil {
			override = &Failure{Code: ErrPreserveFailure}
			state = CaptureErrorState(session.InitInterpID, err, override)
		}
	}
	if result != nil {
		result.Preserved = preserved
		if state.Override != nil {
			result.ErrCode = state.Override.Code
		}
	}
	if state.Override == nil && !state.Uncaught.IsSet() {
		return nil
	}
	excinfo, err := ApplyError(state, "")
	if result != nil {
		result.ExcInfo = excinfo
		if result.ErrCode == ErrNoError && state.Override != nil {
			result.ErrCode = state.Override.Code
		}
	}
	return err
}
