package generator

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

// APIKeyGenerator generates API keys in various formats
type APIKeyGenerator struct{}

const defaultAPIKeyLength = 32

func (g *APIKeyGenerator) Generate(opts Options) (string, error) {
	format := opts.Format
	if format == "" {
		format = "base64"
	}

	length := opts.Length
	if length == 0 {
		length = defaultAPIKeyLength
	}

	var secret string
	var err error

	switch format {
	case "uuid":
		secret, err = generateUUID()
	case "uuidv7":
		secret, err = generateUUIDv7()
	case "hex":
		secret, err = generateHex(length)
	case "base64":
		secret, err = generateBase64(length)
	default:
		return "", fmt.Errorf("%s: %s", ErrUnsupportedFormat, format)
	}

	if err != nil {
		return "", err
	}

	// Add prefix if provided
	if opts.Prefix != "" {
		secret = opts.Prefix + secret
	}

	return secret, nil
}

// generateUUID generates a UUID v4
func generateUUID() (string, error) {
	id := uuid.New()
	return id.String(), nil
}

// generateUUIDv7 generates a UUID v7
func generateUUIDv7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID v7: %w", err)
	}
	return id.String(), nil
}

// generateHex generates a hex-encoded random string
func generateHex(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate hex: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// generateBase64 generates a URL-safe base64-encoded random string
func generateBase64(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate base64: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
