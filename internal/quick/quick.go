// Package quick provides ephemeral secret generation without database storage.
// Secrets are generated in-memory and output to terminal or files directly.
package quick

import (
	"time"

	"github.com/ossydotpy/veil/internal/generator"
)

// Options holds configuration for quick secret generation
type Options struct {
	Type      string // password, apikey, jwt, hex, base64
	Length    int
	Format    string // uuid, hex, base64 (for apikey)
	Prefix    string
	Bits      int  // for jwt
	NoSymbols bool // for passwords
	Count     int  // generate multiple secrets

	// Output options
	ToFile   string // path to append
	EnvName  string // KEY_NAME for env file
	Force    bool   // overwrite existing in .env
	Template string // custom format template
}

// BatchConfig defines multiple secrets to generate at once
type BatchConfig struct {
	Secrets []BatchSecret `json:"secrets"`
}

// BatchSecret defines a single secret in a batch configuration
type BatchSecret struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Length    int    `json:"length,omitempty"`
	Format    string `json:"format,omitempty"`
	Bits      int    `json:"bits,omitempty"`
	NoSymbols bool   `json:"no_symbols,omitempty"`
	Count     int    `json:"count,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Template  string `json:"template,omitempty"`
}

// Result holds the generated secret and metadata
type Result struct {
	Type      string
	Value     string
	EnvName   string
	Timestamp time.Time
}

// Generator handles ephemeral secret generation
type Generator struct{}

// New creates a new quick generator
func New() *Generator {
	return &Generator{}
}

// Generate creates a single ephemeral secret
func (g *Generator) Generate(opts Options) (*Result, error) {
	genType := mapType(opts.Type)

	genOpts := generator.Options{
		Type:      genType,
		Length:    opts.Length,
		Format:    opts.Format,
		Prefix:    opts.Prefix,
		Bits:      opts.Bits,
		NoSymbols: opts.NoSymbols,
	}

	value, err := generator.Generate(genOpts)
	if err != nil {
		return nil, err
	}

	return &Result{
		Type:      opts.Type,
		Value:     value,
		EnvName:   opts.EnvName,
		Timestamp: time.Now(),
	}, nil
}

// GenerateMultiple creates multiple secrets of the same type
func (g *Generator) GenerateMultiple(opts Options) ([]*Result, error) {
	count := opts.Count
	if count <= 0 {
		count = 1
	}

	results := make([]*Result, 0, count)
	for i := 0; i < count; i++ {
		result, err := g.Generate(opts)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GenerateBatch creates multiple secrets from a batch configuration
func (g *Generator) GenerateBatch(config BatchConfig) ([]*Result, error) {
	results := make([]*Result, 0, len(config.Secrets))

	for _, secret := range config.Secrets {
		opts := Options{
			Type:      secret.Type,
			Length:    secret.Length,
			Format:    secret.Format,
			Prefix:    secret.Prefix,
			Bits:      secret.Bits,
			NoSymbols: secret.NoSymbols,
			EnvName:   secret.Name,
			Template:  secret.Template,
		}

		result, err := g.Generate(opts)
		if err != nil {
			return nil, err
		}

		// Apply template if specified in batch config
		if secret.Template != "" {
			result.Value = ApplyTemplate(secret.Template, result)
		}

		results = append(results, result)
	}

	return results, nil
}

// mapType maps quick types to generator types
// Supports shorthand types like "hex" and "base64" for apikey
func mapType(t string) string {
	switch t {
	case "", "password":
		return "password"
	case "apikey":
		return "apikey"
	case "jwt":
		return "jwt"
	case "hex", "base64", "uuid":
		// These are apikey formats, not types
		return "apikey"
	default:
		return t
	}
}

// GetFormatFromType extracts format for apikey from type shorthand
func GetFormatFromType(t string) string {
	switch t {
	case "hex":
		return "hex"
	case "base64":
		return "base64"
	case "uuid":
		return "uuid"
	default:
		return ""
	}
}
