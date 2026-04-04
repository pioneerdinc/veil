package env

import (
	"bufio"
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"
)

func Parse(data []byte) (map[string]string, error) {
	result := make(map[string]string)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if key, value, found := strings.Cut(line, "="); found {
			result[key] = UnescapeValue(value)
		}
	}

	return result, scanner.Err()
}

func Marshal(secrets map[string]string) []byte {
	var buf bytes.Buffer

	keys := slices.Sorted(maps.Keys(secrets))
	for _, key := range keys {
		fmt.Fprintf(&buf, "%s=%s\n", key, EscapeValue(secrets[key]))
	}

	return buf.Bytes()
}

func EscapeValue(value string) string {
	if needsQuoting(value) {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
	}
	return value
}

func UnescapeValue(value string) string {
	value = strings.TrimSpace(value)

	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = value[1 : len(value)-1]
		value = strings.ReplaceAll(value, "\\\"", "\"")
		return value
	}

	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		value = value[1 : len(value)-1]
		return value
	}

	return value
}

func needsQuoting(value string) bool {
	if value == "" {
		return true
	}

	for _, r := range value {
		switch r {
		case ' ', '\t', '\n', '\r', '#', '"', '\'':
			return true
		}
	}

	return false
}
