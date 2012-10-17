package misc

import "math"

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// This exists because '%' isn't really modulus; it's *remainder*.
// e.g., (-1) % 2 = -1 but (-1) mod 2 = 1.
func Mod(x, m int) int {
	r := x % m
	if r < 0 {
		r += m
	}
	return r
}

func Round(n float64) int {
	ceil := math.Ceil(n)
	if ceil-n <= 0.5 {
		return int(ceil)
	}
	return int(ceil) - 1
}
