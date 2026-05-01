package pycrossinterp

import "fmt"

type SharedNamespaceItem struct {
	Name   string
	XIData *XIData
}

type SharedNamespace struct {
	MaxItems  int
	NumNames  int
	NumValues int
	Items     []SharedNamespaceItem
}

func AllocSharedNamespace(maxitems int) (*SharedNamespace, error) {
	if maxitems < 0 {
		return nil, fmt.Errorf("bad internal call")
	}
	if maxitems == 0 {
		return nil, fmt.Errorf("empty namespaces not allowed")
	}
	return &SharedNamespace{
		MaxItems: maxitems,
		Items:    make([]SharedNamespaceItem, maxitems),
	}, nil
}

func CreateSharedNamespace(names []string) (*SharedNamespace, error) {
	ns, err := AllocSharedNamespace(len(names))
	if err != nil {
		return nil, err
	}
	for i, name := range names {
		ns.Items[i].Name = name
		ns.NumNames++
	}
	return ns, nil
}

func (ns *SharedNamespace) Fill(values map[string]any, fallback XIDataFallback) error {
	if ns == nil {
		return fmt.Errorf("missing shared namespace")
	}
	for i := 0; i < ns.MaxItems; i++ {
		item := &ns.Items[i]
		value, ok := values[item.Name]
		if !ok {
			ns.NumValues++
			continue
		}
		item.XIData = NewXIData()
		if err := ObjectGetXIData(nil, 1, value, fallback, item.XIData); err != nil {
			for j := 0; j < i; j++ {
				ns.Items[j].ClearValue()
			}
			ns.NumValues = 0
			return err
		}
		ns.NumValues++
	}
	return nil
}

func (item *SharedNamespaceItem) ClearValue() {
	if item.XIData != nil {
		_ = FreeXIData(0, item.XIData)
		item.XIData = nil
	}
}

func (item *SharedNamespaceItem) Apply(target map[string]any, dflt any) error {
	if target == nil {
		return fmt.Errorf("missing target namespace")
	}
	if item.XIData != nil {
		value, err := item.XIData.NewObject(item.XIData)
		if err != nil {
			return err
		}
		target[item.Name] = value
		return nil
	}
	target[item.Name] = dflt
	return nil
}

func (ns *SharedNamespace) Apply(target map[string]any, dflt any) error {
	if ns == nil {
		return fmt.Errorf("missing shared namespace")
	}
	for i := 0; i < ns.MaxItems; i++ {
		if err := ns.Items[i].Apply(target, dflt); err != nil {
			return err
		}
	}
	return nil
}

func (ns *SharedNamespace) Destroy(currentInterpID int64, lookup func(int64) bool, schedule func(int64, func(any) int, any)) {
	if ns == nil {
		return
	}
	if ns.NumValues == 0 {
		ns.Free()
		return
	}
	interpID := int64(-1)
	for i := 0; i < ns.MaxItems; i++ {
		if ns.Items[i].XIData != nil {
			interpID = ns.Items[i].XIData.InterpID
			break
		}
	}
	if interpID == -1 || interpID == currentInterpID || (lookup != nil && !lookup(interpID)) {
		ns.Free()
		return
	}
	if schedule != nil {
		CallInInterpreter(currentInterpID, interpID, func(any) int {
			ns.Free()
			return 0
		}, ns, schedule)
	}
}

func (ns *SharedNamespace) Free() {
	if ns == nil {
		return
	}
	for i := 0; i < ns.NumNames; i++ {
		ns.Items[i].Name = ""
		ns.Items[i].ClearValue()
	}
	ns.NumNames = 0
	ns.NumValues = 0
}
