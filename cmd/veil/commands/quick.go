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

	opts := flags.ParseQuickFlags(args)
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

func init() {
	Register(NewQuickCommand())
}
