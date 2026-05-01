package pyfrozen

// Module mirrors one entry in CPython's struct _frozen tables.
type Module struct {
	Name      string
	Code      []byte
	IsPackage bool
}

// Alias mirrors one entry in CPython's struct _module_alias table.
// Target is empty when the alias intentionally has no import target.
type Alias struct {
	Name   string
	Target string
}

var bootstrapModules = []Module{
	{Name: "_frozen_importlib"},
	{Name: "_frozen_importlib_external"},
	{Name: "zipimport"},
}

var stdlibModules = []Module{
	{Name: "abc"},
	{Name: "codecs"},
	{Name: "io"},
	{Name: "_collections_abc"},
	{Name: "_sitebuiltins"},
	{Name: "genericpath"},
	{Name: "ntpath"},
	{Name: "posixpath"},
	{Name: "os"},
	{Name: "site"},
	{Name: "stat"},
	{Name: "importlib.util"},
	{Name: "importlib.machinery"},
	{Name: "runpy"},
}

var testModules = []Module{
	{Name: "__hello__"},
	{Name: "__hello_alias__"},
	{Name: "__phello_alias__", IsPackage: true},
	{Name: "__phello_alias__.spam"},
	{Name: "__phello__", IsPackage: true},
	{Name: "__phello__.__init__"},
	{Name: "__phello__.ham", IsPackage: true},
	{Name: "__phello__.ham.__init__"},
	{Name: "__phello__.ham.eggs"},
	{Name: "__phello__.spam"},
	{Name: "__hello_only__"},
}

var aliases = []Alias{
	{Name: "_frozen_importlib", Target: "importlib._bootstrap"},
	{Name: "_frozen_importlib_external", Target: "importlib._bootstrap_external"},
	{Name: "__hello_alias__", Target: "__hello__"},
	{Name: "__phello_alias__", Target: "__hello__"},
	{Name: "__phello_alias__.spam", Target: "__hello__"},
	{Name: "__phello__.__init__", Target: "<__phello__"},
	{Name: "__phello__.ham.__init__", Target: "<__phello__.ham"},
	{Name: "__hello_only__"},
}

// FrozenModules matches CPython's embedding override pointer.
// A nil slice means the built-in registry tables remain in effect.
var FrozenModules []Module

func BootstrapModules() []Module {
	return cloneModules(bootstrapModules)
}

func StdlibModules() []Module {
	return cloneModules(stdlibModules)
}

func TestModules() []Module {
	return cloneModules(testModules)
}

func Aliases() []Alias {
	return cloneAliases(aliases)
}

func SetFrozenModules(modules []Module) {
	FrozenModules = cloneModules(modules)
}

func LookupBootstrap(name string) (Module, bool) {
	return lookupModule(bootstrapModules, name)
}

func LookupStdlib(name string) (Module, bool) {
	return lookupModule(stdlibModules, name)
}

func LookupTest(name string) (Module, bool) {
	return lookupModule(testModules, name)
}

func LookupAlias(name string) (Alias, bool) {
	for _, alias := range aliases {
		if alias.Name == name {
			return alias, true
		}
	}
	return Alias{}, false
}

func cloneModules(src []Module) []Module {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Module, len(src))
	copy(dst, src)
	return dst
}

func cloneAliases(src []Alias) []Alias {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Alias, len(src))
	copy(dst, src)
	return dst
}

func lookupModule(src []Module, name string) (Module, bool) {
	for _, module := range src {
		if module.Name == name {
			return module, true
		}
	}
	return Module{}, false
}
