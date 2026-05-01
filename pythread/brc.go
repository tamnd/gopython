package pythread

import "github.com/tamnd/gopython/pycore"

const BRCNumBuckets = 257

type BRCObject interface {
	OwnerThreadID() uintptr
	ExplicitMerge(delta int) int
	Dealloc()
}

type brcBucket struct {
	mutex   Mutex
	threads []*ThreadState
}

type BRCState struct {
	table [BRCNumBuckets]brcBucket
}

type BRCThreadState struct {
	tid                 uintptr
	objectsToMerge      pycore.ObjectStack[BRCObject]
	localObjectsToMerge pycore.ObjectStack[BRCObject]
}

func (s *BRCState) bucket(threadID uintptr) *brcBucket {
	return &s.table[threadID%BRCNumBuckets]
}

func (s *BRCState) findThreadState(bucket *brcBucket, threadID uintptr) *ThreadState {
	for _, ts := range bucket.threads {
		if ts.brc.tid == threadID {
			return ts
		}
	}
	return nil
}

func (interp *InterpreterState) QueueBRCObject(ob BRCObject) {
	obTid := ob.OwnerThreadID()
	if obTid == 0 {
		if ob.ExplicitMerge(-1) == 0 {
			ob.Dealloc()
		}
		return
	}

	bucket := interp.BRC.bucket(obTid)
	bucket.mutex.Lock()
	tstate := interp.BRC.findThreadState(bucket, obTid)
	if tstate == nil {
		bucket.mutex.Unlock()
		if ob.ExplicitMerge(-1) == 0 {
			ob.Dealloc()
		}
		return
	}

	tstate.brc.objectsToMerge.Push(ob)
	tstate.ExplicitMergePending = true
	bucket.mutex.Unlock()
}

func (t *ThreadState) InitBRCThread(threadID uintptr) {
	bucket := t.Interp.BRC.bucket(threadID)
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	t.ThreadID = threadID
	t.brc.tid = threadID
	bucket.threads = append(bucket.threads, t)
}

func (t *ThreadState) MergeQueuedRefcounts() {
	bucket := t.Interp.BRC.bucket(t.brc.tid)
	bucket.mutex.Lock()
	t.brc.localObjectsToMerge.Merge(&t.brc.objectsToMerge)
	t.ExplicitMergePending = false
	bucket.mutex.Unlock()

	mergeQueuedObjects(&t.brc.localObjectsToMerge)
}

func (t *ThreadState) RemoveBRCThread() {
	if t.brc.tid == 0 {
		return
	}

	bucket := t.Interp.BRC.bucket(t.brc.tid)
	empty := false
	for !empty {
		mergeQueuedObjects(&t.brc.localObjectsToMerge)

		bucket.mutex.Lock()
		empty = t.brc.objectsToMerge.Size() == 0
		if empty {
			filtered := bucket.threads[:0]
			for _, ts := range bucket.threads {
				if ts != t {
					filtered = append(filtered, ts)
				}
			}
			bucket.threads = filtered
		} else {
			t.brc.localObjectsToMerge.Merge(&t.brc.objectsToMerge)
		}
		bucket.mutex.Unlock()
	}
	t.brc.tid = 0
	t.ThreadID = 0
}

func (interp *InterpreterState) BRCAfterFork() {
	for i := range interp.BRC.table {
		interp.BRC.table[i].mutex = Mutex{}
	}
}

func mergeQueuedObjects(toMerge *pycore.ObjectStack[BRCObject]) {
	for {
		ob, ok := toMerge.Pop()
		if !ok {
			return
		}
		if ob.ExplicitMerge(-1) == 0 {
			ob.Dealloc()
		}
	}
}
