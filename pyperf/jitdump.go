package pyperf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"runtime"
	"time"
)

const (
	JitDumpMagic   = 0x4A695444
	JitDumpVersion = 1
)

type PerfEvent uint32

const (
	PerfLoad PerfEvent = iota
	PerfMove
	PerfDebugInfo
	PerfClose
	PerfUnwindingInfo
)

type Header struct {
	Magic         uint32
	Version       uint32
	Size          uint32
	ElfMachTarget uint32
	Reserved      uint32
	ProcessID     uint32
	TimeStamp     uint64
	Flags         uint64
}

type BaseEvent struct {
	Event     uint32
	Size      uint32
	TimeStamp uint64
}

type CodeLoadEvent struct {
	Base      BaseEvent
	ProcessID uint32
	ThreadID  uint32
	VMA       uint64
	CodeAddr  uint64
	CodeSize  uint64
	CodeID    uint64
	Name      string
	CodeBytes []byte
}

type CodeUnwindingInfoEvent struct {
	Base           BaseEvent
	UnwindDataSize uint64
	EHFrameHdrSize uint64
	MappedSize     uint64
	EHFrameHeader  []byte
	UnwindData     []byte
	Padding        []byte
}

func elfMachineArchitecture() uint32 {
	switch runtime.GOARCH {
	case "amd64":
		return 62
	case "386":
		return 3
	case "arm64":
		return 183
	case "arm":
		return 40
	case "riscv64":
		return 243
	default:
		return 0
	}
}

func CurrentMonotonicTicks() int64 {
	return time.Since(time.Unix(0, 0)).Nanoseconds()
}

func CurrentTimeMicroseconds() int64 {
	return time.Now().UnixMicro()
}

func NewHeader(pid uint32) Header {
	return Header{
		Magic:         JitDumpMagic,
		Version:       JitDumpVersion,
		Size:          uint32(binary.Size(Header{})),
		ElfMachTarget: elfMachineArchitecture(),
		ProcessID:     pid,
		TimeStamp:     uint64(CurrentTimeMicroseconds()),
	}
}

func WriteHeader(w io.Writer, pid uint32) error {
	header := NewHeader(pid)
	return binary.Write(w, binary.LittleEndian, header)
}

func NewCodeLoadEvent(pid, tid uint32, codeAddr uint64, codeID uint64, name string, code []byte) CodeLoadEvent {
	baseSize := binary.Size(BaseEvent{}) + 4 + 4 + 8 + 8 + 8 + 8
	totalSize := uint32(baseSize + len(name) + 1 + len(code))
	return CodeLoadEvent{
		Base: BaseEvent{
			Event:     uint32(PerfLoad),
			Size:      totalSize,
			TimeStamp: uint64(CurrentMonotonicTicks()),
		},
		ProcessID: pid,
		ThreadID:  tid,
		VMA:       codeAddr,
		CodeAddr:  codeAddr,
		CodeSize:  uint64(len(code)),
		CodeID:    codeID,
		Name:      name,
		CodeBytes: append([]byte(nil), code...),
	}
}

func (e CodeLoadEvent) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, e.Base); err != nil {
		return err
	}
	fields := []any{e.ProcessID, e.ThreadID, e.VMA, e.CodeAddr, e.CodeSize, e.CodeID}
	for _, field := range fields {
		if err := binary.Write(w, binary.LittleEndian, field); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, e.Name); err != nil {
		return err
	}
	if _, err := w.Write([]byte{0}); err != nil {
		return err
	}
	_, err := w.Write(e.CodeBytes)
	return err
}

func NewCodeUnwindingInfoEvent(ehFrameHeader, unwindData []byte, mappedSize uint64, paddingSize int) CodeUnwindingInfoEvent {
	padding := make([]byte, paddingSize)
	size := uint32(binary.Size(BaseEvent{}) + 8 + 8 + 8 + len(ehFrameHeader) + len(unwindData) + len(padding))
	return CodeUnwindingInfoEvent{
		Base: BaseEvent{
			Event:     uint32(PerfUnwindingInfo),
			Size:      size,
			TimeStamp: uint64(CurrentMonotonicTicks()),
		},
		UnwindDataSize: uint64(len(unwindData)),
		EHFrameHdrSize: uint64(len(ehFrameHeader)),
		MappedSize:     mappedSize,
		EHFrameHeader:  append([]byte(nil), ehFrameHeader...),
		UnwindData:     append([]byte(nil), unwindData...),
		Padding:        padding,
	}
}

func (e CodeUnwindingInfoEvent) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, e.Base); err != nil {
		return err
	}
	for _, field := range []any{e.UnwindDataSize, e.EHFrameHdrSize, e.MappedSize} {
		if err := binary.Write(w, binary.LittleEndian, field); err != nil {
			return err
		}
	}
	for _, chunk := range [][]byte{e.EHFrameHeader, e.UnwindData, e.Padding} {
		if _, err := w.Write(chunk); err != nil {
			return err
		}
	}
	return nil
}

func NewCloseEvent() BaseEvent {
	return BaseEvent{
		Event:     uint32(PerfClose),
		Size:      uint32(binary.Size(BaseEvent{})),
		TimeStamp: uint64(CurrentMonotonicTicks()),
	}
}

func WriteCloseEvent(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, NewCloseEvent())
}

func SerializeJitDump(pid, tid uint32, codeAddr uint64, codeID uint64, name string, code []byte, ehFrameHeader []byte, unwindData []byte, paddingSize int) ([]byte, error) {
	var buf bytes.Buffer
	if err := WriteHeader(&buf, pid); err != nil {
		return nil, err
	}
	load := NewCodeLoadEvent(pid, tid, codeAddr, codeID, name, code)
	if err := load.Write(&buf); err != nil {
		return nil, err
	}
	unwind := NewCodeUnwindingInfoEvent(ehFrameHeader, unwindData, uint64(len(code)+paddingSize), paddingSize)
	if err := unwind.Write(&buf); err != nil {
		return nil, err
	}
	if err := WriteCloseEvent(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MapEntry(codeAddr uint64, codeSize uint32, co CodeObject) string {
	return fmt.Sprintf("%x %x %s", codeAddr, codeSize, PerfMapSymbol(co))
}
