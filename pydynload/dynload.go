package pydynload

import (
	"errors"
	"strings"

	"golang.org/x/net/idna"
)

var ErrUnsupported = errors.New("dynamic loading is unsupported")

const (
	asciiOnlyPrefix = "PyInit"
	nonASCIIprefix  = "PyInitU"
)

type ModuleOrigin int

const (
	OriginBuiltin ModuleOrigin = iota + 1
	OriginCore
	OriginDynamic
)

type LoaderInfo struct {
	Name        string
	NameEncoded string
	Filename    string
	Path        string
	Origin      ModuleOrigin
	HookPrefix  string
	NewContext  string
}

func DynLoadFiletab() []string {
	return dynLoadFiletab()
}

func FindSharedFuncptr(prefix, shortname, pathname string) (uintptr, error) {
	return findSharedFuncptr(prefix, shortname, pathname)
}

func EncodedName(name string) (encoded string, hookPrefix string, err error) {
	shortname := name
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		shortname = name[i+1:]
	}
	if isASCII(shortname) {
		return strings.ReplaceAll(shortname, "-", "_"), asciiOnlyPrefix, nil
	}

	encoded, err = idna.ToASCII(shortname)
	if err != nil {
		return "", "", err
	}
	return strings.ReplaceAll(encoded, "-", "_"), nonASCIIprefix, nil
}

func InitLoaderInfo(name, filename string, origin ModuleOrigin) (LoaderInfo, error) {
	encoded, hookPrefix, err := EncodedName(name)
	if err != nil {
		return LoaderInfo{}, err
	}

	path := name
	if filename != "" {
		path = filename
	}
	return LoaderInfo{
		Name:        name,
		NameEncoded: encoded,
		Filename:    filename,
		Path:        path,
		Origin:      origin,
		HookPrefix:  hookPrefix,
		NewContext:  name,
	}, nil
}

func InitLoaderInfoForBuiltin(name string) (LoaderInfo, error) {
	return InitLoaderInfo(name, "", OriginBuiltin)
}

func InitLoaderInfoForCore(name string) (LoaderInfo, error) {
	info, err := InitLoaderInfoForBuiltin(name)
	if err != nil {
		return LoaderInfo{}, err
	}
	info.Origin = OriginCore
	return info, nil
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 0x7f {
			return false
		}
	}
	return true
}
