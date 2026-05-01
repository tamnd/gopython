package pyperf

import (
	"fmt"
)

const (
	CodeAlignment    = 32
	DefaultArenaSize = 4096 * 16
)

type CodeObject struct {
	QualName string
	Filename string
}

type TrampolineType int

const (
	TrampolineUnset TrampolineType = iota
	TrampolineMap
	TrampolineJitDump
)

type Callbacks struct {
	InitState  func() any
	WriteState func(state any, codeAddr uint64, codeSize uint32, co CodeObject)
	FreeState  func(state any) error
}

type CodeArena struct {
	Memory      []byte
	Current     int
	SizeLeft    int
	CodeSize    int
	ChunkSize   int
	Prev        *CodeArena
	BaseAddress uint64
}

type TrampolineManager struct {
	ArenaSize    int
	CodeTemplate []byte
	CodePadding  int
	Type         TrampolineType
	StatusOK     bool
	Callbacks    Callbacks
	State        any
	Arena        *CodeArena
	NextCodeAddr uint64
	RefCount     int
}

func RoundUp(value, multiple int) int {
	if multiple == 0 {
		return value
	}
	remainder := value % multiple
	if remainder == 0 {
		return value
	}
	return value + multiple - remainder
}

func PerfMapSymbol(co CodeObject) string {
	return fmt.Sprintf("py::%s:%s", co.QualName, co.Filename)
}

func NewTrampolineManager(template []byte, padding int, callbacks Callbacks) *TrampolineManager {
	return &TrampolineManager{
		ArenaSize:    DefaultArenaSize,
		CodeTemplate: append([]byte(nil), template...),
		CodePadding:  padding,
		Callbacks:    callbacks,
		StatusOK:     true,
		NextCodeAddr: 0x10000000,
	}
}

func (m *TrampolineManager) Init(tt TrampolineType) error {
	if m.Callbacks.InitState != nil {
		m.State = m.Callbacks.InitState()
	}
	m.Type = tt
	m.RefCount = 1
	return m.newCodeArena()
}

func (m *TrampolineManager) newCodeArena() error {
	memSize := m.ArenaSize
	codeSize := len(m.CodeTemplate)
	chunkSize := RoundUp(codeSize+m.CodePadding, CodeAlignment)
	if chunkSize == 0 || chunkSize > memSize {
		return fmt.Errorf("invalid trampoline chunk size")
	}
	memory := make([]byte, memSize)
	for i := 0; i+codeSize <= memSize; i += chunkSize {
		copy(memory[i:i+codeSize], m.CodeTemplate)
	}
	arena := &CodeArena{
		Memory:      memory,
		Current:     0,
		SizeLeft:    memSize,
		CodeSize:    codeSize,
		ChunkSize:   chunkSize,
		Prev:        m.Arena,
		BaseAddress: m.NextCodeAddr,
	}
	m.NextCodeAddr += uint64(memSize + chunkSize)
	m.Arena = arena
	return nil
}

func (m *TrampolineManager) CompileTrampoline(co CodeObject) (uint64, uint32, error) {
	if m.Arena == nil || m.Arena.SizeLeft <= m.Arena.ChunkSize {
		if err := m.newCodeArena(); err != nil {
			return 0, 0, err
		}
	}
	addr := m.Arena.BaseAddress + uint64(m.Arena.Current)
	codeSize := uint32(m.Arena.CodeSize)
	m.Arena.Current += m.Arena.ChunkSize
	m.Arena.SizeLeft -= m.Arena.ChunkSize
	if m.Callbacks.WriteState != nil {
		m.Callbacks.WriteState(m.State, addr, codeSize, co)
	}
	m.RefCount++
	return addr, codeSize, nil
}

func (m *TrampolineManager) Fini() error {
	if m.Callbacks.FreeState != nil && m.State != nil {
		if err := m.Callbacks.FreeState(m.State); err != nil {
			return err
		}
	}
	m.Type = TrampolineUnset
	m.State = nil
	m.Arena = nil
	m.RefCount = 0
	return nil
}
