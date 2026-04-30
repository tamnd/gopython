package pycore

const (
	CTFLower  uint = 0x01
	CTFUpper  uint = 0x02
	CTFAlpha       = CTFLower | CTFUpper
	CTFDigit  uint = 0x04
	CTFAlnum       = CTFAlpha | CTFDigit
	CTFSpace  uint = 0x08
	CTFXDigit uint = 0x10
)

var ctypeTable [256]uint
var ctypeToLower [256]byte
var ctypeToUpper [256]byte

func init() {
	for i := range 256 {
		b := byte(i)
		ctypeToLower[i] = b
		ctypeToUpper[i] = b
	}
	for _, b := range []byte{'\t', '\n', '\v', '\f', '\r', ' '} {
		ctypeTable[b] |= CTFSpace
	}
	for b := byte('0'); b <= '9'; b++ {
		ctypeTable[b] |= CTFDigit | CTFXDigit
	}
	for b := byte('A'); b <= 'Z'; b++ {
		ctypeTable[b] |= CTFUpper
		ctypeToLower[b] = b + ('a' - 'A')
	}
	for b := byte('a'); b <= 'z'; b++ {
		ctypeTable[b] |= CTFLower
		ctypeToUpper[b] = b - ('a' - 'A')
	}
	for b := byte('A'); b <= 'F'; b++ {
		ctypeTable[b] |= CTFXDigit
	}
	for b := byte('a'); b <= 'f'; b++ {
		ctypeTable[b] |= CTFXDigit
	}
}

// Charmask mirrors CPython's Py_CHARMASK.
func Charmask(c int) byte {
	return byte(c & 0xff)
}

func CTypeFlags(c int) uint {
	return ctypeTable[Charmask(c)]
}

func IsLower(c int) bool {
	return CTypeFlags(c)&CTFLower != 0
}

func IsUpper(c int) bool {
	return CTypeFlags(c)&CTFUpper != 0
}

func IsAlpha(c int) bool {
	return CTypeFlags(c)&CTFAlpha != 0
}

func IsDigit(c int) bool {
	return CTypeFlags(c)&CTFDigit != 0
}

func IsXDigit(c int) bool {
	return CTypeFlags(c)&CTFXDigit != 0
}

func IsAlnum(c int) bool {
	return CTypeFlags(c)&CTFAlnum != 0
}

func IsSpace(c int) bool {
	return CTypeFlags(c)&CTFSpace != 0
}

func ToLower(c int) byte {
	return ctypeToLower[Charmask(c)]
}

func ToUpper(c int) byte {
	return ctypeToUpper[Charmask(c)]
}
