package pycore

import (
	"fmt"
	"math"
)

func PyOSSnprintf(buf []byte, format string, args ...any) int {
	if len(buf) == 0 {
		panic("PyOSSnprintf requires non-empty buffer")
	}
	if len(buf) > math.MaxInt32-1 {
		buf[len(buf)-1] = 0
		return -666
	}

	out := fmt.Sprintf(format, args...)
	n := copy(buf, out)
	if n < len(buf) {
		buf[n] = 0
	}
	buf[len(buf)-1] = 0
	return len(out)
}
