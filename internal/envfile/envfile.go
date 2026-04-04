package envfile

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

func ParseEnvFile(path string) (map[string]string, error) {
	result := make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

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

	return result, nil
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
