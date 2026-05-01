package pycrossinterp

import (
	"reflect"
	"sync"
)

type GetDataFunc func(any) (any, error)

type RegistryItem struct {
	Type    reflect.Type
	GetData GetDataFunc
}

type Registry struct {
	mu          sync.Mutex
	initialized bool
	items       []RegistryItem
}

type LookupState struct {
	Global Registry
	Local  Registry
}

func (registry *Registry) Init() {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if registry.initialized {
		return
	}
	registry.initialized = true
}

func (registry *Registry) Fini() {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.initialized = false
	registry.items = nil
}

func (registry *Registry) Add(sample any, getdata GetDataFunc) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.items = append(registry.items, RegistryItem{
		Type:    reflect.TypeOf(sample),
		GetData: getdata,
	})
}

func (registry *Registry) Lookup(obj any) GetDataFunc {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	objType := reflect.TypeOf(obj)
	for _, item := range registry.items {
		if item.Type == objType {
			return item.GetData
		}
	}
	return nil
}

func (lookup *LookupState) Init() {
	lookup.Global.Init()
	lookup.Local.Init()
}

func (lookup *LookupState) Fini() {
	lookup.Global.Fini()
	lookup.Local.Fini()
}

func (lookup *LookupState) Lookup(obj any) GetDataFunc {
	if fn := lookup.Local.Lookup(obj); fn != nil {
		return fn
	}
	return lookup.Global.Lookup(obj)
}
