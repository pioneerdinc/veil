package flags

import (
	"fmt"
	"strings"

	"github.com/ossydotpy/veil/internal/exporter"
)

// ExportOptions holds parsed flags for the export command.
type ExportOptions struct {
	exporter.ExportOptions
	ShowHelp bool
}

// ParseExportFlags parses command-line flags for the export command.
func ParseExportFlags(args []string) (ExportOptions, error) {
	opts := ExportOptions{
		ExportOptions: exporter.ExportOptions{
			TargetPath: ".env",
			Format:     "env",
		},
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// All args should be flags (start with -)
		if !strings.HasPrefix(arg, "-") {
			return opts, fmt.Errorf("unexpected argument: %q (unknown flag or misplaced value)", arg)
		}

		switch arg {
		case "--to":
			if i+1 < len(args) {
				opts.TargetPath = args[i+1]
				i++
			}
		case "--force":
			opts.Force = true
		case "--append":
			opts.Append = true
		case "--dry-run":
			opts.DryRun = true
		case "--backup":
			opts.Backup = true
		case "--format":
			if i+1 < len(args) {
				opts.Format = args[i+1]
				i++
			}
		case "--include":
			if i+1 < len(args) {
				opts.Include = append(opts.Include, args[i+1])
				i++
			}
		case "--exclude":
			if i+1 < len(args) {
				opts.Exclude = append(opts.Exclude, args[i+1])
				i++
			}
		case "--help", "-h":
			opts.ShowHelp = true
		default:
			return opts, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	return opts, nil
}
