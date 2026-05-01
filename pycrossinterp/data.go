package pycrossinterp

import "fmt"

type XIData struct {
	InterpID  int64
	Data      any
	Obj       any
	NewObject func(any) (any, error)
	Free      func(any)
}

func NewXIData() *XIData {
	return &XIData{InterpID: -1}
}

func FreeXIData(interpID int64, x *XIData) error {
	if x == nil {
		return nil
	}
	if err := x.Clear(interpID); err != nil {
		return err
	}
	return nil
}

func ReleaseXIData(x *XIData, currentID int64, lookup func(int64) bool, schedule func(int64, func(any) int, any)) int {
	if x == nil {
		return 0
	}
	if (x.Data == nil || x.Free == nil) && x.Obj == nil {
		x.Data = nil
		return 0
	}
	if lookup != nil && !lookup(x.InterpID) {
		_ = x.Clear(0)
		return -1
	}
	if x.InterpID == currentID {
		_ = x.Clear(0)
		return 0
	}
	return CallInInterpreter(currentID, x.InterpID, func(any) int {
		_ = x.Clear(0)
		return 0
	}, x, schedule)
}

func ReleaseAndRawFreeXIData(x *XIData, currentID int64, lookup func(int64) bool, schedule func(int64, func(any) int, any)) int {
	return ReleaseXIData(x, currentID, lookup, schedule)
}

func (x *XIData) Init(interpID int64, shared any, obj any, newObject func(any) (any, error)) {
	*x = XIData{
		InterpID:  interpID,
		Data:      shared,
		Obj:       obj,
		NewObject: newObject,
	}
}

func NewObjectFromXIData(x *XIData) (any, error) {
	if x == nil || x.NewObject == nil {
		return nil, fmt.Errorf("missing new_object func")
	}
	return x.NewObject(x)
}

func (x *XIData) InitWithSize(interpID int64, size int, obj any, newObject func(any) (any, error)) {
	x.Init(interpID, make([]byte, size), obj, newObject)
	x.Free = func(any) {}
}

func (x *XIData) Clear(interpID int64) error {
	if x.InterpID != -1 && interpID != 0 && x.InterpID != interpID {
		return fmt.Errorf("wrong owning interpreter")
	}
	if x.Free != nil && x.Data != nil {
		x.Free(x.Data)
	}
	x.Data = nil
	x.Obj = nil
	x.NewObject = nil
	x.Free = nil
	x.InterpID = -1
	return nil
}

func CheckXIData(x *XIData) error {
	if x.InterpID < 0 {
		return fmt.Errorf("missing interp")
	}
	if x.NewObject == nil {
		return fmt.Errorf("missing new_object func")
	}
	return nil
}

func CallInInterpreter(currentID, targetID int64, fn func(any) int, arg any, schedule func(int64, func(any) int, any)) int {
	if currentID == targetID {
		return fn(arg)
	}
	if schedule != nil {
		schedule(targetID, fn, arg)
	}
	return 0
}

func CallInInterpreterAndRawFree(currentID, targetID int64, fn func(any) int, arg any, schedule func(int64, func(any) int, any)) int {
	if currentID == targetID {
		return fn(arg)
	}
	if schedule != nil {
		schedule(targetID, fn, arg)
	}
	return 0
}

type XIDataFallback int

const (
	XIDataOnly XIDataFallback = iota + 1
	XIDataFullFallback
)

func SetXIDataLookupFailure(obj any, msg string, cause error) error {
	if msg != "" {
		return &NotShareableError{Message: msg, Cause: cause}
	}
	if obj == nil {
		return &NotShareableError{Message: "object does not support cross-interpreter data", Cause: cause}
	}
	return &NotShareableError{
		Message: fmt.Sprintf("%v does not support cross-interpreter data", obj),
		Cause:   cause,
	}
}

func ObjectCheckXIData(lookup *LookupState, obj any) error {
	if lookup == nil {
		return SetXIDataLookupFailure(obj, "", nil)
	}
	getdata := lookup.Lookup(obj)
	if getdata == nil {
		return SetXIDataLookupFailure(obj, "", nil)
	}
	return nil
}

func ObjectGetXIDataNoFallback(lookup *LookupState, interpID int64, obj any, xidata *XIData) error {
	return ObjectGetXIData(lookup, interpID, obj, XIDataOnly, xidata)
}

func ObjectGetXIData(lookup *LookupState, interpID int64, obj any, fallback XIDataFallback, xidata *XIData) error {
	if xidata == nil {
		return fmt.Errorf("missing xidata")
	}
	if xidata.Data != nil || xidata.Obj != nil {
		return fmt.Errorf("xidata not cleared")
	}
	if lookup != nil {
		if getdata := lookup.Lookup(obj); getdata != nil {
			shared, err := getdata(obj)
			if err != nil {
				return SetXIDataLookupFailure(obj, "", err)
			}
			xidata.Init(interpID, shared, obj, func(v any) (any, error) {
				return v.(*XIData).Data, nil
			})
			return CheckXIData(xidata)
		}
	}
	switch fallback {
	case XIDataOnly:
		return SetXIDataLookupFailure(obj, "", nil)
	case XIDataFullFallback:
		if err := PickleToXIData(obj, "", xidata); err == nil {
			xidata.InterpID = interpID
			return CheckXIData(xidata)
		}
		return SetXIDataLookupFailure(obj, "", nil)
	default:
		return fmt.Errorf("unsupported xidata fallback option")
	}
}

func CopyStringObjectRaw(text string) (string, int, error) {
	for _, ch := range text {
		if ch == 0 {
			return "", 0, fmt.Errorf("found embedded NULL character")
		}
	}
	return text, len(text), nil
}
