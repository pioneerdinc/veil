package generator

import (
	"crypto/rand"
	"fmt"
)

// PasswordGenerator generates secure passwords
type PasswordGenerator struct{}

const (
	defaultPasswordLength = 32
	minPasswordLength     = 8
	maxPasswordLength     = 128
)

var (
	// Charsets for password generation
	lowercase    = "abcdefghijklmnopqrstuvwxyz"
	uppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers      = "0123456789"
	symbols      = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	alphanumeric = lowercase + uppercase + numbers
	fullCharset  = lowercase + uppercase + numbers + symbols
)

func (g *PasswordGenerator) Generate(opts Options) (string, error) {
	length := opts.Length
	if length == 0 {
		length = defaultPasswordLength
	}
	if length < minPasswordLength {
		return "", fmt.Errorf("password length must be at least %d characters", minPasswordLength)
	}
	if length > maxPasswordLength {
		return "", fmt.Errorf("password length must not exceed %d characters", maxPasswordLength)
	}

	charset := fullCharset
	if opts.NoSymbols {
		charset = alphanumeric
	}

	return generateRandomString(length, charset)
}

func generateRandomString(length int, charset string) (string, error) {
	result := make([]byte, length)
	charsetLen := byte(len(charset))

	for i := range length {
		randomByte := make([]byte, 1)
		_, err := rand.Read(randomByte)
		if err != nil {
			return "", fmt.Errorf("failed to generate random bytes: %w", err)
		}
		result[i] = charset[randomByte[0]%charsetLen]
	}

	return string(result), nil
}
