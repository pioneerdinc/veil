package commands

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ossydotpy/veil/cmd/veil/flags"
	"github.com/ossydotpy/veil/internal/app"
)

// GenerateCommand generates and stores a secret in a vault.
type GenerateCommand struct {
	BaseCommand
}

func NewGenerateCommand() *GenerateCommand {
	return &GenerateCommand{
		BaseCommand: NewBaseCommand("generate", "Generate and store a secret"),
	}
}

func (c *GenerateCommand) Execute(args []string, deps Dependencies) error {
	if len(args) < 2 {
		return &UsageError{
			Command: "generate",
			Usage:   "veil generate <vault> <name> [--length N] [--no-symbols]",
		}
	}

	stdout := deps.Stdout
	stderr := deps.Stderr
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	vault, name := args[0], args[1]
	opts, err := flags.ParseGenerateFlags(args[2:])
	if err != nil {
		return err
	}

	if opts.ShowHelp {
		c.printHelp(stdout)
		return nil
	}

	secret, err := deps.App.Generate(vault, name, opts.Options)
	if err != nil {
		// Check if it's a warning about existing key in .env
		if errors.Is(err, app.ErrKeyExistsInEnv) {
			printGenerateSuccess(stdout, secret, vault, name, opts.Options.ToEnv, opts.Options.Force)
			fmt.Fprintf(stderr, "Warning: %v\n", err)
			return nil
		}
		return err
	}

	printGenerateSuccess(stdout, secret, vault, name, opts.ToEnv, opts.Force)
	return nil
}

func printGenerateSuccess(w io.Writer, secret, vault, name, toEnv string, force bool) {
	fmt.Fprintf(w, "Generated secret: %s\n", secret)
	fmt.Fprintf(w, "Stored in %s/%s\n", vault, name)
	if toEnv != "" {
		if force {
			fmt.Fprintf(w, "Updated in %s\n", toEnv)
		} else {
			fmt.Fprintf(w, "Appended to %s\n", toEnv)
		}
	}
}

func (c *GenerateCommand) printHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: veil generate <vault> <name> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Generate and store a secret in a vault.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --type <type>    Secret type: password, apikey, jwt (default: password)")
	fmt.Fprintln(w, "  --length N       Password length: 8-128 (default: 32)")
	fmt.Fprintln(w, "  --no-symbols     Alphanumeric only (no special characters)")
	fmt.Fprintln(w, "  --format <fmt>   API key format: uuid, uuidv7, hex, base64 (default: base64)")
	fmt.Fprintln(w, "  --prefix <str>   Prefix for API key (e.g., sk_live_)")
	fmt.Fprintln(w, "  --bits N         JWT secret bits: 128-512 (default: 256)")
	fmt.Fprintln(w, "  --to-env <path>  Append generated secret to .env file")
	fmt.Fprintln(w, "  --force          Overwrite existing key in .env file")
	fmt.Fprintln(w, "  --help, -h       Show this help message")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  veil generate production DB_PASSWORD")
	fmt.Fprintln(w, "  veil generate production API_KEY --type apikey --length 48")
	fmt.Fprintln(w, "  veil generate production JWT_SECRET --type jwt --bits 512")
	fmt.Fprintln(w, "  veil generate production DB_PASSWORD --to-env .env --force")
}

func init() {
	Register(NewGenerateCommand())
}
