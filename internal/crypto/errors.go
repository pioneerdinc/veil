package crypto

import "errors"

// Common errors for the crypto package
var (
	ErrInvalidKeyFormat = errors.New("invalid master key format")

	ErrInvalidKeyLength = errors.New("invalid key length")

	ErrCiphertextTooShort = errors.New("ciphertext too short")

	ErrDecryptionFailed = errors.New("decryption failed")
)
