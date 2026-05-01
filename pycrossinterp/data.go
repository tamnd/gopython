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

func (x *XIData) Init(interpID int64, shared any, obj any, newObject func(any) (any, error)) {
	*x = XIData{
		InterpID:  interpID,
		Data:      shared,
		Obj:       obj,
		NewObject: newObject,
	}
}

func (x *XIData) InitWithSize(interpID int64, size int, obj any, newObject func(any) (any, error)) {
	x.Init(interpID, make([]byte, size), obj, newObject)
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
