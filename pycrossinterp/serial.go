package pycrossinterp

import (
	"encoding/json"
	"fmt"
	"strings"
)

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
	script, ok := obj.(string)
	if !ok {
		return fmt.Errorf("object not a valid script")
	}
	if pure && stringsContainsAny(script, []string{"global ", "nonlocal "}) {
		return fmt.Errorf("object not a valid script")
	}
	if stringsContainsAny(script, []string{"return "}) {
		return fmt.Errorf("object not a valid script")
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
