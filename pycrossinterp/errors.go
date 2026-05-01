package pycrossinterp

import "fmt"

type InterpreterError struct{ Message string }

func (e *InterpreterError) Error() string { return e.Message }

type InterpreterNotFoundError struct{ Message string }

func (e *InterpreterNotFoundError) Error() string { return e.Message }

type NotShareableError struct {
	Message string
	Cause   error
}

func (e *NotShareableError) Error() string { return e.Message }

func SetNotShareableError(msg string) error {
	return &NotShareableError{Message: msg}
}

func FormatNotShareableError(format string, args ...any) error {
	return &NotShareableError{Message: fmt.Sprintf(format, args...)}
}

func UnwrapNotShareableError(err error) error {
	if err == nil {
		return nil
	}
	if ns, ok := err.(*NotShareableError); ok {
		if ns.Cause != nil {
			return ns.Cause
		}
		return ns
	}
	return nil
}
