package utils

import "strings"

// BuildLowercaseSet converts a slice of strings to a map with lowercase keys.
func BuildLowercaseSet(items []string) map[string]bool {
    set := make(map[string]bool)
    for _, item := range items {
        set[strings.ToLower(item)] = true
    }
    return set
}