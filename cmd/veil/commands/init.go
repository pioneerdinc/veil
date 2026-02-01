package commands

import (
	"fmt"
	"os"

	"github.com/ossydotpy/veil/internal/config"
	"github.com/ossydotpy/veil/internal/crypto"
)

// InitCommand generates a new master key for the vault.
type InitCommand struct {
	BaseCommand
}

func NewInitCommand() *InitCommand {
	return &InitCommand{
		BaseCommand: NewBaseCommand("init", "Generate a new master key"),
	}
}

func (c *InitCommand) NeedsDeps() bool      { return false }
func (c *InitCommand) NeedsMasterKey() bool { return false }

func (c *InitCommand) Execute(args []string, deps Dependencies) error {
	stdout := deps.Stdout
	stderr := deps.Stderr
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	cfg := config.LoadConfig()
	if _, err := os.Stat(cfg.DbPath); err == nil {
		fmt.Fprintf(stderr, "Warning: A database already exists at %s\n", cfg.DbPath)
		fmt.Fprintf(stderr, "Generating a new key and using it will make all existing secrets UNREADABLE.\n\n")
	}

	key, err := crypto.GenerateRandomKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	fmt.Fprintln(stdout, Logo)
	fmt.Fprintf(stdout, "\nYour new MASTER_KEY is:\n\n%s\n\nSAVE THIS KEY! If you lose it, your secrets are gone forever.\n", key)
	fmt.Fprintln(stdout, "Export it to your environment:\nexport MASTER_KEY="+key)

	return nil
}

func init() {
	Register(NewInitCommand())
}
