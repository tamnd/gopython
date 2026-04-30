package pycore

func MyStrNICmp(s1, s2 string, size int) int {
	if size == 0 {
		return 0
	}
	p1 := 0
	p2 := 0
	for {
		size--
		c1 := cStringByte(s1, p1)
		c2 := cStringByte(s2, p2)
		if !(size > 0 && c1 != 0 && c2 != 0 &&
			ToLower(int(c1)) == ToLower(int(c2))) {
			return int(ToLower(int(c1))) - int(ToLower(int(c2)))
		}
		p1++
		p2++
	}
}

func MyStrICmp(s1, s2 string) int {
	p1 := 0
	p2 := 0
	for {
		c1 := cStringByte(s1, p1)
		c2 := cStringByte(s2, p2)
		if !(c1 != 0 && c2 != 0 && ToLower(int(c1)) == ToLower(int(c2))) {
			return int(ToLower(int(c1))) - int(ToLower(int(c2)))
		}
		p1++
		p2++
	}
}

func cStringByte(s string, i int) byte {
	if i >= len(s) {
		return 0
	}
	return s[i]
}
