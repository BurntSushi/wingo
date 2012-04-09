// A collection of miscellaneous helper functions.
package main

import (
	"strings"

	"github.com/BurntSushi/xgbutil/keybind"
)

// strIndex returns the index of the first occurrence of needle in haystack.
// Returns -1 if needle is not in haystack.
func strIndex(needle string, haystack []string) int {
	for i, possible := range haystack {
		if needle == possible {
			return i
		}
	}
	return -1
}

// cliIndex returns the index of the first occurrence of needle in haystack.
// Returns -1 if needle is not in haystack.
func cliIndex(needle *client, haystack []*client) int {
	for i, possible := range haystack {
		if needle == possible {
			return i
		}
	}
	return -1
}

// keyMatch is a utility function for comparing two keysym strings for equality.
// It automatically converts a (mods, byte) pair to a string.
func keyMatch(target string, mods uint16, keycode byte) bool {
	guess := keybind.LookupString(X, mods, keycode)
	return strings.ToLower(guess) == strings.ToLower(target)
}

// Why isn't this in the Go standard library?
// Maybe it is and I couldn't find it...
func round(f float64) int {
	i := int(f)
	if f-float64(i) < 0.5 {
		return i
	}
	return i + 1
}

// This exists because '%' isn't really modulus; it's *remainder*.
// e.g., (-1) % 2 = -1 but (-1) mod 2 = 1.
func mod(x, m int) int {
	r := x % m
	if r < 0 {
		r += m
	}
	return r
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
