package flags

import (
	"github.com/ossydotpy/veil/internal/exporter"
)

// ParseExportFlags parses command-line flags for the export command.
func ParseExportFlags(args []string) exporter.ExportOptions {
	opts := exporter.ExportOptions{
		TargetPath: ".env",
		Format:     "env",
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
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
		}
	}

	return opts
}
