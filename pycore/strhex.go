package pycore

import "errors"

var ErrInvalidHexSeparator = errors.New("sep must be length 1")

const hexDigits = "0123456789abcdef"

func StrHex(data []byte) string {
	return string(strHexImpl(data, 0, 0))
}

func StrHexBytes(data []byte) []byte {
	return strHexImpl(data, 0, 0)
}

func StrHexWithSep(data []byte, sep byte, bytesPerSepGroup int) string {
	return string(strHexImpl(data, sep, bytesPerSepGroup))
}

func StrHexBytesWithSep(data []byte, sep byte, bytesPerSepGroup int) []byte {
	return strHexImpl(data, sep, bytesPerSepGroup)
}

func StrHexWithStringSep(
	data []byte,
	sep string,
	bytesPerSepGroup int,
) (string, error) {
	if len(sep) != 1 {
		return "", ErrInvalidHexSeparator
	}
	return string(strHexImpl(data, sep[0], bytesPerSepGroup)), nil
}

func strHexImpl(data []byte, sep byte, bytesPerSepGroup int) []byte {
	arglen := len(data)
	if bytesPerSepGroup == 0 || arglen == 0 {
		return strHexNoSep(data)
	}

	absBytesPerSep := bytesPerSepGroup
	if absBytesPerSep < 0 {
		absBytesPerSep = -absBytesPerSep
	}
	if absBytesPerSep >= arglen {
		return strHexNoSep(data)
	}

	chunks := (arglen - 1) / absBytesPerSep
	out := make([]byte, arglen*2+chunks)

	if bytesPerSepGroup < 0 {
		i := 0
		j := 0
		for chunk := 0; chunk < chunks; chunk++ {
			for range absBytesPerSep {
				c := data[i]
				i++
				out[j] = hexDigits[c>>4]
				out[j+1] = hexDigits[c&0x0f]
				j += 2
			}
			out[j] = sep
			j++
		}
		for i < arglen {
			c := data[i]
			i++
			out[j] = hexDigits[c>>4]
			out[j+1] = hexDigits[c&0x0f]
			j += 2
		}
		return out
	}

	i := arglen - 1
	j := len(out) - 1
	for chunk := 0; chunk < chunks; chunk++ {
		for range absBytesPerSep {
			c := data[i]
			i--
			out[j] = hexDigits[c&0x0f]
			out[j-1] = hexDigits[c>>4]
			j -= 2
		}
		out[j] = sep
		j--
	}
	for i >= 0 {
		c := data[i]
		i--
		out[j] = hexDigits[c&0x0f]
		out[j-1] = hexDigits[c>>4]
		j -= 2
	}
	return out
}

func strHexNoSep(data []byte) []byte {
	out := make([]byte, len(data)*2)
	j := 0
	for _, c := range data {
		out[j] = hexDigits[c>>4]
		out[j+1] = hexDigits[c&0x0f]
		j += 2
	}
	return out
}
