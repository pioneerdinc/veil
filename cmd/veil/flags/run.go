package flags

import (
	"fmt"
)

// RunOptions holds parsed arguments for the run command.
type RunOptions struct {
	Vault    string
	Command  string
	Args     []string
	ShowHelp bool
}

// ParseRunFlags parses command-line flags and positional arguments for the run command.
// It looks for "--" to separate veil arguments from the application command.
func ParseRunFlags(args []string) (RunOptions, error) {
	opts := RunOptions{}

	if len(args) == 0 {
		return opts, nil
	}

	// First argument might be help flag
	if args[0] == "--help" || args[0] == "-h" {
		opts.ShowHelp = true
		return opts, nil
	}

	opts.Vault = args[0]

	// Look for the "--" separator
	commandIdx := -1
	for i := 1; i < len(args); i++ {
		if args[i] == "--" {
			commandIdx = i + 1
			break
		}
		// In the future we can parse other flags before '--' here
		if args[i] == "--help" || args[i] == "-h" {
			opts.ShowHelp = true
			return opts, nil
		}
	}

	if commandIdx != -1 && commandIdx < len(args) {
		opts.Command = args[commandIdx]
		if commandIdx+1 < len(args) {
			opts.Args = args[commandIdx+1:]
		}
	} else if commandIdx == -1 && len(args) > 1 {
		return opts, fmt.Errorf("missing '--' separator to denote the start of the command (e.g. veil run %s -- your_command)", opts.Vault)
	}

	return opts, nil
}
