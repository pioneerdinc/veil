package generator

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// JWTGenerator generates JWT signing secrets
type JWTGenerator struct{}

const (
	defaultJWTBits = 256
	minJWTBits     = 128
	maxJWTBits     = 512
)

func (g *JWTGenerator) Generate(opts Options) (string, error) {
	bits := opts.Bits
	if bits == 0 {
		bits = defaultJWTBits
	}

	if bits < minJWTBits {
		return "", fmt.Errorf("%s: must be at least %d bits", ErrInvalidJWTBits, minJWTBits)
	}
	if bits > maxJWTBits {
		return "", fmt.Errorf("%s: must not exceed %d bits", ErrInvalidJWTBits, maxJWTBits)
	}

	// Convert bits to bytes (8 bits per byte)
	bytesLen := bits / 8
	if bits%8 != 0 {
		bytesLen++ // Round up if not divisible by 8
	}

	bytes := make([]byte, bytesLen)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}
