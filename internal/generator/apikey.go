package generator

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
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
	case "hex":
		secret, err = generateHex(length)
	case "base64":
		secret, err = generateBase64(length)
	default:
		return "", fmt.Errorf("unsupported API key format: %s", format)
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
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}

	// Set version (4) and variant (RFC 4122)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
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
