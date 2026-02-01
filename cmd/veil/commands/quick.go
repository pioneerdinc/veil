package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ossydotpy/veil/cmd/veil/flags"
	"github.com/ossydotpy/veil/internal/quick"
)

// QuickCommand generates ephemeral secrets without storing them.
type QuickCommand struct {
	BaseCommand
}

func NewQuickCommand() *QuickCommand {
	return &QuickCommand{
		BaseCommand: NewBaseCommand("quick", "Generate ephemeral secret (no storage)"),
	}
}

func (c *QuickCommand) NeedsDeps() bool      { return false }
func (c *QuickCommand) NeedsMasterKey() bool { return false }

func (c *QuickCommand) Execute(args []string, deps Dependencies) error {
	stdout := deps.Stdout
	stderr := deps.Stderr
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	opts, err := flags.ParseQuickFlags(args)
	if err != nil {
		return err
	}

	if opts.ShowHelp {
		c.printHelp(stdout)
		return nil
	}

	qg := quick.New()

	// Handle batch mode
	if opts.BatchFile != "" {
		return c.runBatch(qg, opts, stdout)
	}

	// Handle count > 1
	if opts.Count > 1 {
		return c.runMultiple(qg, opts, stdout)
	}

	// Single generation
	return c.runSingle(qg, opts, stdout)
}

func (c *QuickCommand) runBatch(qg *quick.Generator, opts flags.QuickOptions, stdout io.Writer) error {
	data, err := os.ReadFile(opts.BatchFile)
	if err != nil {
		return fmt.Errorf("failed to read batch file: %w", err)
	}

	var config quick.BatchConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse batch file: %w", err)
	}

	results, err := qg.GenerateBatch(config)
	if err != nil {
		return err
	}

	// Handle file append
	if opts.ToFile != "" {
		if err := quick.AppendBatchToEnvFile(opts.ToFile, results, opts.Force); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Batch: %s\n", opts.BatchFile)
		fmt.Fprintf(stdout, "Generated %d secrets appended to %s:\n", len(results), opts.ToFile)
		for _, r := range results {
			fmt.Fprintf(stdout, "  - %s\n", r.EnvName)
		}
		return nil
	}

	// Output to terminal
	fmt.Fprintf(stdout, "Batch: %s\n", opts.BatchFile)
	fmt.Fprintf(stdout, "Generated %d secrets:\n", len(results))
	for _, r := range results {
		fmt.Fprintf(stdout, "  %s: %s\n", r.EnvName, r.Value)
	}

	return nil
}

func (c *QuickCommand) runMultiple(qg *quick.Generator, opts flags.QuickOptions, stdout io.Writer) error {
	results, err := qg.GenerateMultiple(opts.Options)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Generated %d %ss:\n", len(results), opts.Type)
	for i, result := range results {
		fmt.Fprintf(stdout, "%d. %s\n", i+1, result.Value)
	}

	return nil
}

func (c *QuickCommand) runSingle(qg *quick.Generator, opts flags.QuickOptions, stdout io.Writer) error {
	result, err := qg.Generate(opts.Options)
	if err != nil {
		return err
	}

	// Handle file append
	if opts.ToFile != "" {
		if err := quick.AppendToEnvFile(opts.ToFile, opts.EnvName, result.Value, opts.Force); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Generated: %s\n", result.Value)
		if opts.Force {
			fmt.Fprintf(stdout, "Updated %s in %s\n", opts.EnvName, opts.ToFile)
		} else {
			fmt.Fprintf(stdout, "Appended %s to %s\n", opts.EnvName, opts.ToFile)
		}
		return nil
	}

	// Display to terminal with template
	output := quick.FormatOutput(result, opts.Template)
	fmt.Fprintln(stdout, output)

	return nil
}

func (c *QuickCommand) printHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: veil quick [type] [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Generate ephemeral secrets without storing them in the vault.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Types:")
	fmt.Fprintln(w, "  password       Generate a password (default)")
	fmt.Fprintln(w, "  apikey         Generate an API key")
	fmt.Fprintln(w, "  jwt            Generate a JWT secret")
	fmt.Fprintln(w, "  uuid           Generate a UUID (v4)")
	fmt.Fprintln(w, "  uuidv7         Generate a UUID (v7)")
	fmt.Fprintln(w, "  hex            Generate a hex string")
	fmt.Fprintln(w, "  base64         Generate a base64 string")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --length N     Password length (default: 32)")
	fmt.Fprintln(w, "  --no-symbols   Alphanumeric only (no special chars)")
	fmt.Fprintln(w, "  --format       API key format: uuid, uuidv7, hex, base64")
	fmt.Fprintln(w, "  --prefix       Prefix for generated value (e.g., sk_)")
	fmt.Fprintln(w, "  --bits N       JWT secret bits: 128-512 (default: 256)")
	fmt.Fprintln(w, "  --count N      Generate multiple secrets (max: 100)")
	fmt.Fprintln(w, "  --to PATH      Append to .env file")
	fmt.Fprintln(w, "  --name KEY     Environment variable name (required with --to)")
	fmt.Fprintln(w, "  --force        Overwrite existing key in .env")
	fmt.Fprintln(w, "  --template     Custom output format")
	fmt.Fprintln(w, "  --batch FILE   Generate from JSON config file")
	fmt.Fprintln(w, "  --help, -h     Show this help message")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  veil quick                    # Generate default password")
	fmt.Fprintln(w, "  veil quick uuid               # Generate a UUID (v4)")
	fmt.Fprintln(w, "  veil quick uuidv7             # Generate a UUID (v7)")
	fmt.Fprintln(w, "  veil quick password --length 64")
	fmt.Fprintln(w, "  veil quick apikey --to .env --name API_KEY")
}

func init() {
	Register(NewQuickCommand())
}
