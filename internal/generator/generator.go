package generator

import (
	"fmt"
)

type Generator interface {
	Generate(opts Options) (string, error)
}

// Get returns the appropriate generator for the given type
func Get(secretType string) (Generator, error) {
	switch secretType {
	case "", "password":
		return &PasswordGenerator{}, nil
	case "apikey":
		return &APIKeyGenerator{}, nil
	case "jwt":
		return &JWTGenerator{}, nil
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnknownGeneratorType, secretType)
	}
}

func Generate(opts Options) (string, error) {
	gen, err := Get(opts.Type)
	if err != nil {
		return "", err
	}
	return gen.Generate(opts)
}
