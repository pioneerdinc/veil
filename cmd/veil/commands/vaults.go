package commands

import (
	"fmt"
	"os"
)

// VaultsCommand lists all available vaults.
type VaultsCommand struct {
	BaseCommand
}

func NewVaultsCommand() *VaultsCommand {
	return &VaultsCommand{
		BaseCommand: NewBaseCommand("vaults", "List all vaults"),
	}
}

func (c *VaultsCommand) Execute(args []string, deps Dependencies) error {
	stdout := deps.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	for vault, err := range deps.App.ListVaults() {
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, vault)
	}

	return nil
}

func init() {
	Register(NewVaultsCommand())
}
