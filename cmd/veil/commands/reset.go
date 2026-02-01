package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ResetCommand deletes all secrets from the database.
type ResetCommand struct {
	BaseCommand
}

func NewResetCommand() *ResetCommand {
	return &ResetCommand{
		BaseCommand: NewBaseCommand("reset", "Delete all secrets and start fresh"),
	}
}

// NeedsDeps returns true because reset needs the store.
func (c *ResetCommand) NeedsDeps() bool { return true }

// NeedsMasterKey returns false because reset only needs the store, not crypto.
func (c *ResetCommand) NeedsMasterKey() bool { return false }

func (c *ResetCommand) Execute(args []string, deps Dependencies) error {
	stdout := deps.Stdout
	stderr := deps.Stderr
	stdin := deps.Stdin
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	if stdin == nil {
		stdin = os.Stdin
	}

	fmt.Fprintf(stderr, "⚠️  WARNING: This will permanently DELETE ALL SECRETS in the database.\n")
	fmt.Fprintf(stderr, "⚠️  Ensure you have backups before proceeding. This cannot be undone.\n")
	fmt.Fprintf(stderr, "Are you sure? (type 'yes' to confirm): ")

	reader := bufio.NewReader(stdin)
	confirmation, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	confirmation = strings.TrimSpace(confirmation)
	if confirmation != "yes" {
		fmt.Fprintln(stdout, "Aborted.")
		return nil
	}

	if err := deps.Store.Nuke(); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "Database wiped successfully. You can now run 'veil init' to start over.")
	return nil
}

func init() {
	Register(NewResetCommand())
}
