package commands

import (
	"fmt"
	"os"
)

// Version is set at build time via ldflags.
var Version = "dev"

// VersionCommand displays the current version of veil.
type VersionCommand struct {
	BaseCommand
}

func NewVersionCommand() *VersionCommand {
	return &VersionCommand{
		BaseCommand: NewBaseCommand("version", "Show version information"),
	}
}

func (c *VersionCommand) NeedsDeps() bool      { return false }
func (c *VersionCommand) NeedsMasterKey() bool { return false }

func (c *VersionCommand) Execute(args []string, deps Dependencies) error {
	stdout := deps.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	fmt.Fprintf(stdout, "veil version %s\n", Version)
	return nil
}

func init() {
	Register(NewVersionCommand())
}
