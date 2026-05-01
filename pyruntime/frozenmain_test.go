package pyruntime

import (
	"bytes"
	"os"
	"testing"

	"github.com/tamnd/gopython/pyconfig"
)

func TestFrozenMainSuccess(t *testing.T) {
	t.Setenv("PYTHONINSPECT", "")
	var sequence []string
	code, err := FrozenMain([][]byte{[]byte("prog")}, FrozenMainHooks{
		SetBytesArgv: func(config *pyconfig.Config, argv [][]byte) error {
			sequence = append(sequence, "argv")
			return nil
		},
		Initialize: func(config *pyconfig.Config) error {
			sequence = append(sequence, "init")
			return nil
		},
		SetRunningMain: func() error {
			sequence = append(sequence, "set-running")
			return nil
		},
		ImportFrozenMain: func() (int, error) {
			sequence = append(sequence, "import")
			return 1, nil
		},
		ClearRunningMain: func() {
			sequence = append(sequence, "clear-running")
		},
		Finalize: func() error {
			sequence = append(sequence, "finalize")
			return nil
		},
	})
	if err != nil || code != 0 {
		t.Fatalf("FrozenMain = (%d, %v)", code, err)
	}
	if len(sequence) != 6 {
		t.Fatalf("sequence = %#v", sequence)
	}
}

func TestFrozenMainFrozenMissing(t *testing.T) {
	_, err := FrozenMain(nil, FrozenMainHooks{
		ImportFrozenMain: func() (int, error) { return 0, nil },
	})
	if err == nil {
		t.Fatal("FrozenMain should reject missing __main__")
	}
}

func TestFrozenMainInspectAndVerbose(t *testing.T) {
	t.Setenv("PYTHONINSPECT", "1")
	tmp, err := os.CreateTemp(t.TempDir(), "frozenmain-out")
	if err != nil {
		t.Fatalf("CreateTemp returned error: %v", err)
	}
	defer tmp.Close()
	ranStdin := false
	code, err := FrozenMain(nil, FrozenMainHooks{
		Initialize: func(config *pyconfig.Config) error {
			config.Verbose = 1
			return nil
		},
		ImportFrozenMain: func() (int, error) { return 1, nil },
		RunStdin: func() error {
			ranStdin = true
			return nil
		},
		Finalize: func() error { return nil },
		Version:  func() string { return "3.14" },
		Copyright: func() string {
			return "copyright"
		},
		Stdout:     tmp,
		StdinIsTTY: true,
	})
	if err != nil || code != 0 {
		t.Fatalf("FrozenMain = (%d, %v)", code, err)
	}
	if !ranStdin {
		t.Fatal("RunStdin should have run")
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("Seek returned error: %v", err)
	}
	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !bytes.Contains(data, []byte("Python 3.14")) {
		t.Fatalf("verbose output = %q", string(data))
	}
}

func TestFrozenMainFinalizeFailureUses120(t *testing.T) {
	code, err := FrozenMain(nil, FrozenMainHooks{
		ImportFrozenMain: func() (int, error) { return 1, nil },
		Finalize:         func() error { return os.ErrInvalid },
	})
	if err != nil || code != 120 {
		t.Fatalf("FrozenMain = (%d, %v)", code, err)
	}
}
