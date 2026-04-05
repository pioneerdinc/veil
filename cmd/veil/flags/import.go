package flags

import (
	"fmt"
	"strings"

	"github.com/ossydotpy/veil/internal/importer"
)

type ImportOptions struct {
	importer.ImportOptions
	ShowHelp bool
}

func ParseImportFlags(args []string) (ImportOptions, error) {
	opts := ImportOptions{
		ImportOptions: importer.ImportOptions{
			Format: "env",
		},
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if !strings.HasPrefix(arg, "-") {
			return opts, fmt.Errorf("unexpected argument: %q (unknown flag or misplaced value)", arg)
		}

		switch arg {
		case "--from":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--from requires a path argument")
			}
			opts.SourcePath = args[i+1]
			i++
		case "--force":
			opts.Force = true
		case "--dry-run":
			opts.DryRun = true
		case "--format":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--format requires a format argument")
			}
			opts.Format = args[i+1]
			i++
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
