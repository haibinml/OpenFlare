package utils

import "strings"

// Unique returns a new slice containing only the unique elements of the input slice,
// preserving their original order.
func Unique[T comparable](slice []T) []T {
	if slice == nil {
		return nil
	}
	seen := make(map[T]struct{})
	result := make([]T, 0)
	for _, item := range slice {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

// UniqueAndCleanStringSlice trims spaces, removes empty elements, and returns only the unique elements
// of the input string slice. It preserves order and returns nil if the resulting slice is empty.
func UniqueAndCleanStringSlice(slice []string) []string {
	if slice == nil {
		return nil
	}
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, item := range slice {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
