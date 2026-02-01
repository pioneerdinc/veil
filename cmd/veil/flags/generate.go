package flags

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ossydotpy/veil/internal/generator"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128
	minJWTBits        = 128
	maxJWTBits        = 512
)

// GenerateOptions holds parsed flags for the generate command.
type GenerateOptions struct {
	generator.Options
	ShowHelp bool
}

// ParseGenerateFlags parses command-line flags for the generate command.
func ParseGenerateFlags(args []string) (GenerateOptions, error) {
	opts := GenerateOptions{
		Options: generator.Options{
			Type: "password",
		},
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// All args should be flags (start with -)
		if !strings.HasPrefix(arg, "-") {
			return opts, fmt.Errorf("unexpected argument: %q (unknown flag or misplaced value)", arg)
		}

		switch arg {
		case "--type":
			if i+1 < len(args) {
				opts.Type = args[i+1]
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
		case "--length":
			if i+1 < len(args) {
				length, err := strconv.Atoi(args[i+1])
				if err != nil {
					return opts, fmt.Errorf("invalid --length value %q: must be a number", args[i+1])
				}
				if length < 0 {
					return opts, fmt.Errorf("invalid --length value %d: must be a positive number", length)
				}
				if length > 0 && length < minPasswordLength {
					return opts, fmt.Errorf("invalid --length value %d: must be at least %d", length, minPasswordLength)
				}
				if length > maxPasswordLength {
					return opts, fmt.Errorf("invalid --length value %d: must not exceed %d", length, maxPasswordLength)
				}
				opts.Length = length
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
				if bits > 0 && bits < minJWTBits {
					return opts, fmt.Errorf("invalid --bits value %d: must be at least %d", bits, minJWTBits)
				}
				if bits > maxJWTBits {
					return opts, fmt.Errorf("invalid --bits value %d: must not exceed %d", bits, maxJWTBits)
				}
				opts.Bits = bits
				i++
			}
		case "--no-symbols":
			opts.NoSymbols = true
		case "--to-env":
			if i+1 < len(args) {
				opts.ToEnv = args[i+1]
				i++
			}
		case "--force":
			opts.Force = true
		case "--help", "-h":
			opts.ShowHelp = true
		default:
			return opts, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	return opts, nil
}
