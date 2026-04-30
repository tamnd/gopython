package pythread

import (
	"os"
	"sync"
	"testing"
)

func TestDynamicAnnotationCallsAreNoOps(t *testing.T) {
	var lock int
	AnnotateRWLockCreate("file", 1, &lock)
	AnnotateRWLockDestroy("file", 2, &lock)
	AnnotateRWLockAcquired("file", 3, &lock, 1)
	AnnotateRWLockReleased("file", 4, &lock, 0)
	AnnotateBarrierInit("file", 5, &lock, 2, 0)
	AnnotateBarrierWaitBefore("file", 6, &lock)
	AnnotateBarrierWaitAfter("file", 7, &lock)
	AnnotateBarrierDestroy("file", 8, &lock)
	AnnotateCondVarWait("file", 9, &lock, &lock)
	AnnotateCondVarSignal("file", 10, &lock)
	AnnotateCondVarSignalAll("file", 11, &lock)
	AnnotatePublishMemoryRange("file", 12, &lock, 4)
	AnnotateUnpublishMemoryRange("file", 13, &lock, 4)
	AnnotatePCQCreate("file", 14, &lock)
	AnnotatePCQDestroy("file", 15, &lock)
	AnnotatePCQPut("file", 16, &lock)
	AnnotatePCQGet("file", 17, &lock)
	AnnotateNewMemory("file", 18, &lock, 4)
	AnnotateExpectRace("file", 19, &lock, "race")
	AnnotateBenignRace("file", 20, &lock, "benign")
	AnnotateBenignRaceSized("file", 21, &lock, 4, "benign")
	AnnotateMutexIsUsedAsCondVar("file", 22, &lock)
	AnnotateTraceMemory("file", 23, &lock)
	AnnotateThreadName("file", 24, "worker")
	AnnotateIgnoreReadsBegin("file", 25)
	AnnotateIgnoreReadsEnd("file", 26)
	AnnotateIgnoreWritesBegin("file", 27)
	AnnotateIgnoreWritesEnd("file", 28)
	AnnotateIgnoreSyncBegin("file", 29)
	AnnotateIgnoreSyncEnd("file", 30)
	AnnotateEnableRaceDetection("file", 31, 1)
	AnnotateNoOp("file", 32, &lock)
	AnnotateFlushState("file", 33)
}

func TestRunningOnValgrindCachesResult(t *testing.T) {
	t.Setenv("RUNNING_ON_VALGRIND", "1")
	resetRunningOnValgrindForTest()

	first := RunningOnValgrind()

	t.Setenv("RUNNING_ON_VALGRIND", "0")
	second := RunningOnValgrind()

	if first != second {
		t.Fatalf("RunningOnValgrind cache mismatch: first=%d second=%d", first, second)
	}
}

func TestRunningOnValgrindEnvBranch(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  string
		want int
	}{
		{name: "enabled", env: "1", want: expectedRunningOnValgrind("1")},
		{name: "disabled", env: "0", want: expectedRunningOnValgrind("0")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RUNNING_ON_VALGRIND", tc.env)
			resetRunningOnValgrindForTest()
			if got := RunningOnValgrind(); got != tc.want {
				t.Fatalf("RunningOnValgrind() = %d, want %d", got, tc.want)
			}
		})
	}

	os.Unsetenv("RUNNING_ON_VALGRIND")
	resetRunningOnValgrindForTest()
	if got := RunningOnValgrind(); got != 0 {
		t.Fatalf("RunningOnValgrind() without env = %d, want 0", got)
	}
}

func resetRunningOnValgrindForTest() {
	runningOnValgrind = 0
	runningOnValgrindOnce = sync.Once{}
}
