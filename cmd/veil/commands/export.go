package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/ossydotpy/veil/cmd/veil/flags"
	"github.com/ossydotpy/veil/internal/exporter"
)

// ExportCommand exports vault secrets to a file.
type ExportCommand struct {
	BaseCommand
}

func NewExportCommand() *ExportCommand {
	return &ExportCommand{
		BaseCommand: NewBaseCommand("export", "Export vault secrets to .env file"),
	}
}

func (c *ExportCommand) Execute(args []string, deps Dependencies) error {
	if len(args) < 1 {
		return &UsageError{
			Command: "export",
			Usage:   "veil export <vault> [--to <path>] [--force] [--append] [--dry-run]",
		}
	}

	stdout := deps.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	vault := args[0]
	opts, err := flags.ParseExportFlags(args[1:])
	if err != nil {
		return err
	}

	preview, err := deps.App.Export(vault, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		printPreview(stdout, preview, opts.TargetPath)
	} else {
		printExportResult(stdout, preview, opts, vault)
	}

	return nil
}

func printPreview(w io.Writer, preview *exporter.Preview, targetPath string) {
	fmt.Fprintln(w, "DRY RUN - No files will be modified")

	if len(preview.NewKeys) > 0 {
		fmt.Fprintf(w, "Would write to %s:\n", targetPath)
		for _, key := range preview.NewKeys {
			fmt.Fprintf(w, "  + %s\n", key)
		}
	}

	if len(preview.UpdatedKeys) > 0 {
		fmt.Fprintf(w, "\nWould update in %s:\n", targetPath)
		for _, key := range preview.UpdatedKeys {
			fmt.Fprintf(w, "  ~ %s\n", key)
		}
	}

	if len(preview.SkippedKeys) > 0 {
		fmt.Fprintf(w, "\nWould skip (already exist):\n")
		for _, key := range preview.SkippedKeys {
			fmt.Fprintf(w, "  - %s\n", key)
		}
	}

	fmt.Fprintf(w, "\nSummary: %s\n", preview.Summary())
}

func printExportResult(w io.Writer, preview *exporter.Preview, opts exporter.ExportOptions, vault string) {
	if opts.Append {
		skipped := len(preview.SkippedKeys)
		added := len(preview.NewKeys) + len(preview.UpdatedKeys)

		if added == 0 {
			fmt.Fprintf(w, "No changes to %s (all keys already present)\n", opts.TargetPath)
		} else if skipped > 0 {
			fmt.Fprintf(w, "Appended %d secrets to %s (skipped %d already present)\n", added, opts.TargetPath, skipped)
		} else {
			fmt.Fprintf(w, "Appended %d secrets to %s\n", added, opts.TargetPath)
		}
	} else {
		count := len(preview.NewKeys) + len(preview.UpdatedKeys)
		fmt.Fprintf(w, "Exported %d secrets from '%s' to %s\n", count, vault, opts.TargetPath)
	}
}

func init() {
	Register(NewExportCommand())
}
