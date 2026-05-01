package pythread

import (
	"os"
	"sync"
)

func AnnotateRWLockCreate(file string, line int, lock any)                                   {}
func AnnotateRWLockDestroy(file string, line int, lock any)                                  {}
func AnnotateRWLockAcquired(file string, line int, lock any, isWrite int64)                  {}
func AnnotateRWLockReleased(file string, line int, lock any, isWrite int64)                  {}
func AnnotateBarrierInit(file string, line int, barrier any, count, reinitAllowed int64)     {}
func AnnotateBarrierWaitBefore(file string, line int, barrier any)                           {}
func AnnotateBarrierWaitAfter(file string, line int, barrier any)                            {}
func AnnotateBarrierDestroy(file string, line int, barrier any)                              {}
func AnnotateCondVarWait(file string, line int, cv, lock any)                                {}
func AnnotateCondVarSignal(file string, line int, cv any)                                    {}
func AnnotateCondVarSignalAll(file string, line int, cv any)                                 {}
func AnnotatePublishMemoryRange(file string, line int, address any, size int64)              {}
func AnnotateUnpublishMemoryRange(file string, line int, address any, size int64)            {}
func AnnotatePCQCreate(file string, line int, pcq any)                                       {}
func AnnotatePCQDestroy(file string, line int, pcq any)                                      {}
func AnnotatePCQPut(file string, line int, pcq any)                                          {}
func AnnotatePCQGet(file string, line int, pcq any)                                          {}
func AnnotateNewMemory(file string, line int, mem any, size int64)                           {}
func AnnotateExpectRace(file string, line int, mem any, description string)                  {}
func AnnotateBenignRace(file string, line int, mem any, description string)                  {}
func AnnotateBenignRaceSized(file string, line int, mem any, size int64, description string) {}
func AnnotateMutexIsUsedAsCondVar(file string, line int, mu any)                             {}
func AnnotateTraceMemory(file string, line int, arg any)                                     {}
func AnnotateThreadName(file string, line int, name string)                                  {}
func AnnotateIgnoreReadsBegin(file string, line int)                                         {}
func AnnotateIgnoreReadsEnd(file string, line int)                                           {}
func AnnotateIgnoreWritesBegin(file string, line int)                                        {}
func AnnotateIgnoreWritesEnd(file string, line int)                                          {}
func AnnotateIgnoreSyncBegin(file string, line int)                                          {}
func AnnotateIgnoreSyncEnd(file string, line int)                                            {}
func AnnotateEnableRaceDetection(file string, line int, enable int)                          {}
func AnnotateNoOp(file string, line int, arg any)                                            {}
func AnnotateFlushState(file string, line int)                                               {}

var (
	runningOnValgrindOnce sync.Once
	runningOnValgrind     int
)

func RunningOnValgrind() int {
	runningOnValgrindOnce.Do(func() {
		runningOnValgrind = getRunningOnValgrind()
	})
	return runningOnValgrind
}

func getRunningOnValgrind() int {
	value, ok := os.LookupEnv("RUNNING_ON_VALGRIND")
	if !ok {
		return 0
	}
	if runningOnValgrindEnvIsEnabled(value) {
		return 1
	}
	return 0
}
