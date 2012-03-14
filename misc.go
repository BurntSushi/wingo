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
func cliIndex(needle client, haystack []client) int {
    for i, possible := range haystack {
        if needle == possible {
            return i
        }
    }
    return -1
}

