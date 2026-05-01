package pycore

import "fmt"

const InitialStackRefIndex = uint64(8)

type StackRef struct {
	Index uint64
}

var (
	StackRefNull  = StackRef{Index: 0}
	StackRefNone  = StackRef{Index: 2}
	StackRefFalse = StackRef{Index: 4}
	StackRefTrue  = StackRef{Index: 6}
)

type StackRefRegistry struct {
	next       uint64
	open       map[uint64]*stackRefEntry
	closed     map[uint64]*stackRefEntry
	closeDebug bool
}

type stackRefEntry struct {
	obj              any
	className        string
	filename         string
	lineNumber       int
	borrowFilename   string
	borrowLineNumber int
}

func NewStackRefRegistry(closeDebug bool) *StackRefRegistry {
	return &StackRefRegistry{
		next:       InitialStackRefIndex,
		open:       make(map[uint64]*stackRefEntry),
		closed:     make(map[uint64]*stackRefEntry),
		closeDebug: closeDebug,
	}
}

func (r *StackRefRegistry) GetObject(ref StackRef) any {
	if ref.Index == 0 {
		return nil
	}
	if ref.Index >= r.next {
		panic(fmt.Sprintf("Garbled stack ref with ID %d", ref.Index))
	}
	entry := r.open[ref.Index]
	if entry == nil {
		panic(fmt.Sprintf("Accessing closed stack ref with ID %d", ref.Index))
	}
	return entry.obj
}

func (r *StackRefRegistry) Is(a, b StackRef) bool {
	return r.GetObject(a) == r.GetObject(b)
}

func (r *StackRefRegistry) Close(ref StackRef, filename string, lineNumber int) any {
	if ref.Index >= r.next {
		panic(fmt.Sprintf("Invalid StackRef with ID %d at %s:%d", ref.Index, filename, lineNumber))
	}
	if ref.Index < InitialStackRefIndex {
		if ref.Index == 0 {
			panic(fmt.Sprintf("Passing NULL to StackRef.Close at %s:%d", filename, lineNumber))
		}
		entry := r.open[ref.Index]
		if entry == nil {
			panic(fmt.Sprintf("Invalid StackRef with ID %d", ref.Index))
		}
		return entry.obj
	}

	entry := r.open[ref.Index]
	if entry == nil {
		if r.closeDebug {
			if closed := r.closed[ref.Index]; closed != nil {
				panic(fmt.Sprintf(
					"Double close of ref ID %d at %s:%d. Referred to instance of %s. Closed at %s:%d",
					ref.Index,
					filename,
					lineNumber,
					closed.className,
					closed.filename,
					closed.lineNumber,
				))
			}
		}
		panic(fmt.Sprintf("Invalid StackRef with ID %d", ref.Index))
	}

	delete(r.open, ref.Index)
	if r.closeDebug {
		r.closed[ref.Index] = &stackRefEntry{
			obj:        entry.obj,
			className:  entry.className,
			filename:   filename,
			lineNumber: lineNumber,
		}
	}
	return entry.obj
}

func (r *StackRefRegistry) Create(obj any, className, filename string, lineNumber int) StackRef {
	if obj == nil {
		panic("cannot create a stackref for nil")
	}
	newID := r.next
	r.next = newID + 2
	r.open[newID] = &stackRefEntry{
		obj:        obj,
		className:  className,
		filename:   filename,
		lineNumber: lineNumber,
	}
	return StackRef{Index: newID}
}

func (r *StackRefRegistry) RecordBorrow(ref StackRef, filename string, lineNumber int) {
	if ref.Index < InitialStackRefIndex {
		return
	}
	entry := r.open[ref.Index]
	if entry == nil {
		if r.closeDebug {
			if closed := r.closed[ref.Index]; closed != nil {
				panic(fmt.Sprintf(
					"Borrow of closed ref ID %d at %s:%d. Referred to instance of %s. Closed at %s:%d",
					ref.Index,
					filename,
					lineNumber,
					closed.className,
					closed.filename,
					closed.lineNumber,
				))
			}
		}
		panic(fmt.Sprintf("Invalid StackRef with ID %d at %s:%d", ref.Index, filename, lineNumber))
	}
	entry.borrowFilename = filename
	entry.borrowLineNumber = lineNumber
}

func (r *StackRefRegistry) AssociateBuiltin(obj any, className string, ref StackRef) {
	if ref.Index >= InitialStackRefIndex {
		panic("builtin stackrefs must use reserved indices")
	}
	r.open[ref.Index] = &stackRefEntry{
		obj:        obj,
		className:  className,
		filename:   "builtin-object",
		lineNumber: 0,
	}
}

func (r *StackRefRegistry) ReportLeaks(isStaticImmortal func(any) bool) error {
	for _, entry := range r.open {
		if !isStaticImmortal(entry.obj) {
			msg := fmt.Sprintf(
				"Stackref leak. Refers to instance of %s. Created at %s:%d",
				entry.className,
				entry.filename,
				entry.lineNumber,
			)
			if entry.borrowFilename != "" {
				msg += fmt.Sprintf(". Last borrow at %s:%d", entry.borrowFilename, entry.borrowLineNumber)
			}
			return fmt.Errorf("%s", msg)
		}
	}
	return nil
}

func (r *StackRefRegistry) CloseSpecialized(ref StackRef, destruct func(any), filename string, lineNumber int) {
	obj := r.Close(ref, filename, lineNumber)
	destruct(obj)
}

func TagInt(i int64) StackRef {
	return StackRef{Index: uint64(i<<1) + 1}
}

func UntagInt(ref StackRef) int64 {
	if !IsTaggedInt(ref) {
		panic("stackref is not a tagged int")
	}
	return int64(ref.Index) >> 1
}

func IsTaggedInt(ref StackRef) bool {
	return ref.Index&1 == 1
}

func IsNull(ref StackRef) bool {
	return ref.Index == 0
}

func IsNullOrInt(ref StackRef) bool {
	return IsNull(ref) || IsTaggedInt(ref)
}
