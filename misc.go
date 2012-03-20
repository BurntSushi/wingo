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
func cliIndex(needle Client, haystack []Client) int {
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

