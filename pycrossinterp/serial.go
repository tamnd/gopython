package pycrossinterp

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ScriptCode struct {
	Source          string
	Checked         bool
	Pure            bool
	ArgCount        int
	PosOnlyArgCount int
	KwOnlyArgCount  int
	VarArgs         bool
	VarKeywords     bool
	ReturnsValue    bool
	UsesGlobals     bool
	Stateless       bool
}

type ScriptFunction struct {
	Code      ScriptCode
	Stateless bool
}

type PickleContext struct {
	MainFile string
}

type SharedPickleData struct {
	Pickled []byte
	Context PickleContext
}

func PickleToXIData(obj any, mainFile string, xid *XIData) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("object could not be pickled: %w", err)
	}
	shared := SharedPickleData{
		Pickled: data,
		Context: PickleContext{MainFile: mainFile},
	}
	xid.Init(1, shared, obj, func(v any) (any, error) {
		return LoadPickleFromXIData(v.(*XIData))
	})
	return nil
}

func LoadPickleFromXIData(xid *XIData) (any, error) {
	shared, ok := xid.Data.(SharedPickleData)
	if !ok {
		return nil, fmt.Errorf("object could not be unpickled")
	}
	var out any
	if err := json.Unmarshal(shared.Pickled, &out); err != nil {
		return nil, fmt.Errorf("object could not be unpickled: %w", err)
	}
	return out, nil
}

func MarshalToXIData(obj any, xid *XIData) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("object could not be marshalled: %w", err)
	}
	xid.Init(1, data, obj, func(v any) (any, error) {
		return ReadMarshalFromXIData(v.(*XIData))
	})
	return nil
}

func ReadMarshalFromXIData(xid *XIData) (any, error) {
	data, ok := xid.Data.([]byte)
	if !ok {
		return nil, fmt.Errorf("object could not be unmarshalled")
	}
	var out any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("object could not be unmarshalled: %w", err)
	}
	return out, nil
}

func ScriptToXIData(obj any, pure bool, xid *XIData) error {
	script, err := normalizeScriptObject(obj, pure)
	if err != nil {
		return fmt.Errorf("object not a valid script: %w", err)
	}
	xid.Init(1, []byte(script), obj, func(v any) (any, error) {
		return string(v.(*XIData).Data.([]byte)), nil
	})
	return nil
}

func stringsContainsAny(text string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func normalizeScriptObject(obj any, pure bool) (string, error) {
	switch value := obj.(type) {
	case string:
		if pure && stringsContainsAny(value, []string{"global ", "nonlocal "}) {
			return "", fmt.Errorf("uses globals")
		}
		if stringsContainsAny(value, []string{"return "}) {
			return "", fmt.Errorf("returns value")
		}
		return value, nil
	case ScriptCode:
		if err := VerifyScript(value, pure); err != nil {
			return "", err
		}
		return value.Source, nil
	case ScriptFunction:
		code := value.Code
		if pure && !value.Stateless {
			return "", fmt.Errorf("function is not stateless")
		}
		code.Checked = value.Stateless
		code.Stateless = value.Stateless
		if err := VerifyScript(code, pure); err != nil {
			return "", err
		}
		return code.Source, nil
	default:
		return "", fmt.Errorf("unsupported script")
	}
}

func VerifyScript(code ScriptCode, pure bool) error {
	if pure && code.UsesGlobals {
		return fmt.Errorf("uses globals")
	}
	if pure && !code.Checked && !code.Stateless {
		return fmt.Errorf("not stateless")
	}
	if code.ArgCount > 0 || code.PosOnlyArgCount > 0 || code.KwOnlyArgCount > 0 || code.VarArgs || code.VarKeywords {
		return fmt.Errorf("code with args not supported")
	}
	if code.ReturnsValue {
		return fmt.Errorf("code that returns a value is not a script")
	}
	return nil
}
