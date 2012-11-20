package misc

import (
	"fmt"
	"math"
	"runtime"
	"strings"
)

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

// Prints a simple stack trace without panicing.
//
// XXX: I tried using runtime.Stack, but I couldn't get it to work...
func StackTrace() string {
	// var pc uintptr
	var fname string
	var line int
	var ok bool = true

	lines := make([]string, 0, 10)
	for i := 1; i < 200; i++ {
		_, fname, line, ok = runtime.Caller(i)
		if !ok {
			break
		}
		lines = append(lines, fmt.Sprintf("%s:%d %s", fname, line, "N/A"))
	}
	return strings.Join(lines, "\n")
}
