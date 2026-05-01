package pyruntime

import (
	"fmt"
	"strings"
)

func AnyFileObject(filename string, interactive bool, runInteractive func(string) error, runSimple func(string) error) error {
	if filename == "" {
		filename = "???"
	}
	if interactive {
		if runInteractive == nil {
			return nil
		}
		return runInteractive(filename)
	}
	if runSimple == nil {
		return nil
	}
	return runSimple(filename)
}

func InteractiveLoop(defaultPS1, defaultPS2 string, prompts map[string]string, runOne func() (int, error)) (map[string]string, error) {
	if prompts == nil {
		prompts = map[string]string{}
	}
	if prompts["ps1"] == "" {
		prompts["ps1"] = defaultPS1
	}
	if prompts["ps2"] == "" {
		prompts["ps2"] = defaultPS2
	}
	nomemCount := 0
	for {
		ret, err := runOne()
		if err != nil {
			if err.Error() == "MemoryError" {
				nomemCount++
				if nomemCount > 16 {
					return prompts, fmt.Errorf("too many consecutive memory errors")
				}
				continue
			}
			nomemCount = 0
			continue
		}
		nomemCount = 0
		if ret == EOF {
			return prompts, nil
		}
	}
}

const EOF = 11

func SimpleString(command string, name string, runString func(string, string) error) error {
	if runString == nil {
		return nil
	}
	return runString(command, name)
}

func SimpleFile(filename string, closeIt bool, maybePyc bool, runPyc func(string) error, runSource func(string) error) error {
	if filename == "" {
		filename = "???"
	}
	if maybePyc {
		if runPyc == nil {
			return nil
		}
		return runPyc(filename)
	}
	if runSource == nil {
		return nil
	}
	return runSource(filename)
}

func HandleSystemExit(exceptionType string, inspect bool, code any) (bool, int, string) {
	if exceptionType == "KeyboardInterrupt" {
		return false, 0, ""
	}
	if inspect {
		return false, 0, ""
	}
	if exceptionType != "SystemExit" {
		return false, 0, ""
	}
	switch value := code.(type) {
	case int:
		return true, value, ""
	case int64:
		return true, int(value), ""
	case nil:
		return true, 0, ""
	case string:
		return true, 1, value
	default:
		return true, 1, fmt.Sprint(value)
	}
}

func ExceptionMessage(moduleName, qualName, message string) string {
	var prefix string
	if moduleName != "" && moduleName != "builtins" && moduleName != "__main__" {
		prefix = moduleName + "."
	}
	if message == "" {
		return prefix + qualName
	}
	return prefix + qualName + ": " + message
}

func RenderChainedExceptions(messages []string) string {
	return strings.Join(messages, "\n\n")
}
