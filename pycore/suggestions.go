package pycore

import "math"

const (
	MaxCandidateItems = 750
	MaxStringSize     = 40
	MoveCost          = 2
	CaseCost          = 1
)

func UTF8EditCost(a, b string, maxCost int) int {
	if maxCost == -1 {
		maxCost = MoveCost * max(len(a), len(b))
	}
	buffer := make([]int, MaxStringSize)
	return levenshteinDistance([]byte(a), []byte(b), maxCost, buffer)
}

func CalculateSuggestion(candidates []string, name string) (string, bool) {
	if len(candidates) >= MaxCandidateItems {
		return "", false
	}

	suggestionDistance := math.MaxInt
	suggestion := ""
	found := false
	nameSize := len(name)
	buffer := make([]int, MaxStringSize)
	for _, item := range candidates {
		if item == name {
			continue
		}
		itemSize := len(item)
		maxDistance := (nameSize + itemSize + 3) * MoveCost / 6
		maxDistance = min(maxDistance, suggestionDistance-1)
		currentDistance := levenshteinDistance(
			[]byte(name),
			[]byte(item),
			maxDistance,
			buffer,
		)
		if currentDistance > maxDistance {
			continue
		}
		if !found || currentDistance < suggestionDistance {
			suggestion = item
			suggestionDistance = currentDistance
			found = true
		}
	}
	return suggestion, found
}

func levenshteinDistance(a, b []byte, maxCost int, buffer []int) int {
	if len(a) == 0 || len(b) == 0 {
		return (len(a) + len(b)) * MoveCost
	}

	for len(a) > 0 && len(b) > 0 && a[0] == b[0] {
		a = a[1:]
		b = b[1:]
	}
	for len(a) > 0 && len(b) > 0 && a[len(a)-1] == b[len(b)-1] {
		a = a[:len(a)-1]
		b = b[:len(b)-1]
	}
	if len(a) == 0 || len(b) == 0 {
		return (len(a) + len(b)) * MoveCost
	}
	if len(a) > MaxStringSize || len(b) > MaxStringSize {
		return maxCost + 1
	}

	if len(b) < len(a) {
		a, b = b, a
	}
	if (len(b)-len(a))*MoveCost > maxCost {
		return maxCost + 1
	}

	tmp := MoveCost
	for i := range len(a) {
		buffer[i] = tmp
		tmp += MoveCost
	}

	result := 0
	for bIndex, code := range b {
		distance := bIndex * MoveCost
		result = distance
		minimum := math.MaxInt
		for index := range len(a) {
			substitute := distance + substitutionCost(code, a[index])
			distance = buffer[index]
			insertDelete := min(result, distance) + MoveCost
			result = min(insertDelete, substitute)
			buffer[index] = result
			if result < minimum {
				minimum = result
			}
		}
		if minimum > maxCost {
			return maxCost + 1
		}
	}
	return result
}

func substitutionCost(a, b byte) int {
	if leastFiveBits(a) != leastFiveBits(b) {
		return MoveCost
	}
	if a == b {
		return 0
	}
	if 'A' <= a && a <= 'Z' {
		a += 'a' - 'A'
	}
	if 'A' <= b && b <= 'Z' {
		b += 'a' - 'A'
	}
	if a == b {
		return CaseCost
	}
	return MoveCost
}

func leastFiveBits(n byte) byte {
	return n & 31
}
