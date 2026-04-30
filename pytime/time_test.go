package pytime

import (
	"errors"
	"math"
	"testing"
)

func TestFractionSetAndResolution(t *testing.T) {
	var frac Fraction
	if err := frac.Set(6, 8); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if frac.Numer != 3 || frac.Denom != 4 {
		t.Fatalf("reduced fraction = %d/%d, want 3/4", frac.Numer, frac.Denom)
	}
	if got := frac.Resolution(); got != 0.75e-9 {
		t.Fatalf("resolution = %g, want %g", got, 0.75e-9)
	}
}

func TestAddMulAndFractionMul(t *testing.T) {
	if got := Add(2, 3); got != 5 {
		t.Fatalf("Add = %d, want 5", got)
	}
	if got := Mul(4, 5); got != 20 {
		t.Fatalf("Mul = %d, want 20", got)
	}
	frac := Fraction{Numer: 3, Denom: 2}
	if got := FractionMul(5, frac); got != 7 {
		t.Fatalf("FractionMul = %d, want 7", got)
	}
}

func TestFromSecondsAndDoubleConversions(t *testing.T) {
	if got := FromSeconds(2); got != 2*Time(SecToNS) {
		t.Fatalf("FromSeconds = %d", got)
	}
	got, err := FromSecondsDouble(1.5, RoundHalfEven)
	if err != nil {
		t.Fatalf("FromSecondsDouble returned error: %v", err)
	}
	if got != Time(1500000000) {
		t.Fatalf("FromSecondsDouble = %d, want 1500000000", got)
	}
}

func TestFromValueConversions(t *testing.T) {
	got, err := FromSecondsValue(2, RoundFloor)
	if err != nil || got != 2*Time(SecToNS) {
		t.Fatalf("FromSecondsValue = (%d, %v)", got, err)
	}
	got, err = FromMillisecondsValue(3.5, RoundCeiling)
	if err != nil || got != 3500000 {
		t.Fatalf("FromMillisecondsValue = (%d, %v)", got, err)
	}
}

func TestObjectToTimespecAndTimeval(t *testing.T) {
	sec, nsec, err := ObjectToTimespec(1.25, RoundFloor)
	if err != nil || sec != 1 || nsec != 250000000 {
		t.Fatalf("ObjectToTimespec = (%d, %d, %v)", sec, nsec, err)
	}
	sec, usec, err := ObjectToTimeval(-1.25, RoundFloor)
	if err != nil || sec != -2 || usec != 750000 {
		t.Fatalf("ObjectToTimeval = (%d, %d, %v)", sec, usec, err)
	}
}

func TestAsConversions(t *testing.T) {
	if got := AsSecondsDouble(2 * Time(SecToNS)); got != 2 {
		t.Fatalf("AsSecondsDouble = %v, want 2", got)
	}
	if got := AsMicroseconds(1501, RoundCeiling); got != 2 {
		t.Fatalf("AsMicroseconds = %d, want 2", got)
	}
	if got := AsMilliseconds(-1500001, RoundFloor); got != -2 {
		t.Fatalf("AsMilliseconds = %d, want -2", got)
	}
	tv, err := AsTimeval(1500000000, RoundFloor)
	if err != nil || tv.Sec != 1 || tv.Usec != 500000 {
		t.Fatalf("AsTimeval = (%+v, %v)", tv, err)
	}
	ts, err := AsTimespec(-1)
	if err != nil || ts.Sec != -1 || ts.Nsec != 999999999 {
		t.Fatalf("AsTimespec = (%+v, %v)", ts, err)
	}
}

func TestRoundingModes(t *testing.T) {
	if got := divide(5, 2, RoundHalfEven); got != 2 {
		t.Fatalf("half-even 5/2 = %d, want 2", got)
	}
	if got := divide(7, 2, RoundHalfEven); got != 4 {
		t.Fatalf("half-even 7/2 = %d, want 4", got)
	}
	if got := divide(-3, 2, RoundUp); got != -2 {
		t.Fatalf("round-up -3/2 = %d, want -2", got)
	}
}

func TestErrors(t *testing.T) {
	if _, err := FromSecondsDouble(math.NaN(), RoundFloor); !errors.Is(err, ErrTimeNaN) {
		t.Fatalf("unexpected NaN error: %v", err)
	}
	if _, err := FromSecondsValue("x", RoundFloor); !errors.Is(err, ErrTimeType) {
		t.Fatalf("unexpected type error: %v", err)
	}
	if _, _, err := ObjectToTimespec(math.NaN(), RoundFloor); !errors.Is(err, ErrTimeNaN) {
		t.Fatalf("unexpected timespec NaN error: %v", err)
	}
}
