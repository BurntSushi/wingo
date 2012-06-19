package main

import (
	"strconv"

	"github.com/BurntSushi/wingo/logger"
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

// parsePos takes a string and parses an x or y position from it.
// The magic here is that while a string could just be a simple integer,
// it could also be a float greater than 0 but <= 1 in terms of the current
// head's geometry.
func parsePos(pos string, y bool) (int, bool) {
	// First try for an integer
	pos64, err := strconv.ParseInt(pos, 0, 0)
	if err == nil {
		return int(pos64), true
	}

	// Now try float
	posf, err := strconv.ParseFloat(pos, 64)
	if err == nil && posf > 0 && posf <= 1 {
		headGeom := wingo.workspace().Geom()
		if y {
			return headGeom.Y() + int(float64(headGeom.Height())*posf), true
		} else {
			return headGeom.X() + int(float64(headGeom.Width())*posf), true
		}
	}

	logger.Warning.Printf("'%s' is not a valid position.", pos)
	return 0, false
}

// parseDim takes a string and parses a width or height dimension from it.
// The magic here is that while a string could just be a simple integer,
// it could also be a float greater than 0 but <= 1 in terms of the current
// head's geometry.
func parseDim(dim string, height bool) (int, bool) {
	// First try for an integer
	dim64, err := strconv.ParseInt(dim, 0, 0)
	if err == nil {
		return int(dim64), true
	}

	// Now try float
	dimf, err := strconv.ParseFloat(dim, 64)
	if err == nil && dimf > 0 && dimf <= 1 {
		headGeom := wingo.workspace().Geom()
		if height {
			return int(float64(headGeom.Height()) * dimf), true
		} else {
			return int(float64(headGeom.Width()) * dimf), true
		}
	}

	logger.Warning.Printf("'%s' is not a valid dimension.", dim)
	return 0, false
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
