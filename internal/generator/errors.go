package generator

import "errors"

// Common errors for the generator package
var (
	ErrInvalidLength = errors.New("invalid password length")

	ErrInvalidJWTBits = errors.New("invalid JWT bits")

	ErrUnknownGeneratorType = errors.New("unknown generator type")

	ErrUnsupportedFormat = errors.New("unsupported API key format")
)
