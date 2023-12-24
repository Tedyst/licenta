package file

import (
	"math"
)

func shannonEntropy(value string) (bits int) {
	frq := make(map[rune]float64)

	//get frequency of characters
	for _, i := range value {
		frq[i]++
	}

	var sum float64

	for _, v := range frq {
		f := v / float64(len(value))
		sum += f * math.Log2(f)
	}

	bits = int(math.Ceil(sum*-1)) * len(value)
	return
}

func isASCII(s []byte) bool {
	for i := range s {
		if s[i] < 32 || s[i] > 126 {
			return false
		}
	}
	return true
}
