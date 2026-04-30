package pycore

import "errors"

const (
	ulongMax = ^uint64(0)
	longMax  = int64(ulongMax >> 1)
	longMin  = -longMax - 1
)

var ErrRange = errors.New("result out of range")

var digitValue [256]byte

var smallMax = [37]uint64{
	0, 0,
	ulongMax / 2, ulongMax / 3, ulongMax / 4, ulongMax / 5,
	ulongMax / 6, ulongMax / 7, ulongMax / 8, ulongMax / 9,
	ulongMax / 10, ulongMax / 11, ulongMax / 12, ulongMax / 13,
	ulongMax / 14, ulongMax / 15, ulongMax / 16, ulongMax / 17,
	ulongMax / 18, ulongMax / 19, ulongMax / 20, ulongMax / 21,
	ulongMax / 22, ulongMax / 23, ulongMax / 24, ulongMax / 25,
	ulongMax / 26, ulongMax / 27, ulongMax / 28, ulongMax / 29,
	ulongMax / 30, ulongMax / 31, ulongMax / 32, ulongMax / 33,
	ulongMax / 34, ulongMax / 35, ulongMax / 36,
}

var digitLimit = [37]int{
	0, 0, 64, 40, 32, 27, 24, 22, 21, 20,
	19, 18, 17, 17, 16, 16, 16, 15, 15, 15,
	14, 14, 14, 14, 13, 13, 13, 13, 13, 13,
	13, 12, 12, 12, 12, 12, 12,
}

func init() {
	for i := range 256 {
		digitValue[i] = 37
	}
	for c := byte('0'); c <= '9'; c++ {
		digitValue[c] = c - '0'
	}
	for c := byte('A'); c <= 'Z'; c++ {
		digitValue[c] = c - 'A' + 10
	}
	for c := byte('a'); c <= 'z'; c++ {
		digitValue[c] = c - 'a' + 10
	}
}

func PyOSStrtoul(s string, base int) (value uint64, end int, err error) {
	str := 0
	for cStringByte(s, str) != 0 && IsSpace(int(cStringByte(s, str))) {
		str++
	}

	switch base {
	case 0:
		if cStringByte(s, str) == '0' {
			str++
			switch cStringByte(s, str) {
			case 'x', 'X':
				if longDigitValue(s, str+1) >= 16 {
					return 0, str, nil
				}
				str++
				base = 16
			case 'o', 'O':
				if longDigitValue(s, str+1) >= 8 {
					return 0, str, nil
				}
				str++
				base = 8
			case 'b', 'B':
				if longDigitValue(s, str+1) >= 2 {
					return 0, str, nil
				}
				str++
				base = 2
			default:
				for cStringByte(s, str) == '0' {
					str++
				}
				for IsSpace(int(cStringByte(s, str))) {
					str++
				}
				return 0, str, nil
			}
		} else {
			base = 10
		}
	case 16:
		if cStringByte(s, str) == '0' {
			str++
			if cStringByte(s, str) == 'x' || cStringByte(s, str) == 'X' {
				if longDigitValue(s, str+1) >= 16 {
					return 0, str, nil
				}
				str++
			}
		}
	case 8:
		if cStringByte(s, str) == '0' {
			str++
			if cStringByte(s, str) == 'o' || cStringByte(s, str) == 'O' {
				if longDigitValue(s, str+1) >= 8 {
					return 0, str, nil
				}
				str++
			}
		}
	case 2:
		if cStringByte(s, str) == '0' {
			str++
			if cStringByte(s, str) == 'b' || cStringByte(s, str) == 'B' {
				if longDigitValue(s, str+1) >= 2 {
					return 0, str, nil
				}
				str++
			}
		}
	}

	if base < 2 || base > 36 {
		return 0, str, nil
	}

	for cStringByte(s, str) == '0' {
		str++
	}

	ovlimit := digitLimit[base]
	result := uint64(0)
	for c := uint64(longDigitValue(s, str)); c < uint64(base); c = uint64(longDigitValue(s, str)) {
		if ovlimit > 0 {
			result = result*uint64(base) + c
		} else {
			if ovlimit < 0 {
				return ulongMax, spoolDigits(s, str, base), ErrRange
			}
			if result > smallMax[base] {
				return ulongMax, spoolDigits(s, str, base), ErrRange
			}
			result *= uint64(base)
			temp := result + c
			if temp < result {
				return ulongMax, spoolDigits(s, str, base), ErrRange
			}
			result = temp
		}
		str++
		ovlimit--
	}
	return result, str, nil
}

func PyOSStrtol(s string, base int) (value int64, end int, err error) {
	str := 0
	for cStringByte(s, str) != 0 && IsSpace(int(cStringByte(s, str))) {
		str++
	}

	sign := cStringByte(s, str)
	if sign == '+' || sign == '-' {
		str++
	}

	uresult, relEnd, uerr := PyOSStrtoul(s[str:], base)
	end = str + relEnd
	if uresult <= uint64(longMax) {
		result := int64(uresult)
		if sign == '-' {
			result = -result
		}
		return result, end, uerr
	}
	if sign == '-' && uresult == uint64(longMax)+1 {
		return longMin, end, uerr
	}
	return longMax, end, ErrRange
}

func longDigitValue(s string, i int) byte {
	return digitValue[Charmask(int(cStringByte(s, i)))]
}

func spoolDigits(s string, i int, base int) int {
	for int(longDigitValue(s, i)) < base {
		i++
	}
	return i
}
