// A collection of miscellaneous helper functions.
package main

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

// Why isn't this in the Go standard library?
// Maybe it is and I couldn't find it...
func round(f float64) int {
    i := int(f)
    if f - float64(i) < 0.5 {
        return i
    }
    return i + 1
}

func mod(x, m int) int {
    return abs(x) % m
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

