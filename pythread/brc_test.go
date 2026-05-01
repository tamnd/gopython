package pythread

import "testing"

type fakeBRCObject struct {
	owner       uintptr
	refcount    int
	merged      []int
	deallocated bool
}

func (o *fakeBRCObject) OwnerThreadID() uintptr {
	return o.owner
}

func (o *fakeBRCObject) ExplicitMerge(delta int) int {
	o.merged = append(o.merged, delta)
	o.refcount += delta
	return o.refcount
}

func (o *fakeBRCObject) Dealloc() {
	o.deallocated = true
}

func TestBRCQueueObjectToOwningThread(t *testing.T) {
	interp := NewInterpreterState()
	owner := NewThreadState(interp)
	owner.InitBRCThread(11)

	obj := &fakeBRCObject{owner: 11, refcount: 2}
	interp.QueueBRCObject(obj)
	if !owner.ExplicitMergePending {
		t.Fatal("queueing should notify the owning thread")
	}

	owner.MergeQueuedRefcounts()
	if len(obj.merged) != 1 || obj.merged[0] != -1 {
		t.Fatalf("merged deltas = %v, want [-1]", obj.merged)
	}
	if obj.deallocated {
		t.Fatal("object should not deallocate while refcount remains positive")
	}
}

func TestBRCQueueObjectMergesDirectlyWhenOwnerMissing(t *testing.T) {
	interp := NewInterpreterState()
	obj := &fakeBRCObject{owner: 17, refcount: 1}

	interp.QueueBRCObject(obj)
	if len(obj.merged) != 1 || obj.merged[0] != -1 {
		t.Fatalf("merged deltas = %v, want [-1]", obj.merged)
	}
	if !obj.deallocated {
		t.Fatal("missing owner should force direct merge and deallocation")
	}
}

func TestBRCQueueObjectWithZeroOwnerMergesImmediately(t *testing.T) {
	interp := NewInterpreterState()
	obj := &fakeBRCObject{owner: 0, refcount: 1}

	interp.QueueBRCObject(obj)
	if !obj.deallocated {
		t.Fatal("zero-owner object should be merged immediately")
	}
}

func TestBRCRemoveThreadDrainsQueuedObjects(t *testing.T) {
	interp := NewInterpreterState()
	owner := NewThreadState(interp)
	owner.InitBRCThread(23)

	obj := &fakeBRCObject{owner: 23, refcount: 1}
	interp.QueueBRCObject(obj)
	owner.RemoveBRCThread()

	if !obj.deallocated {
		t.Fatal("thread removal should drain queued objects")
	}
	if owner.ThreadID != 0 {
		t.Fatal("thread removal should clear thread id")
	}
}
