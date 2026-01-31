package flags

import (
	"strconv"

	"github.com/ossydotpy/veil/internal/generator"
)

// ParseGenerateFlags parses command-line flags for the generate command.
func ParseGenerateFlags(args []string) generator.Options {
	opts := generator.Options{
		Type: "password",
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
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
				if err == nil {
					opts.Length = length
				}
				i++
			}
		case "--bits":
			if i+1 < len(args) {
				bits, err := strconv.Atoi(args[i+1])
				if err == nil {
					opts.Bits = bits
				}
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
		}
	}

	return opts
}
