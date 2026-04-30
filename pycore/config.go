package pycore

import (
	"fmt"
)

const PythonDirectoryReadme = "Miscellaneous source files for the main Python shared library.\n"

func ConfigDictGet(dict map[string]any, name string) (any, error) {
	item, ok := dict[name]
	if !ok {
		return nil, fmt.Errorf("missing config key: %s", name)
	}
	return item, nil
}

func ConfigDictInvalidType(name string) error {
	return fmt.Errorf("invalid config type: %s", name)
}
