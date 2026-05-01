package pyfrozen

import "testing"

func TestBootstrapRegistryShape(t *testing.T) {
	modules := BootstrapModules()
	if len(modules) != 3 {
		t.Fatalf("bootstrap module count = %d, want 3", len(modules))
	}
	if modules[0].Name != "_frozen_importlib" {
		t.Fatalf("bootstrap[0] = %q", modules[0].Name)
	}
	if modules[1].Name != "_frozen_importlib_external" {
		t.Fatalf("bootstrap[1] = %q", modules[1].Name)
	}
	if modules[2].Name != "zipimport" {
		t.Fatalf("bootstrap[2] = %q", modules[2].Name)
	}
}

func TestStdlibRegistryContainsStartupModules(t *testing.T) {
	for _, name := range []string{
		"abc",
		"codecs",
		"io",
		"_collections_abc",
		"_sitebuiltins",
		"genericpath",
		"ntpath",
		"posixpath",
		"os",
		"site",
		"stat",
		"importlib.util",
		"importlib.machinery",
		"runpy",
	} {
		if _, ok := LookupStdlib(name); !ok {
			t.Fatalf("missing stdlib frozen module %q", name)
		}
	}
}

func TestTestRegistryPackageFlags(t *testing.T) {
	cases := map[string]bool{
		"__hello__":               false,
		"__phello_alias__":        true,
		"__phello__":              true,
		"__phello__.ham":          true,
		"__phello__.ham.eggs":     false,
		"__phello__.spam":         false,
		"__hello_only__":          false,
		"__phello_alias__.spam":   false,
		"__phello__.__init__":     false,
		"__phello__.ham.__init__": false,
	}
	for name, wantPackage := range cases {
		module, ok := LookupTest(name)
		if !ok {
			t.Fatalf("missing test frozen module %q", name)
		}
		if module.IsPackage != wantPackage {
			t.Fatalf("%q package flag = %t, want %t", name, module.IsPackage, wantPackage)
		}
	}
}

func TestAliasTableMatchesFrozenC(t *testing.T) {
	cases := map[string]string{
		"_frozen_importlib":          "importlib._bootstrap",
		"_frozen_importlib_external": "importlib._bootstrap_external",
		"__hello_alias__":            "__hello__",
		"__phello_alias__":           "__hello__",
		"__phello_alias__.spam":      "__hello__",
		"__phello__.__init__":        "<__phello__",
		"__phello__.ham.__init__":    "<__phello__.ham",
		"__hello_only__":             "",
	}
	for name, want := range cases {
		alias, ok := LookupAlias(name)
		if !ok {
			t.Fatalf("missing alias %q", name)
		}
		if alias.Target != want {
			t.Fatalf("%q target = %q, want %q", name, alias.Target, want)
		}
	}
}

func TestRegistrySlicesAreCloned(t *testing.T) {
	modules := BootstrapModules()
	modules[0].Name = "changed"
	if got, _ := LookupBootstrap("_frozen_importlib"); got.Name != "_frozen_importlib" {
		t.Fatalf("bootstrap registry mutated to %q", got.Name)
	}

	aliasList := Aliases()
	aliasList[0].Target = "changed"
	alias, _ := LookupAlias("_frozen_importlib")
	if alias.Target != "importlib._bootstrap" {
		t.Fatalf("alias registry mutated to %q", alias.Target)
	}
}

func TestEmbeddingOverrideCopiesInput(t *testing.T) {
	SetFrozenModules([]Module{{Name: "demo", IsPackage: true}})
	if len(FrozenModules) != 1 || FrozenModules[0].Name != "demo" || !FrozenModules[0].IsPackage {
		t.Fatalf("FrozenModules = %#v", FrozenModules)
	}

	override := []Module{{Name: "changed"}}
	SetFrozenModules(override)
	override[0].Name = "mutated"
	if FrozenModules[0].Name != "changed" {
		t.Fatalf("FrozenModules mutated through caller slice: %#v", FrozenModules)
	}

	SetFrozenModules(nil)
	if FrozenModules != nil {
		t.Fatalf("FrozenModules reset = %#v, want nil", FrozenModules)
	}
}
