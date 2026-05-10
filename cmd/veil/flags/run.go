package flags

import (
	"fmt"
	"strings"
)

type RunOptions struct {
	Include  []string
	Exclude  []string
	ShowHelp bool
}

func ParseRunFlags(args []string) (RunOptions, error) {
	opts := RunOptions{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if !strings.HasPrefix(arg, "-") {
			return opts, fmt.Errorf("unexpected argument: %q (unknown flag or misplaced value)", arg)
		}

		switch arg {
		case "--include":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--include requires a pattern argument")
			}
			opts.Include = append(opts.Include, args[i+1])
			i++
		case "--exclude":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--exclude requires a pattern argument")
			}
			opts.Exclude = append(opts.Exclude, args[i+1])
			i++
		case "--help", "-h":
			opts.ShowHelp = true
		default:
			return opts, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	return opts, nil
}