package pytime

import (
	"errors"
	"math"
	"runtime"
	"sync"
	"time"
)

const (
	SecToMS = int64(1000)
	MSToUS  = int64(1000)
	SecToUS = SecToMS * MSToUS
	USToNS  = int64(1000)
	MSToNS  = MSToUS * USToNS
	SecToNS = SecToMS * MSToNS
	NSToMS  = int64(1000 * 1000)
	NSToUS  = int64(1000)
)

type Time int64

type RoundMode int

const (
	RoundHalfEven RoundMode = iota
	RoundCeiling
	RoundFloor
	RoundUp
)

var (
	ErrTimeOverflow = errors.New("timestamp too large to convert to C PyTime_t")
	ErrTimeNaN      = errors.New("invalid value NaN (not a number)")
	ErrTimeType     = errors.New("argument must be int or float")
)

type Fraction struct {
	Numer Time
	Denom Time
}

type Timeval struct {
	Sec  Time
	Usec int
}

type Timespec struct {
	Sec  Time
	Nsec int
}

type ClockInfo struct {
	Implementation string
	Monotonic      bool
	Adjustable     bool
	Resolution     float64
}

type RuntimeState struct {
	Base Fraction
}

var (
	runtimeState RuntimeState
	initOnce     sync.Once
	startTime    = time.Now()
)

func GCD(x, y Time) Time {
	for y != 0 {
		x, y = y, x%y
	}
	return x
}

func Init(state *RuntimeState) error {
	var err error
	initOnce.Do(func() {
		switch runtime.GOOS {
		case "windows":
			err = runtimeState.Base.Set(Time(SecToNS), 10_000_000)
		case "darwin":
			err = runtimeState.Base.Set(1, 1)
		default:
			err = runtimeState.Base.Set(1, 1)
		}
	})
	if err != nil {
		return err
	}
	if state != nil {
		state.Base = runtimeState.Base
	}
	return nil
}

func (f *Fraction) Set(numer, denom Time) error {
	if numer < 1 || denom < 1 {
		return ErrTimeOverflow
	}
	gcd := GCD(numer, denom)
	f.Numer = numer / gcd
	f.Denom = denom / gcd
	return nil
}

func (f Fraction) Resolution() float64 {
	return float64(f.Numer) / float64(f.Denom) / 1e9
}

func Add(t1, t2 Time) Time {
	if t2 > 0 && t1 > Time(math.MaxInt64)-t2 {
		return Time(math.MaxInt64)
	}
	if t2 < 0 && t1 < Time(math.MinInt64)-t2 {
		return Time(math.MinInt64)
	}
	return t1 + t2
}

func Mul(t, k Time) Time {
	if k == 0 {
		return 0
	}
	if t < Time(math.MinInt64)/k || t > Time(math.MaxInt64)/k {
		if t >= 0 {
			return Time(math.MaxInt64)
		}
		return Time(math.MinInt64)
	}
	return t * k
}

func FractionMul(ticks Time, frac Fraction) Time {
	if frac.Denom == 1 {
		return Mul(ticks, frac.Numer)
	}
	intpart := ticks / frac.Denom
	remainingTicks := ticks % frac.Denom
	remaining := Mul(remainingTicks, frac.Numer) / frac.Denom
	return Add(Mul(intpart, frac.Numer), remaining)
}

func FromSeconds(seconds int) Time {
	return Time(seconds) * Time(SecToNS)
}

func FromMicrosecondsClamp(us Time) Time {
	return Mul(us, Time(USToNS))
}

func FromSecondsDouble(seconds float64, round RoundMode) (Time, error) {
	return fromDouble(seconds, round, SecToNS)
}

func FromSecondsValue(value any, round RoundMode) (Time, error) {
	return fromValue(value, round, SecToNS)
}

func FromMillisecondsValue(value any, round RoundMode) (Time, error) {
	return fromValue(value, round, MSToNS)
}

func ObjectToTimespec(value any, round RoundMode) (Time, int64, error) {
	return objectToDenominator(value, SecToNS, round)
}

func ObjectToTimeval(value any, round RoundMode) (Time, int64, error) {
	return objectToDenominator(value, SecToUS, round)
}

func AsSecondsDouble(ns Time) float64 {
	if ns%Time(SecToNS) == 0 {
		return float64(ns / Time(SecToNS))
	}
	return float64(ns) / 1e9
}

func AsMicroseconds(ns Time, round RoundMode) Time {
	return divide(ns, Time(NSToUS), round)
}

func AsMilliseconds(ns Time, round RoundMode) Time {
	return divide(ns, Time(NSToMS), round)
}

func AsTimeval(ns Time, round RoundMode) (Timeval, error) {
	us := divide(ns, Time(USToNS), round)
	sec, usec, err := divmod(us, Time(SecToUS))
	return Timeval{Sec: sec, Usec: int(usec)}, err
}

func AsTimespec(ns Time) (Timespec, error) {
	sec, nsec, err := divmod(ns, Time(SecToNS))
	return Timespec{Sec: sec, Nsec: int(nsec)}, err
}

func TimeNow() (Time, error) {
	return systemClock(nil)
}

func TimeNowRaw() (Time, error) {
	return systemClock(nil)
}

func TimeWithInfo() (Time, ClockInfo, error) {
	info := ClockInfo{}
	t, err := systemClock(&info)
	return t, info, err
}

func Monotonic() (Time, error) {
	return monotonicClock(nil)
}

func MonotonicRaw() (Time, error) {
	return monotonicClock(nil)
}

func MonotonicWithInfo() (Time, ClockInfo, error) {
	info := ClockInfo{}
	t, err := monotonicClock(&info)
	return t, info, err
}

func PerfCounter() (Time, error) {
	return Monotonic()
}

func PerfCounterRaw() (Time, error) {
	return MonotonicRaw()
}

func PerfCounterWithInfo() (Time, ClockInfo, error) {
	return MonotonicWithInfo()
}

func Localtime(t time.Time) time.Time {
	return t.Local()
}

func GMTime(t time.Time) time.Time {
	return t.UTC()
}

func DeadlineInit(timeout Time) Time {
	now, _ := MonotonicRaw()
	return Add(now, timeout)
}

func DeadlineGet(deadline Time) Time {
	now, _ := MonotonicRaw()
	return deadline - now
}

func roundHalfEven(x float64) float64 {
	rounded := math.Round(x)
	if math.Abs(x-rounded) == 0.5 {
		rounded = 2.0 * math.Round(x/2.0)
	}
	return rounded
}

func roundFloat(x float64, round RoundMode) float64 {
	switch round {
	case RoundHalfEven:
		return roundHalfEven(x)
	case RoundCeiling:
		return math.Ceil(x)
	case RoundFloor:
		return math.Floor(x)
	default:
		if x >= 0 {
			return math.Ceil(x)
		}
		return math.Floor(x)
	}
}

func objectToDenominator(value any, denominator int64, round RoundMode) (Time, int64, error) {
	switch v := value.(type) {
	case float64:
		if math.IsNaN(v) {
			return 0, 0, ErrTimeNaN
		}
		secFloat, frac := math.Modf(v)
		floatpart := roundFloat(frac*float64(denominator), round)
		intpart := secFloat
		if floatpart >= float64(denominator) {
			floatpart -= float64(denominator)
			intpart += 1
		} else if floatpart < 0 {
			floatpart += float64(denominator)
			intpart -= 1
		}
		if intpart < float64(math.MinInt64) || intpart >= -float64(math.MinInt64) {
			return 0, 0, ErrTimeOverflow
		}
		return Time(int64(intpart)), int64(floatpart), nil
	case int:
		return Time(v), 0, nil
	case int64:
		return Time(v), 0, nil
	default:
		return 0, 0, ErrTimeType
	}
}

func fromDouble(value float64, round RoundMode, unitToNS int64) (Time, error) {
	if math.IsNaN(value) {
		return 0, ErrTimeNaN
	}
	d := roundFloat(value*float64(unitToNS), round)
	if d < float64(math.MinInt64) || d >= -float64(math.MinInt64) {
		return 0, ErrTimeOverflow
	}
	return Time(int64(d)), nil
}

func fromValue(value any, round RoundMode, unitToNS int64) (Time, error) {
	switch v := value.(type) {
	case float64:
		return fromDouble(v, round, unitToNS)
	case int:
		return mulChecked(Time(v), Time(unitToNS))
	case int64:
		return mulChecked(Time(v), Time(unitToNS))
	default:
		return 0, ErrTimeType
	}
}

func mulChecked(t, k Time) (Time, error) {
	result := Mul(t, k)
	if result == Time(math.MaxInt64) || result == Time(math.MinInt64) {
		if t != 0 && result/t != k {
			return 0, ErrTimeOverflow
		}
	}
	return result, nil
}

func divideRoundUp(t, k Time) Time {
	if t >= 0 {
		q := t / k
		if t%k != 0 {
			q++
		}
		return q
	}
	q := t / k
	if t%k != 0 {
		q--
	}
	return q
}

func divide(t, k Time, round RoundMode) Time {
	switch round {
	case RoundHalfEven:
		x := t / k
		r := t % k
		absR := r
		if absR < 0 {
			absR = -absR
		}
		absX := x
		if absX < 0 {
			absX = -absX
		}
		if absR > k/2 || (absR == k/2 && absX&1 == 1) {
			if t >= 0 {
				x++
			} else {
				x--
			}
		}
		return x
	case RoundCeiling:
		if t >= 0 {
			return divideRoundUp(t, k)
		}
		return t / k
	case RoundFloor:
		if t >= 0 {
			return t / k
		}
		return divideRoundUp(t, k)
	default:
		return divideRoundUp(t, k)
	}
}

func divmod(t, k Time) (Time, Time, error) {
	q := t / k
	r := t % k
	if r < 0 {
		if q == Time(math.MinInt64) {
			return Time(math.MinInt64), 0, ErrTimeOverflow
		}
		r += k
		q--
	}
	return q, r, nil
}

func systemClock(info *ClockInfo) (Time, error) {
	now := time.Now()
	if info != nil {
		info.Implementation = "time.Now()"
		info.Monotonic = false
		info.Adjustable = true
		info.Resolution = 1e-9
	}
	return Time(now.UnixNano()), nil
}

func monotonicClock(info *ClockInfo) (Time, error) {
	if err := Init(nil); err != nil {
		return 0, err
	}
	if info != nil {
		info.Monotonic = true
		info.Adjustable = false
		info.Resolution = runtimeState.Base.Resolution()
		switch runtime.GOOS {
		case "windows":
			info.Implementation = "QueryPerformanceCounter()"
		case "darwin":
			info.Implementation = "mach_absolute_time()"
		default:
			info.Implementation = "clock_gettime(CLOCK_MONOTONIC)"
		}
	}
	return Time(time.Since(startTime).Nanoseconds()), nil
}
