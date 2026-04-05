package envfile

import (
	"fmt"
	"os"

	"github.com/ossydotpy/veil/internal/encoding/env"
)

func ParseEnvFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	return env.Parse(data)
}

func UnescapeValue(value string) string {
	return env.UnescapeValue(value)
}
