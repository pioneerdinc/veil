package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/ossydotpy/veil/cmd/veil/flags"
	"github.com/ossydotpy/veil/internal/importer"
)

type ImportCommand struct {
	BaseCommand
}

func NewImportCommand() *ImportCommand {
	return &ImportCommand{
		BaseCommand: NewBaseCommand("import", "Import secrets from a file into a vault"),
	}
}

func (c *ImportCommand) Execute(args []string, deps Dependencies) error {
	if len(args) < 1 {
		return &UsageError{
			Command: "import",
			Usage:   "veil import <vault> [--from <path>] [--force] [--dry-run]",
		}
	}

	stdout := deps.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	vault := args[0]
	opts, err := flags.ParseImportFlags(args[1:])
	if err != nil {
		return err
	}

	if opts.ShowHelp {
		c.printHelp(stdout)
		return nil
	}

	if opts.SourcePath == "" {
		return &UsageError{
			Command: "import",
			Usage:   "veil import <vault> --from <path>",
		}
	}

	preview, err := deps.App.Import(vault, opts.ImportOptions)
	if err != nil {
		return err
	}

	if opts.DryRun {
		printImportPreview(stdout, preview, opts.SourcePath)
	} else {
		printImportResult(stdout, preview, opts.ImportOptions, vault)
	}

	return nil
}

func printImportPreview(w io.Writer, preview *importer.Preview, sourcePath string) {
	fmt.Fprintln(w, "DRY RUN - No secrets will be imported")

	if len(preview.NewKeys) > 0 {
		fmt.Fprintf(w, "Would import to vault from %s:\n", sourcePath)
		for _, key := range preview.NewKeys {
			fmt.Fprintf(w, "  + %s\n", key)
		}
	}

	if len(preview.UpdatedKeys) > 0 {
		fmt.Fprintf(w, "\nWould update in vault:\n")
		for _, key := range preview.UpdatedKeys {
			fmt.Fprintf(w, "  ~ %s\n", key)
		}
	}

	if len(preview.SkippedKeys) > 0 {
		fmt.Fprintf(w, "\nWould skip (already exist with same value):\n")
		for _, key := range preview.SkippedKeys {
			fmt.Fprintf(w, "  - %s\n", key)
		}
	}

	fmt.Fprintf(w, "\nSummary: %s\n", preview.Summary())
}

func printImportResult(w io.Writer, preview *importer.Preview, opts importer.ImportOptions, vault string) {
	imported := len(preview.NewKeys) + len(preview.UpdatedKeys)
	skipped := len(preview.SkippedKeys)

	if imported == 0 && skipped == 0 {
		fmt.Fprintln(w, "No secrets found in source file")
	} else if skipped > 0 && imported == 0 {
		fmt.Fprintf(w, "No changes to vault '%s' (all keys already present)\n", vault)
	} else if skipped > 0 {
		fmt.Fprintf(w, "Imported %d secrets to vault '%s' (skipped %d already present)\n", imported, vault, skipped)
	} else {
		fmt.Fprintf(w, "Imported %d secrets to vault '%s'\n", imported, vault)
	}
}

func (c *ImportCommand) printHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: veil import <vault> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Import secrets from a file into a vault.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --from <path>    Source file path (required)")
	fmt.Fprintln(w, "  --format <fmt>   Input format: env (default: env)")
	fmt.Fprintln(w, "  --force          Overwrite existing vault keys")
	fmt.Fprintln(w, "  --dry-run        Preview without importing")
	fmt.Fprintln(w, "  --include        Include only matching keys (can be repeated)")
	fmt.Fprintln(w, "  --exclude        Exclude matching keys (can be repeated)")
	fmt.Fprintln(w, "  --help, -h       Show this help message")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  veil import production --from .env.production")
	fmt.Fprintln(w, "  veil import production --from .env --include 'DB_*' --include 'API_*'")
	fmt.Fprintln(w, "  veil import production --from .env --dry-run")
	fmt.Fprintln(w, "  veil import production --from .env --force")
}

func init() {
	Register(NewImportCommand())
}
