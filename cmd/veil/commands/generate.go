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

	secret, err := deps.App.Generate(vault, name, opts)
	if err != nil {
		// Check if it's a warning about existing key in .env
		if errors.Is(err, app.ErrKeyExistsInEnv) {
			printGenerateSuccess(stdout, secret, vault, name, opts.ToEnv, opts.Force)
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

func init() {
	Register(NewGenerateCommand())
}
