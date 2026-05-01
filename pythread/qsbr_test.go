package pythread

import "testing"

func TestQSBRReserveRegisterAndGrow(t *testing.T) {
	interp := NewInterpreterState()

	indices := make([]int, 0, 9)
	threads := make([]*ThreadState, 0, 9)
	for range 9 {
		index := interp.QSBR.Reserve()
		if index < 0 {
			t.Fatal("reserve failed")
		}
		ts := NewThreadState(interp)
		interp.QSBR.Register(ts, index)
		indices = append(indices, index)
		threads = append(threads, ts)
	}

	if len(interp.QSBR.array) != 16 {
		t.Fatalf("array size = %d, want 16", len(interp.QSBR.array))
	}
	for i, ts := range threads {
		if ts.QSBR() == nil {
			t.Fatalf("thread %d did not receive a qsbr state", i)
		}
		if got := interp.QSBR.indexOf(ts.QSBR()); got != indices[i] {
			t.Fatalf("thread %d index = %d, want %d", i, got, indices[i])
		}
	}
}

func TestQSBRAdvancePollDetachAndUnregister(t *testing.T) {
	interp := NewInterpreterState()
	ts1 := NewThreadState(interp)
	ts2 := NewThreadState(interp)

	i1 := interp.QSBR.Reserve()
	i2 := interp.QSBR.Reserve()
	interp.QSBR.Register(ts1, i1)
	interp.QSBR.Register(ts2, i2)

	interp.QSBR.Attach(ts1.QSBR())
	interp.QSBR.Attach(ts2.QSBR())

	if got, want := ts1.QSBR().seq, QSBRInitial; got != want {
		t.Fatalf("initial seq = %d, want %d", got, want)
	}

	goal := interp.QSBR.Advance()
	if got, want := goal, uint64(3); got != want {
		t.Fatalf("goal = %d, want %d", got, want)
	}
	if interp.QSBR.Poll(ts1.QSBR(), goal) {
		t.Fatal("goal reached before quiescent states")
	}

	interp.QSBR.QuiescentState(ts1.QSBR())
	if interp.QSBR.Poll(ts1.QSBR(), goal) {
		t.Fatal("goal reached before second thread reported quiescent state")
	}

	interp.QSBR.QuiescentState(ts2.QSBR())
	if !interp.QSBR.Poll(ts1.QSBR(), goal) {
		t.Fatal("goal not reached after all threads reported quiescent state")
	}

	interp.QSBR.Detach(ts1.QSBR())
	interp.QSBR.Unregister(ts1)
	if ts1.QSBR() != nil {
		t.Fatal("unregister should clear thread qsbr pointer")
	}
}

func TestQSBRAfterForkKeepsCurrentThread(t *testing.T) {
	interp := NewInterpreterState()
	ts1 := NewThreadState(interp)
	ts2 := NewThreadState(interp)

	i1 := interp.QSBR.Reserve()
	i2 := interp.QSBR.Reserve()
	interp.QSBR.Register(ts1, i1)
	interp.QSBR.Register(ts2, i2)

	interp.QSBR.AfterFork(ts1)

	if !ts1.QSBR().Allocated {
		t.Fatal("current thread qsbr should remain allocated")
	}
	if ts2.QSBR() == nil {
		t.Fatal("other thread pointer should still be addressable for inspection")
	}
	if ts2.QSBR().Allocated {
		t.Fatal("other thread qsbr should be released after fork")
	}
}
