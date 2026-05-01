package pycrossinterp

import "fmt"

type SyncModuleResult struct {
	Module any
	Loaded any
	Failed error
}

type SyncModule struct {
	Filename string
	Cached   SyncModuleResult
}

type MainModuleHooks struct {
	GetMainModule  func() (any, error)
	LoadFromPath   func(filename string, modname string) (any, error)
	SetMainModule  func(module any) error
	CloneNamespace func(module any) (any, error)
}

func (m *SyncModule) Clear() {
	m.Filename = ""
	m.Cached = SyncModuleResult{}
}

func EnsureIsolatedMain(main *SyncModule, hooks MainModuleHooks) error {
	if main.Cached.Failed != nil || main.Cached.Loaded != nil {
		return nil
	}
	if main.Filename == "" {
		return fmt.Errorf("not implemented")
	}
	if hooks.GetMainModule == nil || hooks.LoadFromPath == nil {
		return fmt.Errorf("missing main module hooks")
	}
	mod, err := hooks.GetMainModule()
	if err != nil {
		return err
	}
	loaded, err := hooks.LoadFromPath(main.Filename, "<fake __main__>")
	if err != nil {
		main.Cached.Failed = err
		return err
	}
	main.Cached = SyncModuleResult{
		Module: mod,
		Loaded: loaded,
	}
	return nil
}

func ApplyIsolatedMain(main *SyncModule, hooks MainModuleHooks) error {
	if main.Cached.Failed != nil {
		return main.Cached.Failed
	}
	if main.Cached.Loaded == nil {
		return fmt.Errorf("missing isolated main")
	}
	if hooks.SetMainModule == nil {
		return fmt.Errorf("missing set main hook")
	}
	return hooks.SetMainModule(main.Cached.Loaded)
}

func RestoreMain(main *SyncModule, hooks MainModuleHooks) error {
	if main.Cached.Module == nil {
		return nil
	}
	if hooks.SetMainModule == nil {
		return fmt.Errorf("missing set main hook")
	}
	return hooks.SetMainModule(main.Cached.Module)
}

func CheckMissingMainAttr(err error) bool {
	if err == nil {
		return false
	}
	return len(err.Error()) >= 36 && err.Error()[:36] == "module '__main__' has no attribute '"
}
