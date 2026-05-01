package pyperf

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

func TestHeaderAndEvents(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteHeader(&buf, 123); err != nil {
		t.Fatalf("WriteHeader returned error: %v", err)
	}
	var header Header
	if err := binary.Read(bytes.NewReader(buf.Bytes()), binary.LittleEndian, &header); err != nil {
		t.Fatalf("Read header returned error: %v", err)
	}
	if header.Magic != JitDumpMagic || header.Version != JitDumpVersion || header.ProcessID != 123 {
		t.Fatalf("unexpected header: %+v", header)
	}
}

func TestCodeLoadEventAndMapEntry(t *testing.T) {
	evt := NewCodeLoadEvent(1, 2, 0x1000, 7, "py::fn:file.py", []byte{1, 2, 3})
	var buf bytes.Buffer
	if err := evt.Write(&buf); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("py::fn:file.py")) {
		t.Fatal("load event output missing function name")
	}
	entry := MapEntry(0x1000, 3, CodeObject{QualName: "fn", Filename: "file.py"})
	if !strings.Contains(entry, "py::fn:file.py") {
		t.Fatalf("unexpected map entry: %q", entry)
	}
}

func TestSerializeJitDump(t *testing.T) {
	data, err := SerializeJitDump(11, 22, 0x2000, 9, "py::fn:file.py", []byte{1, 2}, []byte{3, 4}, []byte{5, 6, 7}, 8)
	if err != nil {
		t.Fatalf("SerializeJitDump returned error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("SerializeJitDump returned empty output")
	}
	if !bytes.Contains(data, []byte("py::fn:file.py")) {
		t.Fatal("jitdump output missing symbol name")
	}
}
