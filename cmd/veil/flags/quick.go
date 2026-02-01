package flags

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ossydotpy/veil/internal/quick"
)

const (
	minQuickPasswordLength = 1
	maxQuickPasswordLength = 512
	minQuickJWTBits        = 128
	maxQuickJWTBits        = 512
	maxQuickCount          = 100
)

// QuickOptions holds parsed flags for the quick command.
type QuickOptions struct {
	quick.Options
	BatchFile string
}

// ParseQuickFlags parses command-line flags for the quick command.
func ParseQuickFlags(args []string) (QuickOptions, error) {
	opts := QuickOptions{
		Options: quick.Options{
			Type: "password",
		},
	}

	i := 0

	// First non-flag argument is the type (optional)
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		opts.Type = args[0]

		// Handle shorthand types (hex, base64, uuid -> apikey with format)
		if format := quick.GetFormatFromType(args[0]); format != "" {
			opts.Type = "apikey"
			opts.Format = format
		}

		i = 1
	}

	for ; i < len(args); i++ {
		switch args[i] {
		case "--length":
			if i+1 < len(args) {
				length, err := strconv.Atoi(args[i+1])
				if err != nil {
					return opts, fmt.Errorf("invalid --length value %q: must be a number", args[i+1])
				}
				if length < 0 {
					return opts, fmt.Errorf("invalid --length value %d: must be a positive number", length)
				}
				if length > 0 && length < minQuickPasswordLength {
					return opts, fmt.Errorf("invalid --length value %d: must be at least %d", length, minQuickPasswordLength)
				}
				if length > maxQuickPasswordLength {
					return opts, fmt.Errorf("invalid --length value %d: must not exceed %d", length, maxQuickPasswordLength)
				}
				opts.Length = length
				i++
			}
		case "--format":
			if i+1 < len(args) {
				opts.Format = args[i+1]
				i++
			}
		case "--prefix":
			if i+1 < len(args) {
				opts.Prefix = args[i+1]
				i++
			}
		case "--bits":
			if i+1 < len(args) {
				bits, err := strconv.Atoi(args[i+1])
				if err != nil {
					return opts, fmt.Errorf("invalid --bits value %q: must be a number", args[i+1])
				}
				if bits < 0 {
					return opts, fmt.Errorf("invalid --bits value %d: must be a positive number", bits)
				}
				if bits > 0 && bits < minQuickJWTBits {
					return opts, fmt.Errorf("invalid --bits value %d: must be at least %d", bits, minQuickJWTBits)
				}
				if bits > maxQuickJWTBits {
					return opts, fmt.Errorf("invalid --bits value %d: must not exceed %d", bits, maxQuickJWTBits)
				}
				opts.Bits = bits
				i++
			}
		case "--no-symbols":
			opts.NoSymbols = true
		case "--count":
			if i+1 < len(args) {
				count, err := strconv.Atoi(args[i+1])
				if err != nil {
					return opts, fmt.Errorf("invalid --count value %q: must be a number", args[i+1])
				}
				if count < 0 {
					return opts, fmt.Errorf("invalid --count value %d: must be a positive number", count)
				}
				if count > maxQuickCount {
					return opts, fmt.Errorf("invalid --count value %d: must not exceed %d", count, maxQuickCount)
				}
				opts.Count = count
				i++
			}
		case "--to":
			if i+1 < len(args) {
				opts.ToFile = args[i+1]
				i++
			}
		case "--name":
			if i+1 < len(args) {
				opts.EnvName = args[i+1]
				i++
			}
		case "--force":
			opts.Force = true
		case "--template":
			if i+1 < len(args) {
				opts.Template = args[i+1]
				i++
			}
		case "--batch":
			if i+1 < len(args) {
				opts.BatchFile = args[i+1]
				i++
			}
		}
	}

	return opts, nil
}
