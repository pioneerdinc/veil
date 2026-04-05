package filter

import (
	"maps"
	"slices"
)

func FilterSecrets(secrets map[string]string, include []string, exclude []string) map[string]string {
	result := make(map[string]string)

	for k, v := range secrets {
		if MatchesFilters(k, include, exclude) {
			result[k] = v
		}
	}

	return result
}

func MatchesFilters(key string, include []string, exclude []string) bool {
	if len(include) > 0 {
		included := false
		for _, pattern := range include {
			if MatchPattern(key, pattern) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	if len(exclude) > 0 {
		for _, pattern := range exclude {
			if MatchPattern(key, pattern) {
				return false
			}
		}
	}

	return true
}

func MatchPattern(key, pattern string) bool {
	if pattern == key {
		return true
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}
	if len(pattern) > 0 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(key) >= len(suffix) && key[len(key)-len(suffix):] == suffix
	}
	return false
}

func SortKeys(m map[string]string) []string {
	return slices.Sorted(maps.Keys(m))
}
