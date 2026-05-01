package pythread

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/tamnd/gopython/pytime"
)

func TestParkingLotParkAgain(t *testing.T) {
	var value uint32 = 1
	got := defaultParkingLot.Park(uintptr(1), func() bool {
		return atomic.LoadUint32(&value) == 0
	}, pytime.Time(1_000_000), nil, false)
	if got != ParkAgain {
		t.Fatalf("Park = %d, want ParkAgain", got)
	}
}

func TestParkingLotUnpark(t *testing.T) {
	done := make(chan int, 1)
	go func() {
		got := defaultParkingLot.Park(uintptr(2), func() bool { return true }, -1, "arg", false)
		done <- got
	}()
	time.Sleep(10 * time.Millisecond)
	called := false
	defaultParkingLot.Unpark(uintptr(2), func(arg any, parkArg any, hasMore bool) {
		called = true
		if parkArg != "arg" {
			t.Fatalf("parkArg = %v, want arg", parkArg)
		}
	}, nil)
	if !called {
		t.Fatal("unpark callback not called")
	}
	if got := <-done; got != ParkOK {
		t.Fatalf("Park result = %d, want ParkOK", got)
	}
}

func TestParkingLotTimeout(t *testing.T) {
	got := defaultParkingLot.Park(uintptr(3), func() bool { return true }, pytime.Time(1_000_000), nil, false)
	if got != ParkTimeout {
		t.Fatalf("Park = %d, want ParkTimeout", got)
	}
}
