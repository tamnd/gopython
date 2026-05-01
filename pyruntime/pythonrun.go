package pyruntime

import "fmt"

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
