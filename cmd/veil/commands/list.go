package commands

import (
	"fmt"
	"os"
)

// ListCommand lists all secret names in a vault.
type ListCommand struct {
	BaseCommand
}

func NewListCommand() *ListCommand {
	return &ListCommand{
		BaseCommand: NewBaseCommand("list", "List all secret names in a vault"),
	}
}

func (c *ListCommand) Execute(args []string, deps Dependencies) error {
	if len(args) != 1 {
		return &UsageError{
			Command: "list",
			Usage:   "veil list <vault>",
		}
	}

	stdout := deps.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	vault := args[0]

	for name, err := range deps.App.List(vault) {
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, name)
	}

	return nil
}

func init() {
	Register(NewListCommand())
}
